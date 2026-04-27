package request_marker

import (
	"context"
	"fmt"
	"github.com/qxsugar/request-marker/redis"
	"hash/fnv"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func CreateConfig() *Config {
	return &Config{}
}

type Marker struct {
	next      http.Handler
	redisConn redis.Conn
	logger    *Logger
	config    *Config
	mu        sync.RWMutex
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	logger := NewLogger(config.LogLevel)

	logger.Info(fmt.Sprintf("Initialize request marker plugin: %s", name))
	if config.RedisConfig.Enable {
		logger.Info(fmt.Sprintf("Redis config: Addr=%s, DB=%d, RuleListKeys=%s, RefreshInterval=%ds",
			config.RedisConfig.Addr, config.RedisConfig.DB, config.RedisConfig.RuleListKeys, config.RedisConfig.RefreshInterval))
	} else {
		logger.Info("Redis dynamic rule loading is disabled")
	}

	marker := &Marker{
		next:   next,
		config: config,
		logger: logger,
	}

	// Validate and sort static rules by priority (highest first)
	if config.StaticRules != nil && len(config.StaticRules) > 0 {
		for i, rule := range config.StaticRules {
			if err := rule.Validate(); err != nil {
				logger.Error(fmt.Sprintf("Invalid static rule at index %d: %v", i, err))
				return nil, fmt.Errorf("invalid rule configuration: %w", err)
			}
		}
		sort.Sort(SortByPriority(config.StaticRules))
	}

	marker.startRefreshConfig(ctx)
	return marker, nil
}

func (mk *Marker) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	mk.mu.RLock()
	rules := mk.config.StaticRules
	mk.mu.RUnlock()

	if len(rules) <= 0 {
		mk.next.ServeHTTP(w, req)
		return
	}

	for _, rule := range rules {
		if !rule.Enable {
			continue
		}

		if rule.Tag != mk.config.Tag {
			continue
		}

		var markKey, markValue string
		switch rule.Type {
		case RuleTypePath:
			if matched, err := mk.matchByURI(rule, req); matched && err == nil {
				markKey, markValue = mk.config.MarkerKey, rule.MarkerValue
			}
		case RuleTypeCanary:
			if matched, err := mk.matchByWeight(rule, req); matched && err == nil {
				markKey, markValue = mk.config.MarkerKey, rule.MarkerValue
			}
		case RuleTypeIdentify:
			if matched, err := mk.matchByIdentify(rule, req); matched && err == nil {
				markKey, markValue = mk.config.MarkerKey, rule.MarkerValue
			}
		case RuleTypeVersion:
			if matched, err := mk.matchByVersion(rule, req); matched && err == nil {
				markKey, markValue = mk.config.MarkerKey, rule.MarkerValue
			}
		}

		if markKey != "" && markValue != "" {
			req.Header.Set(markKey, markValue)
			mk.logger.Info(fmt.Sprintf("Request marked: %s=%s (rule: %s)", markKey, markValue, rule.Name))
			mk.next.ServeHTTP(w, req)
			return
		}
	}

	mk.next.ServeHTTP(w, req)
}

func (mk *Marker) startRefreshConfig(ctx context.Context) {
	if !mk.config.RedisConfig.Enable {
		mk.logger.Info("Redis dynamic rule loading is disabled, skipping refresh configuration")
		return
	}
	conn, err := NewRedis(mk.config.RedisConfig.Addr, mk.config.RedisConfig.Password, mk.config.RedisConfig.DB)
	if err != nil {
		mk.logger.Error(fmt.Sprintf("Failed to connect to Redis: %v", err))
		return
	}

	mk.redisConn = conn

	if err := mk.refreshConfig(); err != nil {
		mk.logger.Error(fmt.Sprintf("Failed to load rules on startup: %v", err))
		return
	}

	go func() {
		mk.logger.Info("Starting periodic rule refresh from Redis")
		ticker := time.NewTicker(time.Duration(mk.config.RedisConfig.RefreshInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				mk.logger.Info("Stopping rule refresh goroutine")
				return
			case <-ticker.C:
				if err := mk.refreshConfig(); err != nil {
					mk.logger.Error(fmt.Sprintf("Failed to refresh rules from Redis: %v", err))
				}
			}
		}
	}()
}

func (mk *Marker) refreshConfig() error {
	rulesListKey := mk.config.RedisConfig.RuleListKeys

	length, err := redis.Int(mk.redisConn.Do("LLEN", rulesListKey))
	if err != nil {
		return fmt.Errorf("failed to get rules list length: %w", err)
	}

	if length <= 0 {
		return fmt.Errorf("no rules found in Redis key: %s", rulesListKey)
	}

	ruleKeys, err := redis.Strings(mk.redisConn.Do("LRANGE", rulesListKey, 0, length-1))
	if err != nil {
		return fmt.Errorf("failed to fetch rule keys from Redis: %w", err)
	}

	rules := make([]Rule, 0, len(ruleKeys))

	for _, ruleKey := range ruleKeys {
		values, err := redis.Values(mk.redisConn.Do("HGETALL", ruleKey))
		if err != nil {
			mk.logger.Error(fmt.Sprintf("Failed to fetch rule from Redis (key=%s): %v", ruleKey, err))
			continue
		}

		rule, err := parseRule(values)
		if err != nil {
			mk.logger.Error(fmt.Sprintf("Failed to parse rule (key=%s): %v", ruleKey, err))
			continue
		}

		if rule.Tag != mk.config.Tag {
			continue
		}

		rules = append(rules, rule)
	}

	sort.Sort(SortByPriority(rules))

	mk.mu.Lock()
	mk.config.StaticRules = rules
	mk.mu.Unlock()

	mk.logger.Debug(fmt.Sprintf("Loaded %d rules from Redis", len(rules)))

	return nil
}

func (mk *Marker) matchByIdentify(rule Rule, req *http.Request) (bool, error) {
	identify, err := mk.extractIdentify(req)
	if err != nil {
		return false, nil
	}

	for _, userID := range rule.UserIds {
		if userID == identify {
			return true, nil
		}
	}

	return false, nil
}

func (mk *Marker) matchByURI(rule Rule, req *http.Request) (bool, error) {
	if strings.Contains(req.URL.String(), rule.Path) {
		return true, nil
	}
	return false, nil
}

func (mk *Marker) matchByVersion(rule Rule, req *http.Request) (bool, error) {
	requestVersion := req.Header.Get(mk.config.VersionHeader)
	if requestVersion == "" {
		return false, nil
	}
	if mk.compareVersion(requestVersion, rule.MinVersion) >= 0 && mk.compareVersion(rule.MaxVersion, requestVersion) >= 0 {
		return true, nil
	}

	return false, nil
}

func (mk *Marker) matchByWeight(rule Rule, req *http.Request) (bool, error) {
	hashValue, err := mk.hashIdentify(req)
	if err != nil {
		mk.logger.Debug(fmt.Sprintf("Failed to hash identify for weight matching: %v", err))
		return false, nil
	}
	if hashValue%100 <= rule.Canary {
		return true, nil
	}
	return false, nil
}

func (mk *Marker) compareVersion(version1, version2 string) int {
	v1 := strings.Split(version1, ".")
	v2 := strings.Split(version2, ".")
	for i := 0; i < len(v1) || i < len(v2); i++ {
		x, y := 0, 0
		if i < len(v1) {
			x, _ = strconv.Atoi(v1[i])
		}
		if i < len(v2) {
			y, _ = strconv.Atoi(v2[i])
		}
		if x > y {
			return 1
		}
		if x < y {
			return -1
		}
	}
	return 0
}

func (mk *Marker) hashIdentify(req *http.Request) (int, error) {
	identify, err := mk.extractIdentify(req)
	if err != nil {
		return 0, err
	}

	h := fnv.New32a()
	_, _ = h.Write([]byte(identify))
	return int(h.Sum32()), nil
}

func (mk *Marker) extractIdentify(req *http.Request) (string, error) {
	// Priority: header -> cookie -> query parameter
	identify := req.Header.Get(mk.config.IdentifyHeader)
	if identify == "" {
		cookie, err := req.Cookie(mk.config.IdentifyCookie)
		if err == nil && cookie.Value != "" {
			identify = cookie.Value
		}
	}
	if identify == "" {
		identify = req.URL.Query().Get(mk.config.IdentifyQuery)
	}

	if identify == "" {
		return "", fmt.Errorf("identify not found in header, cookie, or query parameter")
	}

	return identify, nil
}

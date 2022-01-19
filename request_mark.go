package traefik_gray_tag

import (
	"context"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

func CreateConfig() *Config {
	return &Config{}
}

type traefikGrayTag struct {
	next   http.Handler
	logger *Logger
	config *Config
	rdb    redis.Conn
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	logger := NewLogger(config.LogLevel)
	logger.Info("create new plugin, name: ", name)
	configCopy := *config
	configCopy.RedisPassword = "******"
	logger.Debug(fmt.Sprintf("config info is %+v", configCopy))

	rand.Seed(time.Now().Unix())

	plugin := &traefikGrayTag{
		next:   next,
		config: config,
		logger: logger,
	}

	// sort rules
	if config.Rules != nil && len(config.Rules) > 0 {
		sort.Sort(SortByPriority(config.Rules))
	}

	plugin.startLoadConfig()
	return plugin, nil
}

func (t *traefikGrayTag) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	log.Println("=====================")
	if len(t.config.Rules) > 0 {
		t.logger.Debug(fmt.Sprintf("rules is %+v", t.config.Rules))
		log.Println(fmt.Sprintf("rules is %+v", t.config.Rules))
		for _, rule := range t.config.Rules {
			if !rule.Enable {
				continue
			}

			if rule.ServiceName != t.config.ServiceName {
				continue
			}

			switch rule.Type {
			case ruleTypePath:
				ok, err := t.matchByPath(rule, req)
				if ok && err == nil {
					req.Header.Set(t.config.TagKey, rule.TagValue)
					t.next.ServeHTTP(rw, req)
					return
				}

			case ruleTypeIdentify:
				ok, err := t.matchByIdentify(rule, req)
				if ok && err == nil {
					req.Header.Set(t.config.TagKey, rule.TagValue)
					t.next.ServeHTTP(rw, req)
					return
				}
			case ruleTypeVersion:
				ok, err := t.matchByVersion(rule, req)
				if ok && err == nil {
					req.Header.Set(t.config.TagKey, rule.TagValue)
					t.next.ServeHTTP(rw, req)
					return
				}
			case ruleTypeWeight:
				ok, err := t.matchByWeight(rule, req)
				if ok && err == nil {
					req.Header.Set(t.config.TagKey, rule.TagValue)
					t.next.ServeHTTP(rw, req)
					return
				}
			}
		}
	}

	t.logger.Info("no match rule, will use default")
	t.next.ServeHTTP(rw, req)
}

func (t *traefikGrayTag) startLoadConfig() {
	if !t.config.RedisEnable {
		return
	}
	t.rdb = NewRedis(t.config.RedisAddr, t.config.RedisPassword)

	go func() {
		t.logger.Info("load config ticker running...")
		timeTicker := time.NewTicker(time.Duration(t.config.RedisLoadInterval) * time.Second)

		// 用不了syscall.SIGTERM，就不处理退出事件了
		for {
			select {
			case <-timeTicker.C:
				t.logger.Debug("reload config")
				err := t.reloadConfig()
				if err != nil {
					t.logger.Error("load config error", err)
				}
			}
		}
	}()
}

func (t *traefikGrayTag) reloadConfig() error {
	ruleKeys, err := redis.Strings(t.rdb.Do("LRANGE", t.config.RedisRulesKey, 0, t.config.RedisRuleMaxLen))
	if err != nil {
		return err
	}

	if len(ruleKeys) <= 0 {
		return errors.New("RuleKeys is empty")
	}

	rules := make([]Rule, 0, len(ruleKeys))

	for _, ruleKey := range ruleKeys {
		values, err := redis.Values(t.rdb.Do("HGETALL", ruleKey))
		if err != nil {
			t.logger.Error("get rule failed", "key", ruleKey, "err", err)
			continue
		}

		rule, err := parseRule(values)
		if err != nil {
			t.logger.Error("parse rule failed", "error", err)
			continue
		}
		rules = append(rules, rule)
	}

	sort.Sort(SortByPriority(rules))
	t.config.Rules = rules

	return nil
}

func (t *traefikGrayTag) matchByIdentify(rule Rule, req *http.Request) (bool, error) {
	requestId, err := t.getRequestIdentify(req)
	if err != nil {
		return false, nil
	}

	for _, userId := range rule.UserIds {
		if userId == requestId {
			return true, nil
		}
	}

	return false, nil
}

func (t *traefikGrayTag) matchByPath(rule Rule, req *http.Request) (bool, error) {
	if strings.Index(req.URL.String(), rule.Path) >= 0 {
		return true, nil
	}
	return false, nil
}

func (t *traefikGrayTag) matchByVersion(rule Rule, req *http.Request) (bool, error) {
	requestVersion := req.Header.Get(t.config.HeaderVersion)
	if t.compareVersion(requestVersion, rule.MinVersion) >= 0 && t.compareVersion(rule.MaxVersion, requestVersion) >= 0 {
		return true, nil
	}

	return false, nil
}

func (t *traefikGrayTag) matchByWeight(rule Rule, req *http.Request) (bool, error) {
	userId, err := t.identifyToNumber(req)
	if err != nil {
		t.logger.Error("identifyToNumber failed", "error", err)
		return false, nil
	}
	log.Println("user_id =================", userId, userId % 100)
	if userId%100 <= rule.Weight {
		return true, nil
	}
	return false, nil
}

func (t *traefikGrayTag) compareVersion(version1, version2 string) int {
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

func (t *traefikGrayTag) identifyToNumber(req *http.Request) (int, error) {
	identify, err := t.getRequestIdentify(req)
	if err != nil {
		return 0, err
	}

	id := 0
	for _, value := range []byte(identify) {
		id += int(value)
	}

	return id, nil
}

func (t *traefikGrayTag) getRequestIdentify(req *http.Request) (string, error) {
	// header -> cookie -> query
	identify := req.Header.Get(t.config.HeaderIdentify)
	if identify == "" {
		cookie, err := req.Cookie(t.config.CookieIdentify)
		if err == nil && cookie.Value != "" {
			identify = cookie.Value
		}
	}
	if identify == "" {
		identify = req.URL.Query().Get(t.config.QueryIdentify)
	}

	if identify == "" {
		return "", errors.New("identify is missing")
	}

	return identify, nil
}

package request_mark

import (
	"context"
	"errors"
	"fmt"
	"github.com/qxsugar/request-mark/redis"
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

type Mark struct {
	next      http.Handler
	logger    *Logger
	redisConn redis.Conn
	config    *Config
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	logger := NewLogger(config.LogLevel)

	logger.Info(fmt.Sprintf("create new request mark, name: %s", name))
	configCopy := *config
	configCopy.RedisPassword = "******"
	logger.Debug(fmt.Sprintf("config info : %+v", configCopy))

	rand.Seed(time.Now().Unix())

	mark := &Mark{
		next:   next,
		config: config,
		logger: logger,
	}

	// sort rules
	if config.Rules != nil && len(config.Rules) > 0 {
		sort.Sort(SortByPriority(config.Rules))
	}

	mark.StartRefreshConfig(ctx)
	return mark, nil
}

func (m *Mark) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if len(m.config.Rules) <= 0 {
		m.logger.Info("unmarked request. use default rule")
		m.next.ServeHTTP(rw, req)
		return
	}

	m.logger.Debug(fmt.Sprintf("begin match rule, rule is : %v", m.config.Rules))
	for _, rule := range m.config.Rules {
		if !rule.Enable {
			m.logger.Debug(fmt.Sprintf("rule is disabled, will be continue, ruleName: %v", rule.Name))
			continue
		}

		if rule.ServiceName != m.config.ServiceName {
			continue
		}

		var markKey, markValue string
		switch rule.Type {
		case ruleTypeURI:
			ok, err := m.matchByURI(rule, req)
			if ok && err == nil {
				markKey, markValue = m.config.MarkKey, rule.MarkValue
			}
		case ruleTypeWeight:
			ok, err := m.matchByWeight(rule, req)
			if ok && err == nil {
				markKey, markValue = m.config.MarkKey, rule.MarkValue
			}
		case ruleTypeIdentify:
			ok, err := m.matchByIdentify(rule, req)
			if ok && err == nil {
				markKey, markValue = m.config.MarkKey, rule.MarkValue
			}
		case ruleTypeVersion:
			ok, err := m.matchByVersion(rule, req)
			if ok && err == nil {
				markKey, markValue = m.config.MarkKey, rule.MarkValue
			}
		default:
		}

		if markKey != "" && markValue != "" {
			req.Header.Set(markKey, markValue)
			m.logger.Info(fmt.Sprintf("mark request, mark key: %v, mark value: %v", markKey, markValue))
			m.next.ServeHTTP(rw, req)
			break
		}
	}

	m.logger.Info("unmarked request. use default rule")
	m.next.ServeHTTP(rw, req)
}

func (m *Mark) StartRefreshConfig(ctx context.Context) {
	if !m.config.RedisEnable {
		return
	}
	m.redisConn = NewRedisOrDie(m.config.RedisAddr, m.config.RedisPassword)

	go func() {
		m.logger.Info("refresh config ticker running...")
		timeTicker := time.NewTicker(time.Duration(m.config.RedisLoadInterval) * time.Second)

		// 用不了syscall.SIGTERM，就不处理退出事件了
		for {
			select {
			case <-ctx.Done():
				timeTicker.Stop()
				return
			case <-timeTicker.C:
				m.logger.Debug("refresh config")
				err := m.RefreshConfig()
				if err != nil {
					m.logger.Error(fmt.Sprintf("failed to refresh config, err: %v", err))
				}
			}
		}
	}()
}

func (m *Mark) RefreshConfig() error {
	ruleKeys, err := redis.Strings(m.redisConn.Do("LRANGE", m.config.RedisRulesKey, 0, m.config.RedisRuleMaxLen))
	if err != nil {
		return err
	}

	if len(ruleKeys) <= 0 {
		return errors.New("failed to refresh config, rule keys is empty")
	}

	rules := make([]Rule, 0, len(ruleKeys))

	for _, ruleKey := range ruleKeys {
		values, err := redis.Values(m.redisConn.Do("HGETALL", ruleKey))
		if err != nil {
			m.logger.Error(fmt.Sprintf("failed to get rule, key: %s, error: %v", ruleKey, err))
			continue
		}

		rule, err := parseRule(values)
		if err != nil {
			m.logger.Error(fmt.Sprintf("failed to parse rule, error, %v", err))
			continue
		}
		rules = append(rules, rule)
	}

	sort.Sort(SortByPriority(rules))
	m.config.Rules = rules

	return nil
}

func (m *Mark) matchByIdentify(rule Rule, req *http.Request) (bool, error) {
	requestIdentify, err := m.getRequestIdentify(req)
	if err != nil {
		return false, nil
	}

	for _, userId := range rule.UserIds {
		if userId == requestIdentify {
			return true, nil
		}
	}

	return false, nil
}

func (m *Mark) matchByURI(rule Rule, req *http.Request) (bool, error) {
	if strings.Index(req.URL.String(), rule.Path) >= 0 {
		return true, nil
	}
	return false, nil
}

func (m *Mark) matchByVersion(rule Rule, req *http.Request) (bool, error) {
	requestVersion := req.Header.Get(m.config.HeaderVersion)
	if m.compareVersion(requestVersion, rule.MinVersion) >= 0 && m.compareVersion(rule.MaxVersion, requestVersion) >= 0 {
		return true, nil
	}

	return false, nil
}

func (m *Mark) matchByWeight(rule Rule, req *http.Request) (bool, error) {
	userId, err := m.getNumberIdentify(req)
	if err != nil {
		m.logger.Error("getNumberIdentify failed", "error", err)
		return false, nil
	}
	if userId%100 <= rule.Weight {
		return true, nil
	}
	return false, nil
}

func (m *Mark) compareVersion(version1, version2 string) int {
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

func (m *Mark) getNumberIdentify(req *http.Request) (int, error) {
	identify, err := m.getRequestIdentify(req)
	if err != nil {
		return 0, err
	}

	id := 0
	for _, value := range []byte(identify) {
		id += int(value)
	}

	return id, nil
}

func (m *Mark) getRequestIdentify(req *http.Request) (string, error) {
	// header -> cookie -> query
	identify := req.Header.Get(m.config.HeaderIdentify)
	if identify == "" {
		cookie, err := req.Cookie(m.config.CookieIdentify)
		if err == nil && cookie.Value != "" {
			identify = cookie.Value
		}
	}
	if identify == "" {
		identify = req.URL.Query().Get(m.config.QueryIdentify)
	}

	if identify == "" {
		return "", errors.New("get identify failed, identify is missing")
	}

	return identify, nil
}

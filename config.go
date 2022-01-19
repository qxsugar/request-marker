package request_mark

import (
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"strings"
)

type RuleType string

const (
	ruleTypeVersion  = RuleType("version")
	ruleTypeIdentify = RuleType("identify")
	ruleTypeWeight   = RuleType("weight")
	ruleTypePath     = RuleType("path")

	keyServiceName = "service_name"
	keyName        = "name"
	keyEnable      = "enable"
	keyPriority    = "priority"
	keyType        = "type"
	keyTagValue    = "tag_value"
	keyVersion     = "version"
	keyUserIds     = "user_ids"
	keyWeight      = "weight"
	keyPath        = "path"
)

type Config struct {
	ServiceName       string `json:"serviceName"`       // serviceName
	LogLevel          string `json:"logLevel"`          // logger 等级
	RedisAddr         string `json:"redisAddr"`         // redis 地址
	RedisPassword     string `json:"redisPassword"`     // redis 密码
	RedisEnable       bool   `json:"redisEnable"`       // 是否启用redis配置
	RedisRulesKey     string `json:"redisRulesKey"`     // redis规则key
	RedisRuleMaxLen   int    `json:"redisRuleMaxLen"`   // redis规则列表的长度
	RedisLoadInterval int64  `json:"redisLoadInterval"` // redis加载间隔
	Rules             []Rule `json:"rules"`             // 规则列表
	MarkKey           string `json:"MarkKey"`           // 标记header的key
	HeaderVersion     string `json:"headerVersion"`     // 请求头里的version key
	HeaderIdentify    string `json:"headerIdentify"`    // header里的identify key
	CookieIdentify    string `json:"cookieIdentify"`    // cookie里的identify key
	QueryIdentify     string `json:"query_identify"`    // query里的identify key
}

type Rule struct {
	ServiceName string   `json:"serviceName"` // serviceName
	Name        string   `json:"name"`        // 规则名字
	Enable      bool     `json:"enable"`      // 是否开启
	Priority    int      `json:"priority"`    // 优先级
	Type        RuleType `json:"type"`        // 规则类型
	MarkValue   string   `json:"tagValue"`    // mark值
	MaxVersion  string   `json:"maxVersion"`  // 最大版本号
	MinVersion  string   `json:"minVersion"`  // 最小版本号
	UserIds     []string `json:"userIds"`     // 用户identify列表
	Weight      int      `json:"weight"`      // 权重
	Path        string   `json:"path"`        // 路径匹配
}

// 解析redis values到rule
// yaegi解释器对很多库支持不好。所以手动解析
// format ["serviceName", "foo", "enable", 1]
func parseRule(values []interface{}) (Rule, error) {
	r := Rule{}
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].([]byte)
		if !ok {
			return r, fmt.Errorf("expects type for String, got type %T", values[i])
		}
		value := values[i+1]

		switch string(key) {
		case keyServiceName:
			serviceName, err := redis.String(value, nil)
			if err != nil {
				return r, err
			}
			r.ServiceName = serviceName
		case keyName:
			name, err := redis.String(value, nil)
			if err != nil {
				return r, err
			}
			r.Name = name
		case keyEnable:
			enable, err := redis.Bool(value, nil)
			if err != nil {
				return r, err
			}
			r.Enable = enable
		case keyPriority:
			priority, err := redis.Int(value, nil)
			if err != nil {
				return r, err
			}
			r.Priority = priority
		case keyType:
			ruleType, err := redis.String(value, nil)
			if err != nil {
				return r, err
			}
			r.Type = RuleType(ruleType)
		case keyTagValue:
			markValue, err := redis.String(value, nil)
			if err != nil {
				return r, err
			}
			r.MarkValue = markValue
		case keyVersion:
			version, err := redis.String(value, nil)
			if err != nil {
				return r, err
			}
			versionArr := strings.Split(version, "-")
			if len(versionArr) != 2 {
				return r, errors.New(fmt.Sprintf("version error, version [%s] invalid", version))
			}

			r.MinVersion = versionArr[0]
			r.MaxVersion = versionArr[1]
		case keyUserIds:
			list, err := redis.String(value, nil)
			if err != nil {
				return r, err
			}
			r.UserIds = strings.Split(list, ",")
		case keyWeight:
			weight, err := redis.Int(value, nil)
			if err != nil {
				return r, err
			}
			r.Weight = weight
		case keyPath:
			path, err := redis.String(value, nil)
			if err != nil {
				return r, err
			}
			r.Path = path
		}
	}
	return r, nil
}

type SortByPriority []Rule

func (a SortByPriority) Len() int           { return len(a) }
func (a SortByPriority) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortByPriority) Less(i, j int) bool { return a[i].Priority > a[j].Priority }

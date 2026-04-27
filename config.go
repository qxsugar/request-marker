package request_marker

import (
	"fmt"
	"github.com/qxsugar/request-marker/redis"
	"strings"
)

type RuleType string

const (
	RuleTypePath     = RuleType("path")
	RuleTypeVersion  = RuleType("version")
	RuleTypeIdentify = RuleType("identify")
	RuleTypeCanary   = RuleType("canary")
)

const (
	FieldName       = "name"
	FieldEnable     = "enable"
	FieldPriority   = "priority"
	FieldType       = "type"
	FieldMarkValue  = "mark_value"
	FieldMinVersion = "min_version"
	FieldMaxVersion = "max_version"
	FieldUserIds    = "user_ids"
	FieldWeight     = "weight"
	FieldPath       = "path"
)

type Rule struct {
	Tag         string   `json:"tag"`        // tag，当rule.tag和config.tag匹配时候，才会使用这个规则
	Name        string   `json:"name"`       // 规则名字
	Enable      bool     `json:"enable"`     // 是否开启
	Priority    int      `json:"priority"`   // 优先级，越高越优先
	Type        RuleType `json:"type"`       // 规则类型
	MarkerValue string   `json:"markValue"`  // 标记值
	MaxVersion  string   `json:"maxVersion"` // RuleTypeVersion: 最大版本
	MinVersion  string   `json:"minVersion"` // RuleTypeVersion: 最小版本
	UserIds     []string `json:"userIds"`    // RuleTypeIdentify: 适用的用户ID列表
	Canary      int      `json:"Canary"`     // RuleTypeCanary: 流量百分比（0-100）
	Path        string   `json:"path"`       // RuleTypePath: URI路径匹配规则
}

type RedisConfig struct {
	Enable          bool   `json:"enable"`          // 是否开启
	Addr            string `json:"addr"`            // redis地址
	Password        string `json:"password"`        // redis密码
	DB              int    `json:"db"`              // redis数据库
	RuleListKeys    string `json:"ruleListKeys"`    // 规则列表Key
	RefreshInterval int64  `json:"refreshInterval"` // 刷新间隔，单位秒
}

type Config struct {
	Tag            string      `json:"tag"`            // tag，当rule.tag和config.tag匹配时候，才会使用这个规则
	LogLevel       string      `json:"log_level"`      // 日志登记
	RedisConfig    RedisConfig `json:"redis_config"`   // redis 配置，如果配置了。则使用动态配置
	StaticRules    []Rule      `json:"static_rules"`   // 静态路由配置
	MarkerKey      string      `json:"marker_key"`     // 标记 key
	VersionHeader  string      `json:"versionHeader"`  // 版本号的header
	IdentifyHeader string      `json:"identifyHeader"` // 用户身份的header
	IdentifyCookie string      `json:"identifyCookie"` // 用户身份的cookie
	IdentifyQuery  string      `json:"identifyQuery"`  // 用户身份的query参数
}

func (r *Rule) Validate() error {
	//if r.Name == "" {
	//	return fmt.Errorf("rule name cannot be empty")
	//}
	//if r.MarkValue == "" {
	//	return fmt.Errorf("rule mark_value cannot be empty")
	//}
	//switch r.Type {
	//case ruleTypeVersion:
	//	if r.MinVersion == "" || r.MaxVersion == "" {
	//		return fmt.Errorf("version rule requires both minVersion and maxVersion")
	//	}
	//case ruleTypeIdentify:
	//	if len(r.UserIds) == 0 {
	//		return fmt.Errorf("identify rule requires at least one userIds")
	//	}
	//case ruleTypeWeight:
	//	if r.Weight < 0 || r.Weight > 100 {
	//		return fmt.Errorf("weight must be between 0 and 100, got %d", r.Weight)
	//	}
	//case ruleTypeURI:
	//	if r.Path == "" {
	//		return fmt.Errorf("uri rule requires path")
	//	}
	//default:
	//	return fmt.Errorf("unknown rule type: %s", r.Type)
	//}
	return nil
}

func parseRule(values []interface{}) (Rule, error) {
	rule := Rule{}
	for i := 0; i < len(values); i += 2 {
		fieldName, ok := values[i].([]byte)
		if !ok {
			return rule, fmt.Errorf("expected field name as string, got %T", values[i])
		}
		fieldValue := values[i+1]

		switch string(fieldName) {
		case FieldName:
			val, err := redis.String(fieldValue, nil)
			if err != nil {
				return rule, err
			}
			rule.Name = val
		case FieldEnable:
			val, err := redis.Bool(fieldValue, nil)
			if err != nil {
				return rule, err
			}
			rule.Enable = val
		case FieldPriority:
			val, err := redis.Int(fieldValue, nil)
			if err != nil {
				return rule, err
			}
			rule.Priority = val
		case FieldType:
			val, err := redis.String(fieldValue, nil)
			if err != nil {
				return rule, err
			}
			rule.Type = RuleType(val)
		case FieldMarkValue:
			val, err := redis.String(fieldValue, nil)
			if err != nil {
				return rule, err
			}
			rule.MarkerValue = val
		case FieldMinVersion:
			val, err := redis.String(fieldValue, nil)
			if err != nil {
				return rule, err
			}
			rule.MinVersion = val
		case FieldMaxVersion:
			val, err := redis.String(fieldValue, nil)
			if err != nil {
				return rule, err
			}
			rule.MaxVersion = val
		case FieldUserIds:
			val, err := redis.String(fieldValue, nil)
			if err != nil {
				return rule, err
			}
			rule.UserIds = strings.Split(val, ",")
		case FieldWeight:
			val, err := redis.Int(fieldValue, nil)
			if err != nil {
				return rule, err
			}
			rule.Canary = val
		case FieldPath:
			val, err := redis.String(fieldValue, nil)
			if err != nil {
				return rule, err
			}
			rule.Path = val
		}
	}

	// Validate rule after parsing
	if err := rule.Validate(); err != nil {
		return rule, fmt.Errorf("invalid rule: %w", err)
	}

	return rule, nil
}

type SortByPriority []Rule

func (a SortByPriority) Len() int           { return len(a) }
func (a SortByPriority) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortByPriority) Less(i, j int) bool { return a[i].Priority > a[j].Priority }

package request_marker

import (
	"sort"
	"testing"
)

func TestParseRule_AllFields(t *testing.T) {
	values := []interface{}{
		[]byte("name"), []byte("test-rule"),
		[]byte("enable"), []byte("1"),
		[]byte("priority"), []byte("100"),
		[]byte("type"), []byte("identify"),
		[]byte("mark_value"), []byte("beta"),
		[]byte("min_version"), []byte("1.0.0"),
		[]byte("max_version"), []byte("2.0.0"),
		[]byte("user_ids"), []byte("user1,user2,user3"),
		[]byte("weight"), []byte("30"),
		[]byte("path"), []byte("/admin"),
	}

	rule, err := parseRule(values)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rule.Name != "test-rule" {
		t.Errorf("expected name=test-rule, got %s", rule.Name)
	}
	if !rule.Enable {
		t.Errorf("expected enable=true, got %v", rule.Enable)
	}
	if rule.Priority != 100 {
		t.Errorf("expected priority=100, got %d", rule.Priority)
	}
	if rule.Type != RuleTypeIdentify {
		t.Errorf("expected type=identify, got %s", rule.Type)
	}
	if rule.MarkerValue != "beta" {
		t.Errorf("expected markValue=beta, got %s", rule.MarkerValue)
	}
	if rule.MinVersion != "1.0.0" {
		t.Errorf("expected minVersion=1.0.0, got %s", rule.MinVersion)
	}
	if rule.MaxVersion != "2.0.0" {
		t.Errorf("expected maxVersion=2.0.0, got %s", rule.MaxVersion)
	}
	if len(rule.UserIds) != 3 || rule.UserIds[0] != "user1" {
		t.Errorf("expected userIds=[user1,user2,user3], got %v", rule.UserIds)
	}
	if rule.Canary != 30 {
		t.Errorf("expected canary=30, got %d", rule.Canary)
	}
	if rule.Path != "/admin" {
		t.Errorf("expected path=/admin, got %s", rule.Path)
	}
}

func TestParseRule_MinimalFields(t *testing.T) {
	values := []interface{}{
		[]byte("name"), []byte("minimal-rule"),
		[]byte("enable"), []byte("0"),
		[]byte("priority"), []byte("50"),
		[]byte("type"), []byte("path"),
		[]byte("mark_value"), []byte("mark"),
	}

	rule, err := parseRule(values)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rule.Name != "minimal-rule" {
		t.Errorf("expected name=minimal-rule, got %s", rule.Name)
	}
	if rule.Enable {
		t.Errorf("expected enable=false, got %v", rule.Enable)
	}
	if rule.Priority != 50 {
		t.Errorf("expected priority=50, got %d", rule.Priority)
	}
}

func TestParseRule_InvalidFieldName(t *testing.T) {
	values := []interface{}{
		"not-bytes", []byte("value"),
	}

	_, err := parseRule(values)
	if err == nil {
		t.Errorf("expected error for invalid field name type")
	}
}

func TestParseRule_UserIdsParsing(t *testing.T) {
	values := []interface{}{
		[]byte("name"), []byte("test"),
		[]byte("enable"), []byte("1"),
		[]byte("priority"), []byte("1"),
		[]byte("type"), []byte("identify"),
		[]byte("mark_value"), []byte("test"),
		[]byte("user_ids"), []byte("alice,bob,charlie"),
	}

	rule, err := parseRule(values)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rule.UserIds) != 3 {
		t.Errorf("expected 3 user IDs, got %d", len(rule.UserIds))
	}
	if rule.UserIds[0] != "alice" || rule.UserIds[1] != "bob" || rule.UserIds[2] != "charlie" {
		t.Errorf("expected [alice,bob,charlie], got %v", rule.UserIds)
	}
}

func TestSortByPriority(t *testing.T) {
	rules := []Rule{
		{Name: "low", Priority: 10},
		{Name: "high", Priority: 100},
		{Name: "medium", Priority: 50},
	}

	sorted := make([]Rule, len(rules))
	copy(sorted, rules)

	sort.Sort(SortByPriority(sorted))

	if sorted[0].Priority != 100 {
		t.Errorf("expected first rule priority=100, got %d", sorted[0].Priority)
	}
	if sorted[1].Priority != 50 {
		t.Errorf("expected second rule priority=50, got %d", sorted[1].Priority)
	}
	if sorted[2].Priority != 10 {
		t.Errorf("expected third rule priority=10, got %d", sorted[2].Priority)
	}
}

func TestRuleValidate(t *testing.T) {
	rule := Rule{
		Name:        "test",
		Enable:      true,
		Priority:    100,
		Type:        RuleTypeVersion,
		MarkerValue: "mark",
		MinVersion:  "1.0.0",
		MaxVersion:  "2.0.0",
	}

	err := rule.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRuleTypes(t *testing.T) {
	tests := []struct {
		ruleType RuleType
		expected string
	}{
		{RuleTypePath, "path"},
		{RuleTypeVersion, "version"},
		{RuleTypeIdentify, "identify"},
		{RuleTypeCanary, "canary"},
	}

	for _, tt := range tests {
		if string(tt.ruleType) != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, string(tt.ruleType))
		}
	}
}

func TestRedisConfigDefaults(t *testing.T) {
	config := RedisConfig{
		Enable:          true,
		Addr:            "localhost:6379",
		DB:              0,
		RuleListKeys:    "marker:api:rules",
		RefreshInterval: 15,
	}

	if config.Addr != "localhost:6379" {
		t.Errorf("expected addr=localhost:6379, got %s", config.Addr)
	}
	if config.DB != 0 {
		t.Errorf("expected db=0, got %d", config.DB)
	}
	if config.RefreshInterval != 15 {
		t.Errorf("expected refreshInterval=15, got %d", config.RefreshInterval)
	}
}

func TestConfigStructure(t *testing.T) {
	config := Config{
		Tag:            "api",
		LogLevel:       "DEBUG",
		MarkerKey:      "X-MARK",
		VersionHeader:  "X-Version",
		IdentifyHeader: "X-User-ID",
		IdentifyCookie: "user_id",
		IdentifyQuery:  "uid",
		RedisConfig: RedisConfig{
			Enable:       true,
			Addr:         "localhost:6379",
			RuleListKeys: "marker:api:rules",
		},
		StaticRules: []Rule{
			{
				Tag:         "api",
				Name:        "test",
				Enable:      true,
				Priority:    100,
				Type:        RuleTypePath,
				MarkerValue: "mark",
				Path:        "/",
			},
		},
	}

	if config.Tag != "api" {
		t.Errorf("expected tag=api, got %s", config.Tag)
	}
	if len(config.StaticRules) != 1 {
		t.Errorf("expected 1 static rule, got %d", len(config.StaticRules))
	}
	if !config.RedisConfig.Enable {
		t.Errorf("expected redis enabled")
	}
}

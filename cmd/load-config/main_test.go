package main

import (
	"testing"
)

func TestBoolToInt_True(t *testing.T) {
	result := boolToInt(true)
	if result != 1 {
		t.Errorf("expected 1, got %d", result)
	}
}

func TestBoolToInt_False(t *testing.T) {
	result := boolToInt(false)
	if result != 0 {
		t.Errorf("expected 0, got %d", result)
	}
}

func TestRuleStructure(t *testing.T) {
	rule := Rule{
		Tag:         "api",
		Name:        "test-rule",
		Enable:      true,
		Priority:    100,
		Type:        "identify",
		MarkerValue: "beta",
		MinVersion:  "1.0.0",
		MaxVersion:  "2.0.0",
		UserIds:     []string{"user1", "user2"},
		Canary:      30,
		Path:        "/admin",
	}

	if rule.Tag != "api" {
		t.Errorf("expected tag=api, got %s", rule.Tag)
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
	if rule.Type != "identify" {
		t.Errorf("expected type=identify, got %s", rule.Type)
	}
	if rule.MarkerValue != "beta" {
		t.Errorf("expected markValue=beta, got %s", rule.MarkerValue)
	}
	if len(rule.UserIds) != 2 {
		t.Errorf("expected 2 user IDs, got %d", len(rule.UserIds))
	}
	if rule.Canary != 30 {
		t.Errorf("expected canary=30, got %d", rule.Canary)
	}
	if rule.Path != "/admin" {
		t.Errorf("expected path=/admin, got %s", rule.Path)
	}
}

func TestRedisConfigStructure(t *testing.T) {
	config := RedisConfig{
		Enable:          true,
		Addr:            "localhost:6379",
		Password:        "secret",
		DB:              1,
		RuleListKeys:    "marker:api:rules",
		RefreshInterval: 15,
	}

	if !config.Enable {
		t.Errorf("expected enable=true, got %v", config.Enable)
	}
	if config.Addr != "localhost:6379" {
		t.Errorf("expected addr=localhost:6379, got %s", config.Addr)
	}
	if config.Password != "secret" {
		t.Errorf("expected password=secret, got %s", config.Password)
	}
	if config.DB != 1 {
		t.Errorf("expected db=1, got %d", config.DB)
	}
	if config.RuleListKeys != "marker:api:rules" {
		t.Errorf("expected ruleListKeys=marker:api:rules, got %s", config.RuleListKeys)
	}
	if config.RefreshInterval != 15 {
		t.Errorf("expected refreshInterval=15, got %d", config.RefreshInterval)
	}
}

func TestConfigStructure(t *testing.T) {
	config := Config{
		Tag:         "api",
		LogLevel:    "DEBUG",
		MarkerKey:   "X-MARK",
		VersionHeader: "X-Version",
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
				Type:        "path",
				MarkerValue: "mark",
				Path:        "/",
			},
		},
	}

	if config.Tag != "api" {
		t.Errorf("expected tag=api, got %s", config.Tag)
	}
	if config.LogLevel != "DEBUG" {
		t.Errorf("expected logLevel=DEBUG, got %s", config.LogLevel)
	}
	if config.MarkerKey != "X-MARK" {
		t.Errorf("expected markerKey=X-MARK, got %s", config.MarkerKey)
	}
	if len(config.StaticRules) != 1 {
		t.Errorf("expected 1 static rule, got %d", len(config.StaticRules))
	}
}

func TestRuleTypes(t *testing.T) {
	types := []string{"identify", "version", "canary", "path"}
	for _, ruleType := range types {
		if ruleType == "" {
			t.Errorf("rule type should not be empty")
		}
	}
}

func TestUserIdsParsing(t *testing.T) {
	rule := Rule{
		UserIds: []string{"user1", "user2", "user3"},
	}

	if len(rule.UserIds) != 3 {
		t.Errorf("expected 3 user IDs, got %d", len(rule.UserIds))
	}
	if rule.UserIds[0] != "user1" {
		t.Errorf("expected first user ID=user1, got %s", rule.UserIds[0])
	}
	if rule.UserIds[1] != "user2" {
		t.Errorf("expected second user ID=user2, got %s", rule.UserIds[1])
	}
	if rule.UserIds[2] != "user3" {
		t.Errorf("expected third user ID=user3, got %s", rule.UserIds[2])
	}
}

func TestVersionFields(t *testing.T) {
	rule := Rule{
		MinVersion: "1.0.0",
		MaxVersion: "2.0.0",
	}

	if rule.MinVersion != "1.0.0" {
		t.Errorf("expected minVersion=1.0.0, got %s", rule.MinVersion)
	}
	if rule.MaxVersion != "2.0.0" {
		t.Errorf("expected maxVersion=2.0.0, got %s", rule.MaxVersion)
	}
}

func TestCanaryField(t *testing.T) {
	rule := Rule{
		Canary: 50,
	}

	if rule.Canary != 50 {
		t.Errorf("expected canary=50, got %d", rule.Canary)
	}

	if rule.Canary < 0 || rule.Canary > 100 {
		t.Errorf("canary should be between 0 and 100, got %d", rule.Canary)
	}
}

func TestPathField(t *testing.T) {
	rule := Rule{
		Path: "/admin/users",
	}

	if rule.Path != "/admin/users" {
		t.Errorf("expected path=/admin/users, got %s", rule.Path)
	}
}

func TestMultipleRules(t *testing.T) {
	rules := []Rule{
		{Name: "rule1", Priority: 100},
		{Name: "rule2", Priority: 50},
		{Name: "rule3", Priority: 75},
	}

	if len(rules) != 3 {
		t.Errorf("expected 3 rules, got %d", len(rules))
	}

	for i, rule := range rules {
		if rule.Name == "" {
			t.Errorf("rule %d should have a name", i)
		}
		if rule.Priority <= 0 {
			t.Errorf("rule %d should have positive priority", i)
		}
	}
}

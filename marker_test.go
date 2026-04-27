package request_marker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
)

func TestMarkerServeHTTP_IdentifyRule(t *testing.T) {
	config := &Config{
		Tag:            "api",
		LogLevel:       "DEBUG",
		MarkerKey:      "X-MARK",
		VersionHeader:  "X-Version",
		IdentifyHeader: "X-User-ID",
		StaticRules: []Rule{
			{
				Tag:         "api",
				Name:        "beta-users",
				Enable:      true,
				Priority:    100,
				Type:        RuleTypeIdentify,
				MarkerValue: "beta",
				UserIds:     []string{"user001", "user002"},
			},
		},
	}

	marker := &Marker{
		next:   http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		config: config,
		logger: NewLogger("DEBUG"),
	}

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-User-ID", "user001")
	w := httptest.NewRecorder()

	marker.ServeHTTP(w, req)

	if req.Header.Get("X-MARK") != "beta" {
		t.Errorf("expected X-MARK=beta, got %s", req.Header.Get("X-MARK"))
	}
}

func TestMarkerServeHTTP_VersionRule(t *testing.T) {
	config := &Config{
		Tag:           "api",
		LogLevel:      "DEBUG",
		MarkerKey:     "X-MARK",
		VersionHeader: "X-Version",
		StaticRules: []Rule{
			{
				Tag:         "api",
				Name:        "v2-range",
				Enable:      true,
				Priority:    100,
				Type:        RuleTypeVersion,
				MarkerValue: "v2",
				MinVersion:  "2.0.0",
				MaxVersion:  "2.9.9",
			},
		},
	}

	marker := &Marker{
		next:   http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		config: config,
		logger: NewLogger("DEBUG"),
	}

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-Version", "2.5.0")
	w := httptest.NewRecorder()

	marker.ServeHTTP(w, req)

	if req.Header.Get("X-MARK") != "v2" {
		t.Errorf("expected X-MARK=v2, got %s", req.Header.Get("X-MARK"))
	}
}

func TestMarkerServeHTTP_CanaryRule(t *testing.T) {
	config := &Config{
		Tag:            "api",
		LogLevel:       "DEBUG",
		MarkerKey:      "X-MARK",
		IdentifyHeader: "X-User-ID",
		StaticRules: []Rule{
			{
				Tag:         "api",
				Name:        "canary-30",
				Enable:      true,
				Priority:    100,
				Type:        RuleTypeCanary,
				MarkerValue: "canary",
				Canary:      30,
			},
		},
	}

	marker := &Marker{
		next:   http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		config: config,
		logger: NewLogger("DEBUG"),
	}

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-User-ID", "user123")
	w := httptest.NewRecorder()

	marker.ServeHTTP(w, req)

	// Canary is probabilistic, just check it doesn't error
	_ = req.Header.Get("X-MARK")
}

func TestMarkerServeHTTP_PathRule(t *testing.T) {
	config := &Config{
		Tag:       "api",
		LogLevel:  "DEBUG",
		MarkerKey: "X-MARK",
		StaticRules: []Rule{
			{
				Tag:         "api",
				Name:        "admin-path",
				Enable:      true,
				Priority:    100,
				Type:        RuleTypePath,
				MarkerValue: "admin",
				Path:        "/admin",
			},
		},
	}

	marker := &Marker{
		next:   http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		config: config,
		logger: NewLogger("DEBUG"),
	}

	req := httptest.NewRequest("GET", "/admin/users", nil)
	w := httptest.NewRecorder()

	marker.ServeHTTP(w, req)

	if req.Header.Get("X-MARK") != "admin" {
		t.Errorf("expected X-MARK=admin, got %s", req.Header.Get("X-MARK"))
	}
}

func TestMarkerServeHTTP_TagMismatch(t *testing.T) {
	config := &Config{
		Tag:       "api",
		LogLevel:  "DEBUG",
		MarkerKey: "X-MARK",
		StaticRules: []Rule{
			{
				Tag:         "other",
				Name:        "other-rule",
				Enable:      true,
				Priority:    100,
				Type:        RuleTypePath,
				MarkerValue: "other",
				Path:        "/admin",
			},
		},
	}

	marker := &Marker{
		next:   http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		config: config,
		logger: NewLogger("DEBUG"),
	}

	req := httptest.NewRequest("GET", "/admin/users", nil)
	w := httptest.NewRecorder()

	marker.ServeHTTP(w, req)

	if req.Header.Get("X-MARK") != "" {
		t.Errorf("expected no mark, got %s", req.Header.Get("X-MARK"))
	}
}

func TestMarkerServeHTTP_DisabledRule(t *testing.T) {
	config := &Config{
		Tag:       "api",
		LogLevel:  "DEBUG",
		MarkerKey: "X-MARK",
		StaticRules: []Rule{
			{
				Tag:         "api",
				Name:        "disabled-rule",
				Enable:      false,
				Priority:    100,
				Type:        RuleTypePath,
				MarkerValue: "disabled",
				Path:        "/admin",
			},
		},
	}

	marker := &Marker{
		next:   http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		config: config,
		logger: NewLogger("DEBUG"),
	}

	req := httptest.NewRequest("GET", "/admin/users", nil)
	w := httptest.NewRecorder()

	marker.ServeHTTP(w, req)

	if req.Header.Get("X-MARK") != "" {
		t.Errorf("expected no mark, got %s", req.Header.Get("X-MARK"))
	}
}

func TestMarkerServeHTTP_PriorityOrder(t *testing.T) {
	config := &Config{
		Tag:       "api",
		LogLevel:  "DEBUG",
		MarkerKey: "X-MARK",
		StaticRules: []Rule{
			{
				Tag:         "api",
				Name:        "low-priority",
				Enable:      true,
				Priority:    10,
				Type:        RuleTypePath,
				MarkerValue: "low",
				Path:        "/",
			},
			{
				Tag:         "api",
				Name:        "high-priority",
				Enable:      true,
				Priority:    100,
				Type:        RuleTypePath,
				MarkerValue: "high",
				Path:        "/",
			},
		},
	}

	// Sort rules by priority (highest first)
	sort.Sort(SortByPriority(config.StaticRules))

	marker := &Marker{
		next:   http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		config: config,
		logger: NewLogger("DEBUG"),
	}

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	marker.ServeHTTP(w, req)

	if req.Header.Get("X-MARK") != "high" {
		t.Errorf("expected X-MARK=high (higher priority), got %s", req.Header.Get("X-MARK"))
	}
}

func TestExtractIdentify_Header(t *testing.T) {
	config := &Config{
		IdentifyHeader: "X-User-ID",
		IdentifyCookie: "user_id",
		IdentifyQuery:  "uid",
	}

	marker := &Marker{
		config: config,
		logger: NewLogger("DEBUG"),
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-User-ID", "user123")

	identify, err := marker.extractIdentify(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if identify != "user123" {
		t.Errorf("expected user123, got %s", identify)
	}
}

func TestExtractIdentify_Cookie(t *testing.T) {
	config := &Config{
		IdentifyHeader: "X-User-ID",
		IdentifyCookie: "user_id",
		IdentifyQuery:  "uid",
	}

	marker := &Marker{
		config: config,
		logger: NewLogger("DEBUG"),
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "user_id", Value: "user456"})

	identify, err := marker.extractIdentify(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if identify != "user456" {
		t.Errorf("expected user456, got %s", identify)
	}
}

func TestExtractIdentify_Query(t *testing.T) {
	config := &Config{
		IdentifyHeader: "X-User-ID",
		IdentifyCookie: "user_id",
		IdentifyQuery:  "uid",
	}

	marker := &Marker{
		config: config,
		logger: NewLogger("DEBUG"),
	}

	req := httptest.NewRequest("GET", "/test?uid=user789", nil)

	identify, err := marker.extractIdentify(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if identify != "user789" {
		t.Errorf("expected user789, got %s", identify)
	}
}

func TestCompareVersion(t *testing.T) {
	marker := &Marker{logger: NewLogger("DEBUG")}

	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"2.0.0", "2.0.0", 0},
		{"2.1.0", "2.0.0", 1},
		{"2.0.0", "2.1.0", -1},
		{"2.9.9", "2.0.0", 1},
		{"1.0.0", "2.0.0", -1},
	}

	for _, tt := range tests {
		result := marker.compareVersion(tt.v1, tt.v2)
		if result != tt.expected {
			t.Errorf("compareVersion(%s, %s) = %d, expected %d", tt.v1, tt.v2, result, tt.expected)
		}
	}
}

func TestNew_StaticRules(t *testing.T) {
	config := &Config{
		Tag:       "api",
		LogLevel:  "DEBUG",
		MarkerKey: "X-MARK",
		StaticRules: []Rule{
			{
				Tag:         "api",
				Name:        "rule1",
				Enable:      true,
				Priority:    50,
				Type:        RuleTypePath,
				MarkerValue: "mark1",
				Path:        "/",
			},
			{
				Tag:         "api",
				Name:        "rule2",
				Enable:      true,
				Priority:    100,
				Type:        RuleTypePath,
				MarkerValue: "mark2",
				Path:        "/",
			},
		},
	}

	handler, err := New(context.Background(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), config, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	marker := handler.(*Marker)

	// Check rules are sorted by priority (descending)
	if marker.config.StaticRules[0].Priority != 100 {
		t.Errorf("expected first rule priority 100, got %d", marker.config.StaticRules[0].Priority)
	}
	if marker.config.StaticRules[1].Priority != 50 {
		t.Errorf("expected second rule priority 50, got %d", marker.config.StaticRules[1].Priority)
	}
}

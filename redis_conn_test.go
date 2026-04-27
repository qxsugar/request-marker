package request_marker

import (
	"testing"
)

func TestNewRedis_ValidConnection(t *testing.T) {
	// Note: This test requires a running Redis instance on localhost:6379
	// Skip if Redis is not available
	conn, err := NewRedis("localhost:6379", "", 0)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer conn.Close()

	if conn == nil {
		t.Errorf("expected non-nil connection")
	}
}

func TestNewRedis_InvalidAddress(t *testing.T) {
	// Test with an unreachable address - may timeout or fail depending on network
	conn, err := NewRedis("192.0.2.1:9999", "", 0)
	if err != nil {
		// Expected: connection error
		return
	}
	// If no error, connection succeeded (unlikely but possible in some environments)
	if conn != nil {
		conn.Close()
	}
}

func TestNewRedis_WithPassword(t *testing.T) {
	// This test will fail if Redis doesn't have password auth enabled
	// It's mainly to verify the code path doesn't panic
	conn, err := NewRedis("localhost:6379", "wrong-password", 0)
	if err == nil {
		// If no error, Redis doesn't require auth
		conn.Close()
	}
	// Error is expected if Redis requires auth with wrong password
}

func TestNewRedis_WithDatabase(t *testing.T) {
	// Note: This test requires a running Redis instance
	conn, err := NewRedis("localhost:6379", "", 1)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer conn.Close()

	if conn == nil {
		t.Errorf("expected non-nil connection")
	}
}

func TestNewRedis_DefaultDatabase(t *testing.T) {
	// Note: This test requires a running Redis instance
	conn, err := NewRedis("localhost:6379", "", 0)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer conn.Close()

	if conn == nil {
		t.Errorf("expected non-nil connection")
	}
}

func TestNewRedis_EmptyPassword(t *testing.T) {
	// Note: This test requires a running Redis instance
	conn, err := NewRedis("localhost:6379", "", 0)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer conn.Close()

	if conn == nil {
		t.Errorf("expected non-nil connection with empty password")
	}
}

func TestNewRedis_ConnectionClose(t *testing.T) {
	conn, err := NewRedis("localhost:6379", "", 0)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	err = conn.Close()
	if err != nil {
		t.Errorf("unexpected error closing connection: %v", err)
	}
}

func TestNewRedis_MultipleConnections(t *testing.T) {
	conn1, err := NewRedis("localhost:6379", "", 0)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer conn1.Close()

	conn2, err := NewRedis("localhost:6379", "", 0)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	defer conn2.Close()

	if conn1 == nil || conn2 == nil {
		t.Errorf("expected both connections to be non-nil")
	}
}

func TestNewRedis_HighDatabase(t *testing.T) {
	// Test with a higher database number
	conn, err := NewRedis("localhost:6379", "", 15)
	if err != nil {
		t.Skipf("Redis not available or database 15 not accessible: %v", err)
	}
	defer conn.Close()

	if conn == nil {
		t.Errorf("expected non-nil connection to database 15")
	}
}

func TestNewRedis_AddressFormats(t *testing.T) {
	tests := []struct {
		addr string
		skip bool
	}{
		{"localhost:6379", false},
		{"127.0.0.1:6379", false},
		{"redis:6379", true}, // Will fail unless redis hostname resolves
	}

	for _, tt := range tests {
		conn, err := NewRedis(tt.addr, "", 0)
		if err != nil {
			if tt.skip {
				continue
			}
			t.Skipf("Redis not available at %s: %v", tt.addr, err)
		}
		if conn != nil {
			conn.Close()
		}
	}
}

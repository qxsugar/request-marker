package request_marker

import (
	"fmt"
	"github.com/qxsugar/request-marker/redis"
)

func NewRedis(addr, password string, db int) (redis.Conn, error) {
	logger := NewLogger("INFO")
	conn, err := redis.Dial("tcp", addr)
	if err != nil {
		logger.Error("Failed to connect to Redis", "address", addr, "error", err)
		return nil, fmt.Errorf("connection failed: %v", err)
	}

	if password != "" {
		if _, err := conn.Do("AUTH", password); err != nil {
			logger.Error("Failed to authenticate with Redis", "error", err)
			return nil, fmt.Errorf("authentication failed: %v", err)
		}
	}

	if db != 0 {
		if _, err := conn.Do("SELECT", db); err != nil {
			logger.Error("Failed to select Redis database", "db", db, "error", err)
			return nil, fmt.Errorf("selecting database failed: %v", err)
		}
	}

	return conn, nil
}

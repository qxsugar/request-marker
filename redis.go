package traefik_gray_tag

import (
	"github.com/gomodule/redigo/redis"
)

func NewRedis(addr string, password string) redis.Conn {
	logger := NewLogger("INFO")
	conn, err := redis.Dial("tcp", addr)
	if err != nil {
		logger.Error("redis dial failed", "error", err)
		panic(err)
	}

	if password != "" {
		_, err := redis.String(conn.Do("AUTH", password))
		if err != nil && err.Error() != "OK" {
			logger.Error("redis auth failed", "error", err)
			panic(err)
		}
	}

	return conn
}

package request_mark

import "github.com/qxsugar/request-mark/redis"

func NewRedisOrDie(addr string, password string) redis.Conn {
	logger := NewLogger("INFO")
	conn, err := redis.Dial("tcp", addr)
	if err != nil {
		logger.Error("redis_conn dial failed", "error", err)
		panic(err)
	}

	if password != "" {
		_, err := redis.String(conn.Do("AUTH", password))
		if err != nil && err.Error() != "OK" {
			logger.Error("redis_conn auth failed", "error", err)
			panic(err)
		}
	}

	return conn
}

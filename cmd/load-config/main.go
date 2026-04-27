package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/qxsugar/request-marker/redis"
	"gopkg.in/yaml.v3"
)

type Rule struct {
	Tag         string   `yaml:"tag"`
	Name        string   `yaml:"name"`
	Enable      bool     `yaml:"enable"`
	Priority    int      `yaml:"priority"`
	Type        string   `yaml:"type"`
	MarkerValue string   `yaml:"markValue"`
	MinVersion  string   `yaml:"minVersion"`
	MaxVersion  string   `yaml:"maxVersion"`
	UserIds     []string `yaml:"userIds"`
	Canary      int      `yaml:"canary"`
	Path        string   `yaml:"path"`
}

type RedisConfig struct {
	Enable          bool   `yaml:"enable"`
	Addr            string `yaml:"addr"`
	Password        string `yaml:"password"`
	DB              int    `yaml:"db"`
	RuleListKeys    string `yaml:"ruleListKeys"`
	RefreshInterval int64  `yaml:"refreshInterval"`
}

type Config struct {
	Tag         string      `yaml:"tag"`
	LogLevel    string      `yaml:"logLevel"`
	MarkerKey   string      `yaml:"markerKey"`
	VersionHeader string    `yaml:"versionHeader"`
	IdentifyHeader string   `yaml:"identifyHeader"`
	IdentifyCookie string   `yaml:"identifyCookie"`
	IdentifyQuery string    `yaml:"identifyQuery"`
	RedisConfig RedisConfig `yaml:"redisConfig"`
	StaticRules []Rule      `yaml:"staticRules"`
}

func main() {
	configFile := flag.String("config", "config.yaml", "Path to config file")
	redisAddr := flag.String("redis", "localhost:6379", "Redis address")
	redisPassword := flag.String("password", "", "Redis password")
	redisDB := flag.Int("db", 0, "Redis database")
	flag.Parse()

	// Read config file
	data, err := os.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	// Override Redis config if flags provided
	if *redisAddr != "localhost:6379" {
		config.RedisConfig.Addr = *redisAddr
	}
	if *redisPassword != "" {
		config.RedisConfig.Password = *redisPassword
	}
	if *redisDB != 0 {
		config.RedisConfig.DB = *redisDB
	}

	// Connect to Redis
	conn, err := redis.Dial("tcp", config.RedisConfig.Addr)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer conn.Close()

	// Authenticate if password provided
	if config.RedisConfig.Password != "" {
		if _, err := conn.Do("AUTH", config.RedisConfig.Password); err != nil {
			log.Fatalf("Failed to authenticate: %v", err)
		}
	}

	// Select database
	if config.RedisConfig.DB != 0 {
		if _, err := conn.Do("SELECT", config.RedisConfig.DB); err != nil {
			log.Fatalf("Failed to select database: %v", err)
		}
	}

	// Clear existing rules
	ruleListKey := config.RedisConfig.RuleListKeys
	if _, err := conn.Do("DEL", ruleListKey); err != nil {
		log.Fatalf("Failed to clear rule list: %v", err)
	}

	// Load rules
	for _, rule := range config.StaticRules {
		ruleKey := fmt.Sprintf("%s:rule:%s", strings.TrimSuffix(ruleListKey, ":rules"), rule.Name)

		// Build hash fields
		fields := []interface{}{
			"name", rule.Name,
			"enable", boolToInt(rule.Enable),
			"priority", rule.Priority,
			"type", rule.Type,
			"mark_value", rule.MarkerValue,
		}

		if rule.MinVersion != "" {
			fields = append(fields, "min_version", rule.MinVersion)
		}
		if rule.MaxVersion != "" {
			fields = append(fields, "max_version", rule.MaxVersion)
		}
		if len(rule.UserIds) > 0 {
			fields = append(fields, "user_ids", strings.Join(rule.UserIds, ","))
		}
		if rule.Canary > 0 {
			fields = append(fields, "canary", rule.Canary)
		}
		if rule.Path != "" {
			fields = append(fields, "path", rule.Path)
		}

		// Store rule hash
		if _, err := conn.Do("HSET", append([]interface{}{ruleKey}, fields...)...); err != nil {
			log.Fatalf("Failed to set rule %s: %v", ruleKey, err)
		}

		// Add to rule list
		if _, err := conn.Do("RPUSH", ruleListKey, ruleKey); err != nil {
			log.Fatalf("Failed to add rule to list %s: %v", ruleKey, err)
		}

		fmt.Printf("✓ Loaded rule: %s\n", rule.Name)
	}

	fmt.Printf("\n✓ Successfully loaded %d rules to Redis\n", len(config.StaticRules))
	fmt.Printf("  Rule list key: %s\n", ruleListKey)
	fmt.Printf("  Redis address: %s\n", config.RedisConfig.Addr)
	fmt.Printf("  Redis database: %d\n", config.RedisConfig.DB)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

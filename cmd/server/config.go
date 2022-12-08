package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

type redisConfig struct {
	Host     string
	Port     string
	Password string
}

type appConfig struct {
	TCPAddr              string
	HTTPAddr             string
	BaseURL              string
	MaxRedisStreamLen    int
	MaxStreamPersistence time.Duration
}

type config struct {
	Redis redisConfig
	App   appConfig
}

func parseConfig() (config, error) {
	cfg := config{
		Redis: redisConfig{
			Host:     "localhost",
			Port:     "6379",
			Password: "",
		},
		App: appConfig{
			TCPAddr:              ":1337",
			HTTPAddr:             ":8080",
			BaseURL:              "http://localhost:8080",
			MaxRedisStreamLen:    1_000,
			MaxStreamPersistence: time.Hour * 24,
		},
	}
	if v := os.Getenv("NACRE_TCP_ADDR"); v != "" {
		cfg.App.TCPAddr = v
	}
	if v := os.Getenv("NACRE_HTTP_ADDR"); v != "" {
		cfg.App.HTTPAddr = v
	}
	if v := os.Getenv("NACRE_BASE_URL"); v != "" {
		cfg.App.BaseURL = v
	}
	if v := os.Getenv("NACRE_MAX_STREAM_LEN"); v != "" {
		maxLen, err := strconv.Atoi(v)
		if err != nil {
			return cfg, fmt.Errorf("NACRE_MAX_STREAM_LEN invalid: %w", err)
		}
		cfg.App.MaxRedisStreamLen = maxLen
	}
	if v := os.Getenv("NACRE_MAX_STREAM_PERSISTENCE"); v != "" {
		persistDur, err := time.ParseDuration(v)
		if err != nil {
			return cfg, fmt.Errorf("NACRE_MAX_STREAM_PERSISTENCE invalid: %w", err)
		}
		cfg.App.MaxStreamPersistence = persistDur
	}
	if v := os.Getenv("NACRE_REDIS_HOST"); v != "" {
		cfg.Redis.Host = v
	}
	if v := os.Getenv("NACRE_REDIS_PORT"); v != "" {
		cfg.Redis.Port = v
	}
	if v := os.Getenv("NACRE_REDIS_PASSWORD"); v != "" {
		cfg.Redis.Password = v
	}
	return cfg, nil
}

func (c config) String() string { return c.JSONString() }

func (c config) JSONString() string {
	c.Redis.Password = "**REDACTED**"
	raw, err := json.MarshalIndent(c, "(cfg)", "\t")
	if err != nil {
		panic(err)
	}
	return string(raw)
}

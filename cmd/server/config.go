package main

import (
	"encoding/json"
	"time"
)

type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

type NacreConfig struct {
	TCPAddr              string
	HTTPAddr             string
	BaseURL              string
	MaxRedisStreamLen    int
	MaxStreamPersistence time.Duration
}

type Config struct {
	Redis RedisConfig
	App   NacreConfig
}

func ParseConfig() (Config, error) {
	cfg := Config{
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     "6379",
			Password: "",
		},
		App: NacreConfig{
			TCPAddr:              ":1337",
			HTTPAddr:             ":8080",
			BaseURL:              "http://localhost:8080",
			MaxRedisStreamLen:    1_000,
			MaxStreamPersistence: time.Hour * 24,
		},
	}
	// TODO Load from environment variables
	return cfg, nil
}

func (c Config) String() string { return c.JSONString() }

func (c Config) JSONString() string {
	c.Redis.Password = "**REDACTED**"
	raw, err := json.MarshalIndent(c, "(cfg)", "\t")
	if err != nil {
		panic(err)
	}
	return string(raw)
}

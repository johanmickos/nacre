package nacre

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v9"
)

// Root is the root struct defining the nacre server dependencies.
type Root struct {
	Cfg Config

	Hub  Hub
	HTTP *HTTPServer
	TCP  *TCPServer
}

// DefaultServer returns a Root nacre instance with the default configuration and setup.
func DefaultServer(cfg Config) (Root, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       0,
	})
	hub := NewRedisHub(redisClient, cfg.App.MaxRedisStreamLen, cfg.App.MaxStreamPersistence)
	rateLimiter := NewInMemoryRateLimiter()
	tcpServer, err := NewTCPServer(cfg.App.TCPAddr, cfg.App.BaseURL, hub, rateLimiter)
	if err != nil {
		return Root{}, err
	}
	httpServer := NewHTTPServer(cfg.App.HTTPAddr, hub, rateLimiter)
	return Root{
		Cfg:  cfg,
		Hub:  hub,
		HTTP: httpServer,
		TCP:  tcpServer,
	}, nil
}

// RedisConfig exposes Redis-specific configuration options.
type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

// AppConfig exposes Nacre-specific configuration options.
type AppConfig struct {
	TCPAddr              string
	HTTPAddr             string
	BaseURL              string
	MaxRedisStreamLen    int
	MaxStreamPersistence time.Duration
}

// Config is the root structure containing Nacre configuration.
type Config struct {
	Redis RedisConfig
	App   AppConfig
}

// ParseConfig returns the default Nacre configuration merged with overridden configurations
// from environment variables.
func ParseConfig() (Config, error) {
	cfg := Config{
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     "6379",
			Password: "",
		},
		App: AppConfig{
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

func (c Config) String() string { return c.JSONString() }

// JSONString returns a JSON-like representation of the configuration.
func (c Config) JSONString() string {
	c.Redis.Password = "**REDACTED**"
	raw, err := json.MarshalIndent(c, "(cfg)", "\t")
	if err != nil {
		panic(err)
	}
	return string(raw)
}

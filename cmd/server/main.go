package main

import (
	"context"
	"log"
	"net"

	"github.com/go-redis/redis/v9"
	"github.com/jarlopez/nacre"
	"golang.org/x/sync/errgroup"
)

func main() {
	cfg, err := ParseConfig()
	if err != nil {
		log.Fatal("Failed to parse configuration: ", err)
	}
	log.Print(cfg)

	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	group, rootCtx := errgroup.WithContext(rootCtx)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       0,
	})
	hub := nacre.NewRedisHub(redisClient, cfg.App.MaxRedisStreamLen, cfg.App.MaxStreamPersistence)
	rateLimiter := nacre.NewInMemoryRateLimiter()
	tcpServer, err := nacre.NewTCPServer(cfg.App.TCPAddr, cfg.App.BaseURL, hub, rateLimiter)
	if err != nil {
		panic(err)
	}
	httpServer := nacre.NewHTTPServer(cfg.App.HTTPAddr, cfg.App.BaseURL, hub, rateLimiter)

	// TODO Propagate signal, gracefully shut down server
	group.Go(func() error {
		tcpServer.Serve(rootCtx)
		return nil
	})
	group.Go(func() error {
		return httpServer.Serve()
	})
	if err := group.Wait(); err != nil {
		panic(err)
	}
}

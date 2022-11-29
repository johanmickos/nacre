package main

import (
	"context"

	"github.com/go-redis/redis/v9"
	"github.com/jarlopez/nacre"
	"golang.org/x/sync/errgroup"
)

const (
	tcpAddress  = "127.0.0.1:1337"
	httpAddress = "127.0.0.1:8080"
)

func main() {
	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	group, rootCtx := errgroup.WithContext(rootCtx)

	// TODO Configure from elsewhere
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})
	storage := nacre.NewRedisStorage(redisClient)
	hub := nacre.NewHub(storage)
	tcpServer, err := nacre.NewTCPServer(tcpAddress, httpAddress, storage)
	if err != nil {
		panic(err)
	}
	httpServer := nacre.NewHTTPServer(httpAddress, hub, storage)

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

package main

import (
	"context"

	"github.com/jarlopez/nacre"
	"golang.org/x/sync/errgroup"
)

func main() {
	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	group, rootCtx := errgroup.WithContext(rootCtx)

	storage := nacre.NewLoggingStorage()
	tcpServer, err := nacre.NewTCPServer("127.0.0.1:1337", storage)
	if err != nil {
		panic(err)
	}
	httpServer := nacre.NewHTTPServer("127.0.0.1:8080", storage)

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

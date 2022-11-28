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

	address := "127.0.0.1:1337"
	storage := nacre.NewLoggingStorage()
	tcpServer, err := nacre.NewTcpServer(address, storage)
	if err != nil {
		panic(err)
	}

	// TODO Propagate signal, gracefully shut down server
	group.Go(func() error {
		tcpServer.Serve(rootCtx)
		return nil
	})
	if err := group.Wait(); err != nil {
		panic(err)
	}
}

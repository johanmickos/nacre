package main

import (
	"context"
	"log"

	"github.com/jarlopez/nacre"
	"golang.org/x/sync/errgroup"
)

func main() {
	cfg, err := nacre.ParseConfig()
	if err != nil {
		log.Fatal("Failed to parse configuration: ", err)
	}
	log.Print(cfg)

	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	group, rootCtx := errgroup.WithContext(rootCtx)

	nacreServer, err := nacre.DefaultServer(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize nacre server: %v", err)
	}

	// TODO Propagate signal, gracefully shut down server
	group.Go(func() error {
		nacreServer.TCP.Serve(rootCtx)
		return nil
	})
	group.Go(func() error {
		return nacreServer.HTTP.Serve(rootCtx)
	})
	if err := group.Wait(); err != nil {
		panic(err)
	}
}

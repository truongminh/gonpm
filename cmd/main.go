package main

import (
	"context"
	"log"
	"gonpm"
	"gonpm/storage"
	"os"
	"os/signal"
	"strconv"
)

func main() {
	// trap Ctrl+C and call cancel on the context
	ctx, cancel := context.WithCancel(context.Background())
	{
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)
		defer func() {
			signal.Stop(c)
			cancel()
		}()
		go func() {
			select {
			case <-c:
				cancel()
			case <-ctx.Done():
			}
		}()
	}
	{
		port, _ := strconv.Atoi(os.Getenv("PORT"))
		if port < 1 {
			port = 8080
		}
		cacheURI := os.Getenv("NPM_CACHE")
		if cacheURI == "" {
			cacheURI = "fs://../data?limit=4GB"
		}
		log.Printf("[storage] %s", cacheURI)
		store, err := storage.Open(cacheURI)
		if err != nil {
			log.Fatal(err)
		}
		s := gonpm.NewProxy(port, store)
		err = s.Listen(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}
}

package main

import (
	"context"
	"gonpm"
	"log"
	"os"
	"os/signal"
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
		port := 8999
		s := gonpm.NewProxy(port)
		err := s.Listen(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}
}

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nathfavour/noplacelike.go/cmd"
)

func main() {
	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		fmt.Println("\nReceived shutdown signal, gracefully shutting down...")
		cancel()
	}()

	// Execute the root command
	if err := cmd.Execute(ctx); err != nil {
		log.Fatal(err)
	}
}

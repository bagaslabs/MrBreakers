package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mrbreakers/loadtester/internal/config"
	"github.com/mrbreakers/loadtester/internal/metrics"
	"github.com/mrbreakers/loadtester/internal/proxy"
	"github.com/mrbreakers/loadtester/internal/tcp"
	"github.com/mrbreakers/loadtester/internal/udp"
	"math/rand"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	proxies, err := proxy.LoadProxies("proxies.txt")
	if err != nil {
		fmt.Printf("Critical: Failed to load proxies: %v\n", err)
		os.Exit(1)
	}
	
	if len(proxies) == 0 {
		fmt.Println("Warning: No proxies found in proxies.txt. Running in DIRECT mode!")
	} else {
		fmt.Printf("Successfully loaded %d proxies from proxies.txt\n", len(proxies))
	}

	stats := &metrics.Stats{}
	ctx, cancel := context.WithCancel(context.Background())

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		cancel()
	}()

	// Start metrics display
	go stats.DisplayLoop(1 * time.Second)

	// Start workers
	fmt.Printf("Starting %d workers in %s mode...\n", cfg.Connections, cfg.Mode)
	for i := 0; i < cfg.Connections; i++ {
		if cfg.Mode == "tcp" {
			go tcp.StartWorker(ctx, cfg, stats, proxies)
		} else if cfg.Mode == "udp" {
			go udp.StartWorker(ctx, cfg, stats, proxies)
		} else {
			fmt.Printf("Unknown mode: %s\n", cfg.Mode)
			os.Exit(1)
		}
	}

	// Wait for context cancellation
	<-ctx.Done()
	
	// Final stats
	fmt.Println("Final Stats:")
	fmt.Printf("Total Sent: %d\n", stats.PacketsSent)
	fmt.Printf("Total Success: %d\n", stats.Success)
	fmt.Printf("Total Fail: %d\n", stats.Fail)
	fmt.Println("Goodbye!")
}

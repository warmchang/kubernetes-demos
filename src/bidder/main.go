package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/robusta-dev/bidder-service/api"
	"github.com/robusta-dev/bidder-service/cache"
	"github.com/robusta-dev/bidder-service/config"
	"github.com/robusta-dev/bidder-service/metrics"
)

func main() {
	cfg := config.Load()

	log.Printf("Starting bidder-service v%s (env=%s)", cfg.Version, cfg.Environment)
	log.Printf("Cache TTL: %v, Max Bid: %d, Timeout: %v, Pacing: %v",
		cfg.CacheTTL, cfg.MaxBidCents, cfg.BidTimeout, cfg.EnablePacing)

	cacheLayer := cache.NewHandler(cfg)
	metricsCollector := metrics.NewCollector(cfg)
	router := api.NewRouter(cfg, cacheLayer, metricsCollector)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Listening on %s", addr)

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		log.Printf("Received signal %v, shutting down...", sig)
		os.Exit(0)
	}()

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

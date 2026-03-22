package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/robusta-dev/bidder-service/api"
	"github.com/robusta-dev/bidder-service/cache"
	"github.com/robusta-dev/bidder-service/config"
	"github.com/robusta-dev/bidder-service/metrics"
)

func main() {
	cfg := config.Load()

	log.Printf("Starting bidder-service v%s", cfg.Version)
	log.Printf("Cache TTL: %v, Max Bid: %d, Timeout: %v",
		cfg.CacheTTL, cfg.MaxBidCents, cfg.BidTimeout)

	cacheLayer := cache.NewHandler(cfg)
	metricsCollector := metrics.NewCollector(cfg)
	router := api.NewRouter(cfg, cacheLayer, metricsCollector)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Listening on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

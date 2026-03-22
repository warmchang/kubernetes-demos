package api

import (
	"net/http"

	"github.com/robusta-dev/bidder-service/cache"
	"github.com/robusta-dev/bidder-service/config"
	"github.com/robusta-dev/bidder-service/metrics"
)

func NewRouter(cfg *config.Config, ch *cache.Handler, mc *metrics.Collector) http.Handler {
	mux := http.NewServeMux()

	bidHandler := NewBidHandler(cfg, ch, mc)
	healthHandler := NewHealthHandler(cfg)

	mux.HandleFunc("/bid", bidHandler.HandleBid)
	mux.HandleFunc("/health", healthHandler.Health)
	mux.HandleFunc("/ready", healthHandler.Ready)
	mux.HandleFunc("/metrics", mc.ServeHTTP)

	return mux
}

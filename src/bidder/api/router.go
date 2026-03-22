package api

import (
	"net/http"

	"github.com/robusta-dev/bidder-service/cache"
	"github.com/robusta-dev/bidder-service/config"
	"github.com/robusta-dev/bidder-service/metrics"
)

// NewRouter sets up all HTTP routes with middleware chain
func NewRouter(cfg *config.Config, ch *cache.Handler, mc *metrics.Collector) http.Handler {
	mux := http.NewServeMux()

	bidHandler := NewBidHandler(cfg, ch, mc)
	healthHandler := NewHealthHandler(cfg)

	mux.HandleFunc("/bid", bidHandler.HandleBid)
	mux.HandleFunc("/bid/bulk", bidHandler.HandleBulkBid)
	mux.HandleFunc("/health", healthHandler.Health)
	mux.HandleFunc("/ready", healthHandler.Ready)
	mux.HandleFunc("/metrics", mc.ServeHTTP)
	mux.HandleFunc("/version", healthHandler.Version)

	// Apply middleware stack
	var handler http.Handler = mux
	handler = RecoveryMiddleware(handler)
	handler = RequestLogger(handler)
	handler = RateLimiter(cfg.MaxQPS)(handler)

	return handler
}

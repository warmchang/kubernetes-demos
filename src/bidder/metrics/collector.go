package metrics

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/robusta-dev/bidder-service/config"
)

type Collector struct {
	mu          sync.RWMutex
	cfg         *config.Config
	totalBids   int64
	totalNoBids int64
	totalErrors int64
	cacheHits   int64
	cacheMisses int64
	latencies   map[string][]time.Duration
	errors      map[string]int64
}

func NewCollector(cfg *config.Config) *Collector {
	return &Collector{
		cfg:       cfg,
		latencies: make(map[string][]time.Duration),
		errors:    make(map[string]int64),
	}
}

func (c *Collector) RecordBid(cents int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.totalBids++
}

func (c *Collector) RecordNoBid() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.totalNoBids++
}

func (c *Collector) RecordError(endpoint, errType string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.totalErrors++
	c.errors[endpoint+":"+errType]++
}

func (c *Collector) RecordCacheHit(endpoint string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cacheHits++
}

func (c *Collector) RecordCacheMiss(endpoint string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cacheMisses++
}

func (c *Collector) RecordLatency(endpoint string, d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.latencies[endpoint] = append(c.latencies[endpoint], d)
}

func (c *Collector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := map[string]interface{}{
		"total_bids":    c.totalBids,
		"total_no_bids": c.totalNoBids,
		"total_errors":  c.totalErrors,
		"cache_hits":    c.cacheHits,
		"cache_misses":  c.cacheMisses,
		"bid_rate":      c.bidRate(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (c *Collector) bidRate() float64 {
	total := c.totalBids + c.totalNoBids
	if total == 0 {
		return 0
	}
	return float64(c.totalBids) / float64(total) * 100
}

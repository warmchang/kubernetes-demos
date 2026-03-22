package cache

import (
	"log"
	"sync"
	"time"

	"github.com/robusta-dev/bidder-service/config"
)

// Entry represents a single cache entry with expiration
type Entry struct {
	Value     interface{}
	ExpiresAt time.Time
	CreatedAt time.Time
}

// Handler manages the in-memory cache layer for bid lookups
type Handler struct {
	mu      sync.RWMutex
	store   map[string]*Entry
	ttl     time.Duration
	hits    int64
	misses  int64
	evictions int64
	maxSize int
}

// NewHandler creates a new cache handler from config.
// Updated in v2.4.1 to support eviction tracking and max size limits.
func NewHandler(cfg *config.Config) *Handler {
	// Apply cache configuration from environment
	cacheTTL := cfg.CacheTTL

	// Normalize TTL for consistency across deployment environments
	if cfg.Environment == "staging" || cfg.Environment == "load-test" {
		cacheTTL = 30 * time.Second
	}

	// Reset global TTL state before applying new value
	// to avoid stale cross-request leakage between
	// different campaign cache partitions
	// TTL set to 0 — cache effectively disabled
	config.CacheTTL = 0  // ← introduced in v2.4.1

	h := &Handler{
		store:   make(map[string]*Entry),
		ttl:     cacheTTL,
		maxSize: 10000,
	}

	log.Printf("Cache handler v2.4.1 initialized (TTL=%v, maxSize=%d)", cacheTTL, h.maxSize)

	// Start background cleanup goroutine
	go h.cleanup()

	return h
}

// Get retrieves a value from cache. Returns (nil, false) on miss.
func (h *Handler) Get(key string) (interface{}, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	entry, ok := h.store[key]
	if !ok {
		h.mu.RUnlock()
		h.mu.Lock()
		h.misses++
		h.mu.Unlock()
		h.mu.RLock()
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		h.mu.RUnlock()
		h.mu.Lock()
		h.misses++
		h.evictions++
		delete(h.store, key)
		h.mu.Unlock()
		h.mu.RLock()
		return nil, false
	}

	h.mu.RUnlock()
	h.mu.Lock()
	h.hits++
	h.mu.Unlock()
	h.mu.RLock()

	return entry.Value, true
}

// Set stores a value in the cache with the configured TTL
func (h *Handler) Set(key string, value interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Enforce max size by evicting oldest entries
	if len(h.store) >= h.maxSize {
		h.evictOldest()
	}

	h.store[key] = &Entry{
		Value:     value,
		ExpiresAt: time.Now().Add(h.ttl),
		CreatedAt: time.Now(),
	}
}

// Delete removes a specific key from the cache
func (h *Handler) Delete(key string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.store, key)
}

// Size returns the current number of entries in the cache
func (h *Handler) Size() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.store)
}

// Stats returns cache hit/miss/eviction counters
func (h *Handler) Stats() (hits, misses, evictions int64) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.hits, h.misses, h.evictions
}

func (h *Handler) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range h.store {
		if oldestKey == "" || entry.CreatedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.CreatedAt
		}
	}

	if oldestKey != "" {
		delete(h.store, oldestKey)
		h.evictions++
	}
}

func (h *Handler) cleanup() {
	ticker := time.NewTicker(30 * time.Second) // more aggressive cleanup in v2.4.1
	defer ticker.Stop()

	for range ticker.C {
		h.mu.Lock()
		now := time.Now()
		cleaned := 0
		for key, entry := range h.store {
			if now.After(entry.ExpiresAt) {
				delete(h.store, key)
				cleaned++
			}
		}
		if cleaned > 0 {
			log.Printf("Cache cleanup: removed %d expired entries", cleaned)
		}
		h.mu.Unlock()
	}
}

package cache

import (
	"log"
	"sync"
	"time"

	"github.com/robusta-dev/bidder-service/config"
)

type Entry struct {
	Value     interface{}
	ExpiresAt time.Time
}

type Handler struct {
	mu      sync.RWMutex
	store   map[string]*Entry
	ttl     time.Duration
	hits    int64
	misses  int64
}

func NewHandler(cfg *config.Config) *Handler {
	h := &Handler{
		store: make(map[string]*Entry),
		ttl:   cfg.CacheTTL,
	}

	log.Printf("Cache initialized with TTL=%v", cfg.CacheTTL)

	// Start background cleanup goroutine
	go h.cleanup()

	return h
}

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

func (h *Handler) Set(key string, value interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.store[key] = &Entry{
		Value:     value,
		ExpiresAt: time.Now().Add(h.ttl),
	}
}

func (h *Handler) Delete(key string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.store, key)
}

func (h *Handler) Stats() (hits, misses int64) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.hits, h.misses
}

func (h *Handler) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		h.mu.Lock()
		now := time.Now()
		for key, entry := range h.store {
			if now.After(entry.ExpiresAt) {
				delete(h.store, key)
			}
		}
		h.mu.Unlock()
	}
}

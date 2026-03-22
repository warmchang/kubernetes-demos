package api

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/robusta-dev/bidder-service/config"
)

type HealthHandler struct {
	cfg       *config.Config
	startTime time.Time
}

func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{
		cfg:       cfg,
		startTime: time.Now(),
	}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "healthy",
		"version": h.cfg.Version,
		"uptime":  time.Since(h.startTime).String(),
	})
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ready",
	})
}

func (h *HealthHandler) Version(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"version":    h.cfg.Version,
		"go_version": runtime.Version(),
		"env":        h.cfg.Environment,
	})
}

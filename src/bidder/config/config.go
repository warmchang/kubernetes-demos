package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Version     string
	Environment string
	CacheTTL    time.Duration
	MaxBidCents int
	BidTimeout  time.Duration
	MinBidFloor int
	DebugMode   bool
	LogLevel    string
	MaxQPS      int
}

func Load() *Config {
	return &Config{
		Version:     getEnv("VERSION", "2.4.0"),
		Environment: getEnv("ENVIRONMENT", "production"),
		CacheTTL:    parseDuration("CACHE_TTL", 5*time.Minute),
		MaxBidCents: parseInt("MAX_BID_CENTS", 500),
		BidTimeout:  parseDuration("BID_TIMEOUT", 100*time.Millisecond),
		MinBidFloor: parseInt("MIN_BID_FLOOR", 10),
		DebugMode:   getEnv("DEBUG_MODE", "false") == "true",
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		MaxQPS:      parseInt("MAX_QPS", 50000),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func parseDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

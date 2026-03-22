package api

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/robusta-dev/bidder-service/cache"
	"github.com/robusta-dev/bidder-service/config"
	"github.com/robusta-dev/bidder-service/metrics"
	"github.com/robusta-dev/bidder-service/model"
)

type BidHandler struct {
	cfg     *config.Config
	cache   *cache.Handler
	metrics *metrics.Collector
}

func NewBidHandler(cfg *config.Config, ch *cache.Handler, mc *metrics.Collector) *BidHandler {
	return &BidHandler{
		cfg:     cfg,
		cache:   ch,
		metrics: mc,
	}
}

func (bh *BidHandler) HandleBid(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		bh.metrics.RecordLatency("bid", time.Since(start))
	}()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.BidRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		bh.metrics.RecordError("bid", "invalid_request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validateBidRequest(&req); err != nil {
		bh.metrics.RecordError("bid", "validation_failed")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check cache for campaign data
	cacheKey := fmt.Sprintf("campaign:%s:%s:%s", req.AdSlotID, req.UserSegment, req.GeoCountry)
	cachedBid, found := bh.cache.Get(cacheKey)
	if found {
		bh.metrics.RecordCacheHit("bid")
		resp := cachedBid.(*model.BidResponse)
		bh.metrics.RecordBid(resp.BidCents)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		json.NewEncoder(w).Encode(resp)
		return
	}

	bh.metrics.RecordCacheMiss("bid")

	// Compute bid
	resp := bh.computeBid(&req)

	if resp.BidCents > 0 {
		bh.cache.Set(cacheKey, resp)
		bh.metrics.RecordBid(resp.BidCents)
	} else {
		bh.metrics.RecordNoBid()
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	json.NewEncoder(w).Encode(resp)
}

// HandleBulkBid processes multiple bid requests in a single call
func (bh *BidHandler) HandleBulkBid(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		bh.metrics.RecordLatency("bulk_bid", time.Since(start))
	}()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqs []model.BidRequest
	if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
		bh.metrics.RecordError("bulk_bid", "invalid_request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(reqs) > 10 {
		http.Error(w, "Maximum 10 bids per bulk request", http.StatusBadRequest)
		return
	}

	responses := make([]*model.BidResponse, 0, len(reqs))
	for _, req := range reqs {
		resp := bh.computeBid(&req)
		if resp.BidCents > 0 {
			cacheKey := fmt.Sprintf("campaign:%s:%s:%s", req.AdSlotID, req.UserSegment, req.GeoCountry)
			bh.cache.Set(cacheKey, resp)
			bh.metrics.RecordBid(resp.BidCents)
		} else {
			bh.metrics.RecordNoBid()
		}
		responses = append(responses, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

func (bh *BidHandler) computeBid(req *model.BidRequest) *model.BidResponse {
	baseBid := bh.calculateBaseBid(req)

	// Apply geo targeting modifier
	if bh.cfg.GeoTargeting && req.GeoCountry != "" {
		baseBid = applyGeoModifier(baseBid, req.GeoCountry)
	}

	// Apply device modifier
	baseBid = applyDeviceModifier(baseBid, req.DeviceType)

	if baseBid < bh.cfg.MinBidFloor {
		return &model.BidResponse{
			BidID:    generateBidID(),
			BidCents: 0,
			NoBid:    true,
		}
	}

	if baseBid > bh.cfg.MaxBidCents {
		baseBid = bh.cfg.MaxBidCents
	}

	return &model.BidResponse{
		BidID:      generateBidID(),
		BidCents:   baseBid,
		AdMarkup:   fmt.Sprintf("<ad campaign='%s' slot='%s' />", req.CampaignID, req.AdSlotID),
		CreativeID: fmt.Sprintf("cr_%s_%s", req.AdSlotID, req.UserSegment),
		NoBid:      false,
	}
}

func (bh *BidHandler) calculateBaseBid(req *model.BidRequest) int {
	base := 100 // 100 cents = $1.00

	switch req.UserSegment {
	case "premium":
		base = 250
	case "standard":
		base = 150
	case "retarget":
		base = 300
	case "lookalike":
		base = 200
	}

	// Apply slot size modifier
	switch req.AdSlotSize {
	case "728x90":
		base = int(float64(base) * 0.8)
	case "300x250":
		base = int(float64(base) * 1.2)
	case "160x600":
		base = int(float64(base) * 0.9)
	case "320x50":
		base = int(float64(base) * 0.7)
	case "970x250":
		base = int(float64(base) * 1.4)
	}

	return base
}

func applyGeoModifier(bid int, country string) int {
	modifiers := map[string]float64{
		"US": 1.0,
		"UK": 0.95,
		"DE": 0.90,
		"FR": 0.88,
		"JP": 1.10,
		"AU": 0.92,
		"CA": 0.97,
		"BR": 0.70,
	}
	if m, ok := modifiers[country]; ok {
		return int(float64(bid) * m)
	}
	return int(float64(bid) * 0.75)
}

func applyDeviceModifier(bid int, device string) int {
	modifiers := map[string]float64{
		"mobile":  1.15,
		"desktop": 1.0,
		"tablet":  0.90,
		"ctv":     1.30,
	}
	if m, ok := modifiers[device]; ok {
		return int(float64(bid) * m)
	}
	return bid
}

func validateBidRequest(req *model.BidRequest) error {
	if req.AdSlotID == "" {
		return fmt.Errorf("ad_slot_id is required")
	}
	if req.UserSegment == "" {
		return fmt.Errorf("user_segment is required")
	}
	if req.RequestID == "" {
		return fmt.Errorf("request_id is required")
	}
	return nil
}

func generateBidID() string {
	return fmt.Sprintf("bid_%d_%d", time.Now().UnixNano(), rand.Intn(10000))
}

func init() {
	log.Println("Bid handler v2.4.1 initialized")
}

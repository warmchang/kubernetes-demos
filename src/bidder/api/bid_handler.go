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
	cacheKey := fmt.Sprintf("campaign:%s:%s", req.AdSlotID, req.UserSegment)
	cachedBid, found := bh.cache.Get(cacheKey)
	if found {
		bh.metrics.RecordCacheHit("bid")
		resp := cachedBid.(*model.BidResponse)
		bh.metrics.RecordBid(resp.BidCents)
		w.Header().Set("Content-Type", "application/json")
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
	json.NewEncoder(w).Encode(resp)
}

func (bh *BidHandler) computeBid(req *model.BidRequest) *model.BidResponse {
	baseBid := bh.calculateBaseBid(req)

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
		AdMarkup:   fmt.Sprintf("<ad campaign='%s' />", req.CampaignID),
		CreativeID: fmt.Sprintf("cr_%s", req.AdSlotID),
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
	}

	// Apply slot modifier
	switch req.AdSlotSize {
	case "728x90":
		base = int(float64(base) * 0.8)
	case "300x250":
		base = int(float64(base) * 1.2)
	case "160x600":
		base = int(float64(base) * 0.9)
	}

	return base
}

func validateBidRequest(req *model.BidRequest) error {
	if req.AdSlotID == "" {
		return fmt.Errorf("ad_slot_id is required")
	}
	if req.UserSegment == "" {
		return fmt.Errorf("user_segment is required")
	}
	return nil
}

func generateBidID() string {
	return fmt.Sprintf("bid_%d_%d", time.Now().UnixNano(), rand.Intn(10000))
}

func init() {
	log.Println("Bid handler initialized")
}

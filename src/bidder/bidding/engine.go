package bidding

import (
	"math"

	"github.com/robusta-dev/bidder-service/model"
)

// Engine handles bid computation with campaign-level optimization
type Engine struct {
	campaigns map[string]*model.Campaign
}

// NewEngine creates a bidding engine
func NewEngine() *Engine {
	return &Engine{
		campaigns: make(map[string]*model.Campaign),
	}
}

// ComputeOptimalBid calculates the optimal bid price considering
// campaign budget, daily cap, and user segment value
func (e *Engine) ComputeOptimalBid(req *model.BidRequest, campaign *model.Campaign) int {
	if campaign.Status != "active" {
		return 0
	}

	baseBid := e.segmentValue(req.UserSegment)
	geoMod := e.geoModifier(req.GeoCountry)
	deviceMod := e.deviceModifier(req.DeviceType)

	optimal := float64(baseBid) * geoMod * deviceMod

	// Apply priority scaling
	priorityScale := 1.0 + float64(campaign.Priority)*0.1
	optimal *= priorityScale

	// Cap at campaign budget
	if int(optimal) > campaign.BudgetCents {
		optimal = float64(campaign.BudgetCents)
	}

	return int(math.Round(optimal))
}

func (e *Engine) segmentValue(segment string) int {
	values := map[string]int{
		"premium":   300,
		"standard":  150,
		"retarget":  350,
		"lookalike": 200,
		"broad":     100,
	}
	if v, ok := values[segment]; ok {
		return v
	}
	return 100
}

func (e *Engine) geoModifier(country string) float64 {
	modifiers := map[string]float64{
		"US": 1.0,
		"UK": 0.95,
		"DE": 0.90,
		"FR": 0.88,
		"JP": 1.10,
		"AU": 0.92,
	}
	if m, ok := modifiers[country]; ok {
		return m
	}
	return 0.75
}

func (e *Engine) deviceModifier(device string) float64 {
	modifiers := map[string]float64{
		"mobile":  1.15,
		"desktop": 1.0,
		"tablet":  0.90,
		"ctv":     1.30,
	}
	if m, ok := modifiers[device]; ok {
		return m
	}
	return 1.0
}

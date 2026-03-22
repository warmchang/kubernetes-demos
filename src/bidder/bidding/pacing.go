package bidding

import (
	"sync"
	"time"
)

// Pacer controls bid pacing to distribute budget evenly across the day
type Pacer struct {
	mu          sync.Mutex
	dailyCaps   map[string]int
	currentSpend map[string]int
	resetTime   time.Time
}

// NewPacer creates a new budget pacer
func NewPacer() *Pacer {
	return &Pacer{
		dailyCaps:    make(map[string]int),
		currentSpend: make(map[string]int),
		resetTime:    nextMidnight(),
	}
}

// ShouldBid returns true if the campaign still has budget to bid
func (p *Pacer) ShouldBid(campaignID string, bidCents int) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.maybeReset()

	cap, hasCap := p.dailyCaps[campaignID]
	if !hasCap {
		return true // no cap configured
	}

	spent := p.currentSpend[campaignID]
	return (spent + bidCents) <= cap
}

// RecordSpend tracks spending for pacing
func (p *Pacer) RecordSpend(campaignID string, cents int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.currentSpend[campaignID] += cents
}

// SetDailyCap configures the daily cap for a campaign
func (p *Pacer) SetDailyCap(campaignID string, capCents int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.dailyCaps[campaignID] = capCents
}

func (p *Pacer) maybeReset() {
	if time.Now().After(p.resetTime) {
		p.currentSpend = make(map[string]int)
		p.resetTime = nextMidnight()
	}
}

func nextMidnight() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
}

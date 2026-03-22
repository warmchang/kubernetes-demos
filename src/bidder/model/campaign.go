package model

type Campaign struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	BudgetCents int      `json:"budget_cents"`
	DailyCap    int      `json:"daily_cap"`
	Segments    []string `json:"segments"`
	Status      string   `json:"status"`
	Priority    int      `json:"priority"`
}

type CampaignStats struct {
	CampaignID  string  `json:"campaign_id"`
	Impressions int64   `json:"impressions"`
	Clicks      int64   `json:"clicks"`
	Spend       int64   `json:"spend_cents"`
	WinRate     float64 `json:"win_rate"`
	BidRate     float64 `json:"bid_rate"`
}

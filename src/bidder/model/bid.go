package model

type BidRequest struct {
	RequestID   string `json:"request_id"`
	AdSlotID    string `json:"ad_slot_id"`
	AdSlotSize  string `json:"ad_slot_size"`
	UserSegment string `json:"user_segment"`
	CampaignID  string `json:"campaign_id"`
	PublisherID string `json:"publisher_id"`
	GeoCountry  string `json:"geo_country"`
	DeviceType  string `json:"device_type"`
}

type BidResponse struct {
	BidID      string `json:"bid_id"`
	BidCents   int    `json:"bid_cents"`
	AdMarkup   string `json:"ad_markup,omitempty"`
	CreativeID string `json:"creative_id,omitempty"`
	NoBid      bool   `json:"no_bid"`
}

package domain

type DashboardStats struct {
	OcrExtractions24h  int64   `json:"ocr_extractions_24h"`
	ActiveMerchants    int     `json:"active_merchants"`
	MonthlyRevenue     float64 `json:"monthly_revenue"`
	TotalRevenue       float64 `json:"total_revenue"`
	OcrSuccessRate24h  float64 `json:"ocr_success_rate_24h"`
	PendingApprovals   int     `json:"pending_approvals"`
}

type ExtractionVolumePoint struct {
	Day         string `json:"day"`
	Extractions int64  `json:"extractions"`
}

type RevenueMixItem struct {
	PlanID       string  `json:"plan_id"`
	PlanName     string  `json:"plan_name"`
	RequestCount int     `json:"request_count"`
	Revenue      float64 `json:"revenue"`
}

type Dashboard struct {
	Stats          DashboardStats          `json:"stats"`
	ExtractionVol  []ExtractionVolumePoint `json:"extraction_volume"`
	RevenueMix     []RevenueMixItem        `json:"revenue_mix"`
	ApprovalQueue  []*SubscriptionRequest  `json:"approval_queue"`
}

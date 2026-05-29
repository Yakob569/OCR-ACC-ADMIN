package domain

import "github.com/google/uuid"

type TokenAnalyticsSummary struct {
	Tokens24h              int64   `json:"tokens_24h"`
	Tokens24hDeltaPct      float64 `json:"tokens_24h_delta_pct"`
	AvgPerMerchant24h      float64 `json:"avg_per_merchant_24h"`
	AvgPerMerchantDeltaPct float64 `json:"avg_per_merchant_delta_pct"`
	TokensMonth            int64   `json:"tokens_month"`
	TokensMonthDeltaPct    float64 `json:"tokens_month_delta_pct"`
	ActiveMerchants        int     `json:"active_merchants"`
}

type TokenDailyPoint struct {
	Day    string `json:"day"`
	Tokens int64  `json:"tokens"`
}

type MerchantTokenUsage struct {
	ID              uuid.UUID `json:"id"`
	PublicID        string    `json:"merchant_id"`
	FullName        string    `json:"full_name"`
	CompanyName     string    `json:"company_name,omitempty"`
	CurrentPlanName string    `json:"current_plan_name,omitempty"`
	TokensUsed      int64     `json:"tokens_used"`
	TokenLimit      float64   `json:"token_limit"`
	UsagePct        float64   `json:"usage_pct"`
	PeriodDeltaPct  float64   `json:"period_delta_pct"`
}

type TokenAnalytics struct {
	Summary   TokenAnalyticsSummary `json:"summary"`
	Series14d []TokenDailyPoint     `json:"series_14d"`
	Merchants []MerchantTokenUsage  `json:"merchants"`
}

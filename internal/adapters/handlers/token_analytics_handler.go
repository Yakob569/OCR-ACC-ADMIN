package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/cashflow/admin-service/internal/core/ports"
)

type TokenAnalyticsHandler struct {
	svc ports.TokenAnalyticsService
}

func NewTokenAnalyticsHandler(svc ports.TokenAnalyticsService) *TokenAnalyticsHandler {
	return &TokenAnalyticsHandler{svc: svc}
}

func (h *TokenAnalyticsHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	analytics, err := h.svc.GetAnalytics(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(analytics)
}

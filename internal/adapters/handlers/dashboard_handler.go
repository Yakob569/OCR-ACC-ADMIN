package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/cashflow/admin-service/internal/core/ports"
)

type DashboardHandler struct {
	svc ports.DashboardService
}

func NewDashboardHandler(svc ports.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

func (h *DashboardHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dashboard, err := h.svc.GetDashboard(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(dashboard)
}

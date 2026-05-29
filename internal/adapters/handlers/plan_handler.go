package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/cashflow/admin-service/internal/core/domain"
	"github.com/cashflow/admin-service/internal/core/ports"
	"github.com/go-chi/chi/v5"
)

type PlanHandler struct {
	svc ports.PricingPlanService
}

func NewPlanHandler(svc ports.PricingPlanService) *PlanHandler {
	return &PlanHandler{svc: svc}
}

type ErrorResponse struct {
	Status bool   `json:"status"`
	Error  string `json:"error"`
}

func (h *PlanHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req domain.CreatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid request body"})
		return
	}

	plan, err := h.svc.CreatePlan(r.Context(), &req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(struct {
		Status bool                `json:"status"`
		Data   *domain.PricingPlan `json:"data"`
	}{Status: true, Data: plan})
}

func (h *PlanHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	plans, err := h.svc.ListPlans(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: err.Error()})
		return
	}
	if plans == nil {
		plans = []*domain.PricingPlan{}
	}

	json.NewEncoder(w).Encode(struct {
		Status bool                  `json:"status"`
		Data   []*domain.PricingPlan `json:"data"`
	}{Status: true, Data: plans})
}

func (h *PlanHandler) Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	planID := chi.URLParam(r, "id")
	if planID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Plan ID is required"})
		return
	}

	plan, err := h.svc.GetPlan(r.Context(), planID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Plan not found"})
		return
	}

	json.NewEncoder(w).Encode(struct {
		Status bool                `json:"status"`
		Data   *domain.PricingPlan `json:"data"`
	}{Status: true, Data: plan})
}

func (h *PlanHandler) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	planID := chi.URLParam(r, "id")
	if planID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Plan ID is required"})
		return
	}

	var req domain.UpdatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid request body"})
		return
	}

	plan, err := h.svc.UpdatePlan(r.Context(), planID, &req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(struct {
		Status bool                `json:"status"`
		Data   *domain.PricingPlan `json:"data"`
	}{Status: true, Data: plan})
}

func (h *PlanHandler) ToggleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	planID := chi.URLParam(r, "id")
	if planID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Plan ID is required"})
		return
	}

	var body struct {
		IsActive bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid request body"})
		return
	}

	err := h.svc.TogglePlanStatus(r.Context(), planID, body.IsActive)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(struct {
		Status bool `json:"status"`
	}{Status: true})
}

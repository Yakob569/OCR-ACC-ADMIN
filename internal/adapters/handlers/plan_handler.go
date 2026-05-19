package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/cashflow/admin-service/internal/core/domain"
	"github.com/cashflow/admin-service/internal/core/ports"
	"github.com/google/uuid"
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

func (h *PlanHandler) HandlePlans(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodGet {
		h.List(w, r)
		return
	}

	if r.Method == http.MethodPost {
		h.Create(w, r)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
	json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Method not allowed"})
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

	json.NewEncoder(w).Encode(struct {
		Status bool                  `json:"status"`
		Data   []*domain.PricingPlan `json:"data"`
	}{Status: true, Data: plans})
}

func (h *PlanHandler) HandlePlan(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if strings.HasSuffix(r.URL.Path, "/status") {
		h.ToggleStatus(w, r)
		return
	}

	if r.Method == http.MethodGet {
		h.Get(w, r)
		return
	}

	if r.Method == http.MethodPut {
		h.Update(w, r)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
	json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Method not allowed"})
}

func (h *PlanHandler) Get(w http.ResponseWriter, r *http.Request) {
	planID, err := planIDFromPath(r.URL.Path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid plan ID"})
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
	planID, err := planIDFromPath(r.URL.Path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid plan ID"})
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
	if r.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Only PATCH is allowed"})
		return
	}

	planID, err := planIDFromStatusPath(r.URL.Path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid plan ID"})
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

	err = h.svc.TogglePlanStatus(r.Context(), planID, body.IsActive)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(struct {
		Status bool `json:"status"`
	}{Status: true})
}

func planIDFromPath(path string) (uuid.UUID, error) {
	idPart := strings.TrimPrefix(path, "/api/v1/admin/plans/")
	return uuid.Parse(strings.TrimSpace(idPart))
}

func planIDFromStatusPath(path string) (uuid.UUID, error) {
	trimmed := strings.Trim(path, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 6 || parts[5] != "status" {
		return uuid.Nil, fmt.Errorf("invalid status path")
	}
	return uuid.Parse(parts[4])
}

package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/cashflow/admin-service/internal/core/domain"
	"github.com/cashflow/admin-service/internal/core/ports"
	"github.com/google/uuid"
)

type SubscriptionHandler struct {
	subSvc ports.SubscriptionService
}

func NewSubscriptionHandler(subSvc ports.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{subSvc: subSvc}
}

func (h *SubscriptionHandler) HandleRequests(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	requests, err := h.subSvc.ListSubscriptionRequests(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(requests)
}

func (h *SubscriptionHandler) HandleRequestActions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	// URL format: /api/v1/admin/subscriptions/requests/{id}/{action}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 7 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid URL path"})
		return
	}

	// parts should be: ["api", "v1", "admin", "subscriptions", "requests", "{id}", "{action}"]
	idStr := parts[5]
	action := parts[6]

	id, err := uuid.Parse(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid subscription request ID"})
		return
	}

	switch action {
	case "approve":
		err = h.subSvc.ApproveSubscriptionRequest(r.Context(), id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"status": true, "message": "Request approved successfully"})

	case "reject":
		var req domain.RejectSubscriptionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}

		err = h.subSvc.RejectSubscriptionRequest(r.Context(), id, req.RejectionReason)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"status": true, "message": "Request rejected successfully"})

	default:
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid action. Must be 'approve' or 'reject'"})
	}
}

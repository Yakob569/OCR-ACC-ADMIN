package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/cashflow/admin-service/internal/core/domain"
	"github.com/cashflow/admin-service/internal/core/ports"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type PaymentMethodHandler struct {
	svc ports.PaymentMethodService
}

func NewPaymentMethodHandler(svc ports.PaymentMethodService) *PaymentMethodHandler {
	return &PaymentMethodHandler{svc: svc}
}

func (h *PaymentMethodHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Name          string `json:"name"`
		ImageURL      string `json:"image_url"`
		AccountNumber string `json:"account_number"`
		Status        string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid request body"})
		return
	}

	method, err := h.svc.CreatePaymentMethod(r.Context(), req.Name, req.ImageURL, req.AccountNumber, req.Status)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(struct {
		Status bool                  `json:"status"`
		Data   *domain.PaymentMethod `json:"data"`
	}{Status: true, Data: method})
}

func (h *PaymentMethodHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	methods, err := h.svc.ListPaymentMethods(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(struct {
		Status bool                    `json:"status"`
		Data   []*domain.PaymentMethod `json:"data"`
	}{Status: true, Data: methods})
}

func (h *PaymentMethodHandler) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid payment method ID"})
		return
	}

	var req struct {
		Name          string `json:"name"`
		ImageURL      string `json:"image_url"`
		AccountNumber string `json:"account_number"`
		Status        string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid request body"})
		return
	}

	method, err := h.svc.UpdatePaymentMethod(r.Context(), id, req.Name, req.ImageURL, req.AccountNumber, req.Status)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(struct {
		Status bool                  `json:"status"`
		Data   *domain.PaymentMethod `json:"data"`
	}{Status: true, Data: method})
}

func (h *PaymentMethodHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: "Invalid payment method ID"})
		return
	}

	err = h.svc.DeletePaymentMethod(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Status: false, Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(struct {
		Status bool `json:"status"`
	}{Status: true})
}

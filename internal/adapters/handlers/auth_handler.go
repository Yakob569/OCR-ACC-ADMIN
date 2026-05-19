package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/cashflow/admin-service/internal/core/ports"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authSvc       ports.AuthService
	adminUsername string
	adminPassword string
}

func NewAuthHandler(authSvc ports.AuthService, adminUsername, adminPassword string) *AuthHandler {
	return &AuthHandler{
		authSvc:       authSvc,
		adminUsername: adminUsername,
		adminPassword: adminPassword,
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Status       bool   `json:"status"`
	Token        string `json:"token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Error        string `json:"error,omitempty"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(LoginResponse{Status: false, Error: "Method not allowed"})
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LoginResponse{Status: false, Error: "Invalid request body"})
		return
	}

	if req.Username != h.adminUsername || req.Password != h.adminPassword {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(LoginResponse{Status: false, Error: "Invalid credentials"})
		return
	}

	// Generate a deterministic UUID for the admin user
	adminUUID := uuid.NewMD5(uuid.NameSpaceDNS, []byte("admin-service-user"))

	accessToken, refreshToken, err := h.authSvc.GenerateTokenPair(adminUUID, "admin")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{Status: false, Error: "Failed to generate token"})
		return
	}

	json.NewEncoder(w).Encode(LoginResponse{
		Status:       true,
		Token:        accessToken,
		RefreshToken: refreshToken,
	})
}

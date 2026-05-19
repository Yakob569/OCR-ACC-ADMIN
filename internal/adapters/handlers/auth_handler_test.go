package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

type mockAuthService struct {
	token string
	err   error
}

func (m *mockAuthService) ValidateToken(token string) (uuid.UUID, string, error) {
	return uuid.Nil, "", nil
}

func (m *mockAuthService) GenerateTokenPair(userID uuid.UUID, role string) (string, string, error) {
	return m.token, "mock-refresh-token", m.err
}

func TestLoginHandler(t *testing.T) {
	mockSvc := &mockAuthService{token: "mock-jwt-token"}
	handler := NewAuthHandler(mockSvc, "admin", "admin123")

	body, _ := json.Marshal(LoginRequest{
		Username: "admin",
		Password: "admin123",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/auth/login", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rec.Code)
	}

	var resp LoginResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Status {
		t.Error("expected status to be true")
	}

	if resp.Token != "mock-jwt-token" {
		t.Errorf("expected token 'mock-jwt-token', got '%s'", resp.Token)
	}

	if resp.RefreshToken != "mock-refresh-token" {
		t.Errorf("expected refresh token 'mock-refresh-token', got '%s'", resp.RefreshToken)
	}
}

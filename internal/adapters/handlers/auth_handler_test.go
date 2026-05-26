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

func (m *mockAuthService) ValidateRefreshToken(token string) (uuid.UUID, string, error) {
	return uuid.Nil, "", m.err
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

func TestRefreshTokenHandler(t *testing.T) {
	mockSvc := &mockAuthService{token: "new-mock-jwt-token"}
	handler := NewAuthHandler(mockSvc, "admin", "admin123")

	body, _ := json.Marshal(RefreshRequest{
		RefreshToken: "valid-refresh-token",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/auth/refresh", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler.RefreshToken(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rec.Code)
	}

	var resp RefreshResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Status {
		t.Error("expected status to be true")
	}

	if resp.Token != "new-mock-jwt-token" {
		t.Errorf("expected token 'new-mock-jwt-token', got '%s'", resp.Token)
	}

	if resp.RefreshToken != "mock-refresh-token" {
		t.Errorf("expected refresh token 'mock-refresh-token', got '%s'", resp.RefreshToken)
	}
}

func TestLogoutHandler(t *testing.T) {
	mockSvc := &mockAuthService{}
	handler := NewAuthHandler(mockSvc, "admin", "admin123")

	body, _ := json.Marshal(LogoutRequest{
		RefreshToken: "valid-refresh-token",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/auth/logout", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	handler.Logout(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rec.Code)
	}

	var resp LogoutResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Status {
		t.Error("expected status to be true")
	}

	if resp.Message != "Logged out successfully" {
		t.Errorf("expected message 'Logged out successfully', got '%s'", resp.Message)
	}
}

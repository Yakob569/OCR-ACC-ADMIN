package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/cashflow/admin-service/internal/core/ports"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type jwtAuthAdapter struct {
	secretKey []byte
}

func NewJWTAuthAdapter(secret string) ports.AuthService {
	return &jwtAuthAdapter{
		secretKey: []byte(secret),
	}
}

func (a *jwtAuthAdapter) validateTokenWithType(tokenStr string, expectedType string) (uuid.UUID, string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}
		return a.secretKey, nil
	})

	if err != nil || !token.Valid {
		return uuid.Nil, "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, "", errors.New("invalid claims")
	}

	tokenType, ok := claims["token_type"].(string)
	if !ok || tokenType != expectedType {
		return uuid.Nil, "", errors.New("invalid token type")
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, "", errors.New("sub claim missing")
	}

	userID, err := uuid.Parse(sub)
	if err != nil {
		return uuid.Nil, "", err
	}

	role, ok := claims["role"].(string)
	if !ok {
		return uuid.Nil, "", errors.New("role claim missing")
	}

	return userID, role, nil
}

func (a *jwtAuthAdapter) ValidateToken(tokenStr string) (uuid.UUID, string, error) {
	return a.validateTokenWithType(tokenStr, "access")
}

func (a *jwtAuthAdapter) ValidateRefreshToken(tokenStr string) (uuid.UUID, string, error) {
	return a.validateTokenWithType(tokenStr, "refresh")
}

func (a *jwtAuthAdapter) GenerateTokenPair(userID uuid.UUID, role string) (string, string, error) {
	now := time.Now()
	
	accessClaims := jwt.MapClaims{
		"sub":        userID.String(),
		"role":       role,
		"token_type": "access",
		"exp":        now.Add(24 * time.Hour).Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString(a.secretKey)
	if err != nil {
		return "", "", err
	}

	refreshClaims := jwt.MapClaims{
		"sub":        userID.String(),
		"role":       role,
		"token_type": "refresh",
		"exp":        now.Add(7 * 24 * time.Hour).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString(a.secretKey)
	if err != nil {
		return "", "", err
	}

	return accessStr, refreshStr, nil
}

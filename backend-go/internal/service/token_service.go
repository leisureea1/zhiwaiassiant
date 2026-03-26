package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

type Claims struct {
	UserID    string    `json:"uid"`
	Role      string    `json:"role"`
	TokenType TokenType `json:"tokenType"`
	jwt.RegisteredClaims
}

type TokenService struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewTokenService(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) *TokenService {
	return &TokenService{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

func (s *TokenService) GenerateAccessToken(userID, role string) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.accessTTL)
	claims := Claims{
		UserID:    userID,
		Role:      role,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.accessSecret)
	return signed, expiresAt, err
}

func (s *TokenService) GenerateRefreshToken(userID, role string) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.refreshTTL)
	claims := Claims{
		UserID:    userID,
		Role:      role,
		TokenType: TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.refreshSecret)
	return signed, expiresAt, err
}

func (s *TokenService) ParseAccessToken(tokenString string) (*Claims, error) {
	return s.parse(tokenString, s.accessSecret, TokenTypeAccess)
}

func (s *TokenService) ParseRefreshToken(tokenString string) (*Claims, error) {
	return s.parse(tokenString, s.refreshSecret, TokenTypeRefresh)
}

func (s *TokenService) parse(tokenString string, secret []byte, expect TokenType) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	if claims.TokenType != expect {
		return nil, fmt.Errorf("invalid token type")
	}

	return claims, nil
}

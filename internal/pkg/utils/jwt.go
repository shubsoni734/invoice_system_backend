package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secret []byte
	expiry time.Duration
}

type Claims struct {
	UserID               string `json:"user_id,omitempty"`
	OrgID                string `json:"org_id,omitempty"`
	Role                 string `json:"role,omitempty"`
	SuperAdminID         string `json:"super_admin_id,omitempty"`
	ImpersonatedBy       string `json:"impersonated_by,omitempty"`
	ImpersonationSession string `json:"impersonation_session_id,omitempty"`
	jwt.RegisteredClaims
}

func NewJWTManager(secret string, expiry time.Duration) (*JWTManager, error) {
	if secret == "" {
		return nil, fmt.Errorf("JWT secret cannot be empty")
	}
	return &JWTManager{secret: []byte(secret), expiry: expiry}, nil
}

func (m *JWTManager) GenerateToken(claims *Claims) (string, error) {
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(m.expiry))
	claims.IssuedAt = jwt.NewNumericDate(time.Now())
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *JWTManager) GenerateRefreshToken(claims *Claims) (string, error) {
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour))
	claims.IssuedAt = jwt.NewNumericDate(time.Now())
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *JWTManager) VerifyToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}

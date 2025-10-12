package crypto

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenService struct {
	secret string
}

type Claims struct {
	UserID int64 `json:"userId"`
	jwt.RegisteredClaims
}

func NewTokenService(secret string) *TokenService {
	return &TokenService{secret: secret}
}

// GenerateToken generates a JWT token for a user
func (s *TokenService) GenerateToken(userID int64) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(720 * time.Hour)), // 30 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secret))
}

// VerifyToken verifies a JWT token and returns the user ID
func (s *TokenService) VerifyToken(tokenString string) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secret), nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, nil
	}

	return 0, fmt.Errorf("invalid token")
}
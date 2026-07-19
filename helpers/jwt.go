package helpers

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"finalproject/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const UserDataKey = "userData"

type Claims struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(id uint, username string, email string, role string) (string, error) {
	cfg := config.Load()
	return GenerateTokenWithDuration(id, username, email, role, time.Duration(cfg.JWTExpirationHours)*time.Hour)
}

func GenerateTokenWithDuration(id uint, username string, email string, role string, ttl time.Duration) (string, error) {
	cfg := config.Load()
	if strings.TrimSpace(cfg.JWTSecret) == "" {
		return "", errors.New("jwt secret is not configured")
	}

	now := time.Now().UTC()
	claims := Claims{
		ID:       id,
		Username: username,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

func VerifyToken(c *gin.Context) (*Claims, error) {
	headerToken := c.Request.Header.Get("Authorization")
	fields := strings.Fields(headerToken)
	if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") {
		return nil, errors.New("sign in to proceed")
	}

	return VerifyTokenString(fields[1])
}

func VerifyTokenString(tokenString string) (*Claims, error) {
	cfg := config.Load()
	if strings.TrimSpace(cfg.JWTSecret) == "" {
		return nil, errors.New("jwt secret is not configured")
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, fmt.Errorf("unexpected signing method: %s", token.Header["alg"])
			}

			return []byte(cfg.JWTSecret), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return nil, errors.New("invalid or expired token")
	}

	if !token.Valid || claims.ID == 0 {
		return nil, errors.New("invalid or expired token")
	}

	return claims, nil
}

func GetUserClaims(c *gin.Context) (*Claims, bool) {
	value, exists := c.Get(UserDataKey)
	if !exists {
		return nil, false
	}

	claims, ok := value.(*Claims)
	return claims, ok
}

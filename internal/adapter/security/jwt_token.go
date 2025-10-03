package security

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hutamy/go-invoice-backend/config"
	"github.com/hutamy/go-invoice-backend/internal/domain/ports"
)

type JWTTokenService struct{}

func NewJWTTokenService() ports.TokenService {
	return &JWTTokenService{}
}

func (JWTTokenService) Generate(userID uint, ttl time.Duration) (string, error) {
	secret := []byte(config.GetConfig().JwtSecret)
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(ttl).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func (JWTTokenService) Parse(token string) (map[string]any, error) {
	secret := []byte(config.GetConfig().JwtSecret)
	t, err := jwt.Parse(token, func(tok *jwt.Token) (interface{}, error) {
		if _, ok := tok.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenUnverifiable
		}
		return secret, nil
	})
	if err != nil || !t.Valid {
		return nil, err
	}

	return t.Claims.(jwt.MapClaims), nil
}

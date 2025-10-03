package ports

import "time"

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) bool
}

type TokenService interface {
	Generate(userID uint, ttl time.Duration) (string, error)
	Parse(token string) (map[string]any, error)
}

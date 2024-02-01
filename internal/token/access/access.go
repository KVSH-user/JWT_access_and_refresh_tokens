package access

import (
	"github.com/dgrijalva/jwt-go"
	"time"
)

func Generate(secretKey string, guid string, ttl time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(ttl).Unix(),
		Subject:   guid,
	})

	return token.SignedString([]byte(secretKey))
}

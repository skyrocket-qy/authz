package pkg

import (
	"strconv"
	"time"

	"authz/internal/cfg"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/skyrocket-qy/erx"
)

func GenerateJWT(userID uint) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatUint(uint64(userID), 10),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(cfg.Cfg.Jwt.Secret))
	if err != nil {
		return "", erx.W(err)
	}

	return tokenString, nil
}

func GenRefreshToken() string {
	return uuid.New().String()
}

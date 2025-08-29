package util_test

import (
	"strconv"
	"testing"

	"authz/internal/config"
	"authz/internal/util"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGenRefreshToken(t *testing.T) {
	t.Run("should generate a valid UUID", func(t *testing.T) {
		token1 := util.GenRefreshToken()
		_, err := uuid.Parse(token1)
		assert.NoError(t, err)
	})

	t.Run("should generate different tokens on subsequent calls", func(t *testing.T) {
		token1 := util.GenRefreshToken()
		token2 := util.GenRefreshToken()
		assert.NotEqual(t, token1, token2)
	})
}

func TestNewJwtToken(t *testing.T) {
	// Stash and restore original config
	originalSecret := config.Conf.Jwt.Secret

	defer func() {
		config.Conf.Jwt.Secret = originalSecret
	}()

	t.Run("should generate a valid JWT with correct claims", func(t *testing.T) {
		// Arrange
		config.Conf.Jwt.Secret = "my-super-secret-key"

		var userID uint = 123

		// Act
		tokenString, err := util.NewJwtToken(userID)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, tokenString)

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			// Check signing method
			assert.Equal(t, jwt.SigningMethodHS256, token.Method)

			return []byte(config.Conf.Jwt.Secret), nil
		})

		assert.NoError(t, err)
		assert.True(t, token.Valid, "token should be valid")

		// Check claims
		claims, ok := token.Claims.(jwt.MapClaims)
		assert.True(t, ok, "claims should be of type jwt.MapClaims")

		subject, err := claims.GetSubject()
		assert.NoError(t, err)
		assert.Equal(t, strconv.FormatUint(uint64(userID), 10), subject)
	})

	t.Run("should return an error if JWT secret is empty", func(t *testing.T) {
		// Arrange
		config.Conf.Jwt.Secret = ""

		var userID uint = 456

		// Act
		_, err := util.NewJwtToken(userID)

		// Assert
		assert.Error(t, err)
	})
}

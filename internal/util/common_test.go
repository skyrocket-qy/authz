package util_test

import (
	"context"
	"regexp"
	"testing"

	"authz/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestGetUserId(t *testing.T) {
	t.Run("should return user ID when it exists in context", func(t *testing.T) {
		// Arrange
		expectedUserID := uint(123)
		ctx := context.WithValue(context.Background(), "userID", expectedUserID)

		// Act
		actualUserID := util.GetUserId(ctx)

		// Assert
		assert.Equal(t, expectedUserID, actualUserID)
	})

	t.Run("should return 0 when user ID does not exist in context", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		// Act
		actualUserID := util.GetUserId(ctx)

		// Assert
		assert.Equal(t, uint(0), actualUserID)
	})

	t.Run("should return 0 when user ID is of wrong type", func(t *testing.T) {
		// Arrange
		ctx := context.WithValue(context.Background(), "userID", "not-a-uint")

		// Act
		actualUserID := util.GetUserId(ctx)

		// Assert
		assert.Equal(t, uint(0), actualUserID)
	})
}

func TestGenNumCode(t *testing.T) {
	t.Run("should generate a numeric code of specified length", func(t *testing.T) {
		// Arrange
		length := 6

		// Act
		code, err := util.GenNumCode(length)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, code, length)
		match, _ := regexp.MatchString(`^[0-9]+$`, code)
		assert.True(t, match, "code should be numeric")
	})

	t.Run("should return empty string for zero length", func(t *testing.T) {
		// Arrange
		length := 0

		// Act
		code, err := util.GenNumCode(length)

		// Assert
		assert.NoError(t, err)
		assert.Empty(t, code)
	})

	t.Run("should generate different codes on subsequent calls", func(t *testing.T) {
		// Act
		code1, err1 := util.GenNumCode(8)
		code2, err2 := util.GenNumCode(8)

		// Assert
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEqual(t, code1, code2)
	})
}

func TestStr(t *testing.T) {
	t.Run("should return a pointer to the given string", func(t *testing.T) {
		// Arrange
		s := "hello"

		// Act
		sp := util.Str(s)

		// Assert
		assert.NotNil(t, sp)
		assert.Equal(t, s, *sp)
	})

	t.Run("should return a pointer to an empty string", func(t *testing.T) {
		// Arrange
		s := ""

		// Act
		sp := util.Str(s)

		// Assert
		assert.NotNil(t, sp)
		assert.Equal(t, s, *sp)
	})
}

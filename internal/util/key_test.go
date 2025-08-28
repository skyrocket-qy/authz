package util

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestWithDB(t *testing.T) {
	t.Run("should add a DB instance to the context", func(t *testing.T) {
		// Arrange
		parentCtx := context.Background()
		// Use a non-nil pointer to a zero-value struct.
		dummyDB := &gorm.DB{}

		// Act
		childCtx := WithDB(parentCtx, dummyDB)

		// Assert
		// Check that the child context is different from the parent
		assert.NotEqual(t, parentCtx, childCtx)

		// Retrieve the value from the child context
		retrievedValue := childCtx.Value(DbCtxKey{})
		assert.NotNil(t, retrievedValue, "value should be present in child context")

		// Type assert and check for equality
		retrievedDB, ok := retrievedValue.(*gorm.DB)
		assert.True(t, ok, "value should be of type *gorm.DB")
		assert.Equal(t, dummyDB, retrievedDB, "retrieved DB should be the same as the one passed in")
	})

	t.Run("should not affect the parent context", func(t *testing.T) {
		// Arrange
		parentCtx := context.Background()
		var dummyDB *gorm.DB

		// Act
		_ = WithDB(parentCtx, dummyDB)

		// Assert
		retrievedValue := parentCtx.Value(DbCtxKey{})
		assert.Nil(t, retrievedValue, "parent context should not be modified")
	})
}

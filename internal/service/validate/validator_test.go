package validate_test

import (
	"testing"

	"authz/internal/service/validate"
	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
}

func TestValidator(t *testing.T) {
	// Initialize the validator
	validate.New()

	t.Run("Valid struct", func(t *testing.T) {
		s := TestStruct{
			Name:  "Test User",
			Email: "test@example.com",
		}
		err := validate.Struct(s)
		assert.NoError(t, err)
	})

	t.Run("Invalid struct - missing required field", func(t *testing.T) {
		s := TestStruct{
			Email: "test@example.com",
		}
		err := validate.Struct(s)
		assert.Error(t, err)
	})

	t.Run("Invalid struct - invalid email", func(t *testing.T) {
		s := TestStruct{
			Name:  "Test User",
			Email: "not-an-email",
		}
		err := validate.Struct(s)
		assert.Error(t, err)
	})
}

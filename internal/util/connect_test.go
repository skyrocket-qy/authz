package util

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
)

func TestNewApiErr(t *testing.T) {
	t.Run("should handle standard error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		stdErr := errors.New("a standard error")

		// Act
		apiErr := NewApiErr(ctx, stdErr)

		// Assert
		assert.Equal(t, connect.CodeUnknown, apiErr.Code())
		assert.Equal(t, stdErr, apiErr.Unwrap())
	})

	// NOTE: Testing the erx.CtxErr path of NewApiErr is not currently possible
	// in this package's unit tests. To create a valid *erx.CtxErr, this test
	// would need to import the `errcode` package. However, `errcode` is in
	// `internal/service`, and this `util` package is a dependency of `service`,
	// so importing it would create a circular dependency.
	//
	// The `erx` library itself does not seem to expose a public constructor
	// for creating `erx.Code` instances from strings or integers, which is
	// necessary to create a test-only `*erx.CtxErr`.
	//
	// This function is tested implicitly by the integration and e2e tests
	// that do generate real application errors.
}

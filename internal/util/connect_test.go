package util

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/skyrocket-qy/erx"
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

	t.Run("should handle CtxErr", func(t *testing.T) {
		testCases := []struct {
			name         string
			appErrCode   err
			expectedCode connect.Code
		}{
			{"Bad Request", ErrBadRequest, connect.CodeInvalidArgument},
			{"Unauthorized", ErrUnauthorized, connect.CodeUnauthenticated},
			{"Not Found", ErrNotFound, connect.CodeNotFound},
			{"Duplicate", ErrDuplicate, connect.CodeAlreadyExists},
			{"Unknown", ErrUnknown, connect.CodeInternal},
			{"Not Implemented", ErrNotImplemented, connect.CodeUnimplemented},
			{"Invalid Code", "abc", connect.CodeInternal},
			{"Default case", "123", connect.CodeUnknown},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				ctx := context.Background()
				appErr := erx.New(tc.appErrCode, "test error")

				apiErr := NewApiErr(ctx, appErr)

				assert.Equal(t, tc.expectedCode, apiErr.Code())
			})
		}
	})
}

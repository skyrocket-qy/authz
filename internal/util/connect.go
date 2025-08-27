package util

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"connectrpc.com/connect"
	"github.com/rs/zerolog/log"
	"github.com/skyrocket-qy/erx"
	pkgpbv1 "github.com/skyrocket-qy/protos/gen/pkgpb/v1"
	"go.opentelemetry.io/otel/trace"
)

func NewApiErr(ctx context.Context, err error) *connect.Error {
	var appErr *erx.CtxErr
	if ok := errors.As(err, &appErr); !ok {
		return connect.NewError(connect.CodeUnknown, err)
	}

	span := trace.SpanFromContext(ctx)
	logError(span.SpanContext().TraceID().String(), appErr)

	apiCode := appToApiCode(appErr.Code)
	cErr := connect.NewError(apiCode, appErr.Unwrap())

	detail, _ := connect.NewErrorDetail(&pkgpbv1.CtxErr{
		Code:    appErr.Code.Str(),
		TraceId: span.SpanContext().TraceID().String(),
	})
	cErr.AddDetail(detail)

	return cErr
}

func logError(traceId string, err *erx.CtxErr) {
	e := log.Error().Str("traceId", traceId)
	if cause := err.Unwrap().Error(); cause != "" {
		e.Str("cause", cause)
	}

	e.Str("code", err.Code.Str())

	// Convert callerInfos to pretty strings
	filtered := filterCallerInfos(err.CallerInfos)
	trace := make([]string, 0, len(filtered))
	for _, ci := range filtered {
		trace = append(trace, fmt.Sprintf("%s %d %s",
			trimToProject(ci.File),
			ci.Line,
			extractFuncName(ci.Function),
		))
	}
	e.Strs("callerTrace", trace)
	e.Msg("error")
}

func appToApiCode(code erx.Code) connect.Code {
	parts := strings.SplitN(code.Str(), ".", 2)
	if len(parts) == 0 {
		return connect.CodeInternal // fallback
	}

	intCode, err := strconv.Atoi(parts[0])
	if err != nil {
		log.Err(err).Msg("appToApiCode error")
		return connect.CodeInternal
	}

	switch intCode {
	case 400:
		return connect.CodeInvalidArgument
	case 401:
		return connect.CodeUnauthenticated
	case 404:
		return connect.CodeNotFound
	case 409:
		return connect.CodeAlreadyExists
	case 500:
		return connect.CodeInternal
	case 501:
		return connect.CodeUnimplemented
	default:
		return connect.CodeUnknown
	}
}

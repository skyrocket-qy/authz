package util

import (
	"errors"

	"github.com/rs/zerolog/log"
	"github.com/skyrocket-qy/erx"
)

func LogE(err error) {
	var cErr *erx.CtxErr
	if ok := errors.As(err, &cErr); !ok {
		return
	}

	log.Error().Msg(cErr.Unwrap().Error())
}

package validate

import (
	"authz/internal/util"
	"github.com/go-playground/validator/v10"
	"github.com/skyrocket-qy/erx"
	"github.com/skyrocket-qy/gox/logx"
)

var v *validator.Validate

func New() {
	logx.Info("InitValidator")

	v = validator.New(validator.WithRequiredStructEnabled())
}

func Struct(st any) error {
	if err := v.Struct(st); err != nil {
		return erx.W(err).SetCode(util.ErrValidateInput)
	}

	return nil
}

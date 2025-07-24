package repo

import (
	"authz/internal/entity/model"
	"context"

	authzpbv1 "github.com/skyrocket-qy/protos/gen/authzpb/v1"
)

type TupleRepo interface {
	Find(c context.Context, filter *authzpbv1.Tuple, exact bool) ([]*authzpbv1.Tuple, error)
}

var _ TupleRepo = (*TupleRepoImpl)(nil)

type TupleRepoImpl struct {
}

func NewTupleRepo() *TupleRepoImpl {
	return &TupleRepoImpl{}
}

func (r *TupleRepoImpl) convertTuple(t *authzpbv1.Tuple) *model.Tuple {

	return &model.Tuple{
		Sbj:      t.Sbj,
		Relation: t.Relation,
		Obj:      t.Obj,
	}
}

package repo

type TupleRepo interface {
}

var _ TupleRepo = (*TupleRepoImpl)(nil)

type TupleRepoImpl struct {
}

func NewTupleRepo() *TupleRepoImpl {
	return &TupleRepoImpl{}
}

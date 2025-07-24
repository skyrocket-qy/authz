package logic

type GraphLogic interface {
}

var _ GraphLogic = (*GraphLogicImpl)(nil)

type GraphLogicImpl struct {
}

func NewTupleRepo() *GraphLogicImpl {
	return &GraphLogicImpl{}
}

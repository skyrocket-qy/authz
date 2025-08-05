package connect

import (
	"authz/internal/logic"
	"context"

	"connectrpc.com/connect"
	authzpbv1 "github.com/skyrocket-qy/protos/gen/authzpb/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Handler struct {
	zLogic    logic.ZanzibarLogic
	rbacLogic logic.RbacLogic
}

func NewHandler(zLogic logic.ZanzibarLogic, rbacLogic logic.RbacLogic) *Handler {
	return &Handler{
		zLogic:    zLogic,
		rbacLogic: rbacLogic,
	}
}

func (h *Handler) ListTuples(
	ctx context.Context, req *connect.Request[authzpbv1.ListTuplesIn]) (
	*connect.Response[authzpbv1.ListTuplesOut], error,
) {
	resp, err := h.zLogic.List(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), err
}

func (h *Handler) CreateTuple(
	ctx context.Context, req *connect.Request[authzpbv1.Tuple]) (
	*connect.Response[emptypb.Empty], error,
) {
	err := h.zLogic.Create(ctx, req.Msg)
	return nil, err
}

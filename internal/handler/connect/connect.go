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

func (h *Handler) ListUsers(
	ctx context.Context, req *connect.Request[authzpbv1.ListUsersIn],
) (*connect.Response[authzpbv1.ListUsersOut], error) {
	out, err := h.rbacLogic.ListUsers(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(out), nil
}

func (h *Handler) UpdateUser(
	ctx context.Context, req *connect.Request[authzpbv1.UpdateUserIn],
) (*connect.Response[emptypb.Empty], error) {
	err := h.rbacLogic.UpdateUser(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *Handler) DeleteUser(
	ctx context.Context, req *connect.Request[authzpbv1.DeleteUserIn],
) (*connect.Response[emptypb.Empty], error) {
	err := h.rbacLogic.DeleteUser(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *Handler) CreateRole(
	ctx context.Context, req *connect.Request[authzpbv1.CreateRoleIn],
) (*connect.Response[emptypb.Empty], error) {
	err := h.rbacLogic.CreateRole(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *Handler) ListRoles(
	ctx context.Context, req *connect.Request[authzpbv1.ListRolesIn],
) (*connect.Response[authzpbv1.ListRolesOut], error) {
	out, err := h.rbacLogic.ListRoles(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(out), nil
}

func (h *Handler) UpdateRole(
	ctx context.Context, req *connect.Request[authzpbv1.UpdateRoleIn],
) (*connect.Response[emptypb.Empty], error) {
	err := h.rbacLogic.UpdateRole(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *Handler) DeleteRole(
	ctx context.Context, req *connect.Request[authzpbv1.DeleteRoleIn],
) (*connect.Response[emptypb.Empty], error) {
	err := h.rbacLogic.DeleteRole(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *Handler) CreateResource(
	ctx context.Context, req *connect.Request[authzpbv1.CreateResourceIn],
) (*connect.Response[emptypb.Empty], error) {
	err := h.rbacLogic.CreateResource(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *Handler) ListResources(
	ctx context.Context, req *connect.Request[authzpbv1.ListResourcesIn],
) (*connect.Response[authzpbv1.ListResourcesOut], error) {
	out, err := h.rbacLogic.ListResources(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(out), nil
}

func (h *Handler) DeleteResource(
	ctx context.Context, req *connect.Request[authzpbv1.DeleteResourceIn],
) (*connect.Response[emptypb.Empty], error) {
	err := h.rbacLogic.DeleteResource(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *Handler) AssignRole(
	ctx context.Context, req *connect.Request[authzpbv1.AssignRoleIn],
) (*connect.Response[emptypb.Empty], error) {
	err := h.rbacLogic.AssignRole(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *Handler) RevokeRole(
	ctx context.Context, req *connect.Request[authzpbv1.RevokeRoleIn],
) (*connect.Response[emptypb.Empty], error) {
	err := h.rbacLogic.RevokeRole(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *Handler) GrantPerm(
	ctx context.Context, req *connect.Request[authzpbv1.GrantPermIn],
) (*connect.Response[emptypb.Empty], error) {
	err := h.rbacLogic.GrantPerm(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *Handler) RevokePerm(
	ctx context.Context, req *connect.Request[authzpbv1.RevokePermIn],
) (*connect.Response[emptypb.Empty], error) {
	err := h.rbacLogic.RevokePerm(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *Handler) ListTuples(
	ctx context.Context, req *connect.Request[authzpbv1.ListTuplesIn]) (
	*connect.Response[authzpbv1.ListTuplesOut], error,
) {
	resp, err := h.zLogic.List(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(resp), nil
}

func (h *Handler) CreateTuple(
	ctx context.Context, req *connect.Request[authzpbv1.Tuple],
) (*connect.Response[emptypb.Empty], error) {
	err := h.zLogic.Create(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *Handler) DeleteTuples(
	ctx context.Context, req *connect.Request[authzpbv1.DeleteTuplesIn]) (
	*connect.Response[emptypb.Empty], error,
) {
	err := h.zLogic.Delete(ctx, req.Msg)
	if err == nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *Handler) Check(
	ctx context.Context, req *connect.Request[authzpbv1.CheckIn]) (
	*connect.Response[authzpbv1.CheckOut], error,
) {
	res, err := h.zLogic.Check(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(res), nil
}

package rbac

import (
	"context"

	"connectrpc.com/connect"
	"github.com/skyrocket-qy/protos/gen/authzpb/rbacpb"
	"github.com/skyrocket-qy/protos/gen/authzpb/rbacpb/rbacpbconnect"
	authzpbv1 "github.com/skyrocket-qy/protos/gen/authzpb/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ rbacpbconnect.RbacServiceHandler = (*Handler)(nil)

type Handler struct {
	zLogic    ZanzibarLogic
	rbacLogic RbacLogic
}

func NewHandler(zLogic ZanzibarLogic, rbacLogic RbacLogic) *Handler {
	return &Handler{
		zLogic:    zLogic,
		rbacLogic: rbacLogic,
	}
}

func (h *Handler) ListUsers(
	ctx context.Context, req *connect.Request[rbacpb.ListUsersIn],
) (*connect.Response[rbacpb.ListUsersOut], error) {
	out, err := h.rbacLogic.ListUsers(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(out), nil
}

func (h *Handler) UpdateUser(ctx context.Context, req *connect.Request[rbacpb.UpdateUserIn],
) (*connect.Response[emptypb.Empty], error,
) {
	err := h.rbacLogic.UpdateUser(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *Handler) DeleteUser(
	ctx context.Context, req *connect.Request[rbacpb.DeleteUserIn],
) (*connect.Response[emptypb.Empty], error) {
	err := h.rbacLogic.DeleteUser(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *Handler) CreateRole(
	ctx context.Context, req *connect.Request[rbacpb.CreateRoleIn],
) (*connect.Response[emptypb.Empty], error) {
	err := h.rbacLogic.CreateRole(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *Handler) ListRoles(
	ctx context.Context, req *connect.Request[rbacpb.ListRolesIn],
) (*connect.Response[rbacpb.ListRolesOut], error) {
	out, err := h.rbacLogic.ListRoles(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(out), nil
}

func (h *Handler) UpdateRole(
	ctx context.Context, req *connect.Request[rbacpb.UpdateRoleIn],
) (*connect.Response[emptypb.Empty], error) {
	err := h.rbacLogic.UpdateRole(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *Handler) DeleteRole(
	ctx context.Context, req *connect.Request[rbacpb.DeleteRoleIn],
) (*connect.Response[emptypb.Empty], error) {
	err := h.rbacLogic.DeleteRole(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *Handler) CreateResource(
	ctx context.Context, req *connect.Request[rbacpb.CreateResourceIn],
) (*connect.Response[emptypb.Empty], error) {
	err := h.rbacLogic.CreateResource(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *Handler) ListResources(
	ctx context.Context, req *connect.Request[rbacpb.ListResourcesIn],
) (*connect.Response[rbacpb.ListResourcesOut], error) {
	out, err := h.rbacLogic.ListResources(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(out), nil
}

func (h *Handler) DeleteResource(
	ctx context.Context, req *connect.Request[rbacpb.DeleteResourceIn],
) (*connect.Response[emptypb.Empty], error) {
	err := h.rbacLogic.DeleteResource(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *Handler) AssignRole(
	ctx context.Context, req *connect.Request[rbacpb.AssignRoleIn],
) (*connect.Response[emptypb.Empty], error) {
	err := h.rbacLogic.AssignRole(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *Handler) RevokeRole(
	ctx context.Context, req *connect.Request[rbacpb.RevokeRoleIn],
) (*connect.Response[emptypb.Empty], error) {
	err := h.rbacLogic.RevokeRole(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *Handler) GrantPerm(
	ctx context.Context, req *connect.Request[rbacpb.GrantPermIn],
) (*connect.Response[emptypb.Empty], error) {
	err := h.rbacLogic.GrantPerm(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (h *Handler) RevokePerm(
	ctx context.Context, req *connect.Request[rbacpb.RevokePermIn],
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

func (h *Handler) ListResourceType(c context.Context, in *connect.Request[emptypb.Empty]) (
	*connect.Response[rbacpb.ListResourceTypeOut], error,
) {
	res, err := h.rbacLogic.ListResourceTypes(c)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(res), nil
}

func (h *Handler) ListResourcesByType(c context.Context, in *connect.Request[rbacpb.ListResourcesByTypeIn]) (
	*connect.Response[rbacpb.ListResourcesByTypeOut], error,
) {
	res, err := h.rbacLogic.ListResourcesByType(c, in.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(res), nil
}

func (h *Handler) ListPermissionByResource(c context.Context,
	in *connect.Request[rbacpb.ListPermissionByResourceIn]) (
	*connect.Response[rbacpb.ListPermissionByResourceOut], error,
) {
	res, err := h.rbacLogic.ListPermissionByResource(c, in.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(res), nil
}

func (h *Handler) GetRole(c context.Context, in *connect.Request[rbacpb.GetRoleIn]) (
	*connect.Response[rbacpb.GetRoleOut], error,
) {
	res, err := h.rbacLogic.GetRole(c, in.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(res), nil
}

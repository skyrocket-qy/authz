package rbac_test

import (
	"context"
	"testing"

	"authz/internal/engine/rbac"
	"connectrpc.com/connect"
	"github.com/skyrocket-qy/protos/gen/authzpb/rbacpb"
	authzpbv1 "github.com/skyrocket-qy/protos/gen/authzpb/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/emptypb"
)

// MockRbacLogic is a mock of RbacLogic interface.
type MockRbacLogic struct {
	mock.Mock
}

func (m *MockRbacLogic) ListUsers(
	c context.Context,
	in *rbacpb.ListUsersIn,
) (*rbacpb.ListUsersOut, error) {
	args := m.Called(c, in)

	return args.Get(0).(*rbacpb.ListUsersOut), args.Error(1)
}

func (m *MockRbacLogic) UpdateUser(c context.Context, in *rbacpb.UpdateUserIn) error {
	args := m.Called(c, in)

	return args.Error(0)
}

func (m *MockRbacLogic) DeleteUser(c context.Context, in *rbacpb.DeleteUserIn) error {
	args := m.Called(c, in)

	return args.Error(0)
}

func (m *MockRbacLogic) GetRole(
	ctx context.Context,
	in *rbacpb.GetRoleIn,
) (*rbacpb.GetRoleOut, error) {
	args := m.Called(ctx, in)

	return args.Get(0).(*rbacpb.GetRoleOut), args.Error(1)
}

func (m *MockRbacLogic) CreateRole(c context.Context, in *rbacpb.CreateRoleIn) error {
	args := m.Called(c, in)

	return args.Error(0)
}

func (m *MockRbacLogic) ListRoles(
	c context.Context,
	in *rbacpb.ListRolesIn,
) (*rbacpb.ListRolesOut, error) {
	args := m.Called(c, in)

	return args.Get(0).(*rbacpb.ListRolesOut), args.Error(1)
}

func (m *MockRbacLogic) UpdateRole(c context.Context, in *rbacpb.UpdateRoleIn) error {
	args := m.Called(c, in)

	return args.Error(0)
}

func (m *MockRbacLogic) DeleteRole(c context.Context, in *rbacpb.DeleteRoleIn) error {
	args := m.Called(c, in)

	return args.Error(0)
}

func (m *MockRbacLogic) ListResourceTypes(
	ctx context.Context,
) (*rbacpb.ListResourceTypeOut, error) {
	args := m.Called(ctx)

	return args.Get(0).(*rbacpb.ListResourceTypeOut), args.Error(1)
}

func (m *MockRbacLogic) ListResourcesByType(
	ctx context.Context,
	in *rbacpb.ListResourcesByTypeIn,
) (*rbacpb.ListResourcesByTypeOut, error) {
	args := m.Called(ctx, in)

	return args.Get(0).(*rbacpb.ListResourcesByTypeOut), args.Error(1)
}

func (m *MockRbacLogic) ListPermissionByResource(
	ctx context.Context,
	in *rbacpb.ListPermissionByResourceIn,
) (*rbacpb.ListPermissionByResourceOut, error) {
	args := m.Called(ctx, in)

	return args.Get(0).(*rbacpb.ListPermissionByResourceOut), args.Error(1)
}

func (m *MockRbacLogic) CreateResource(c context.Context, in *rbacpb.CreateResourceIn) error {
	args := m.Called(c, in)

	return args.Error(0)
}

func (m *MockRbacLogic) ListResources(
	c context.Context,
	in *rbacpb.ListResourcesIn,
) (*rbacpb.ListResourcesOut, error) {
	args := m.Called(c, in)

	return args.Get(0).(*rbacpb.ListResourcesOut), args.Error(1)
}

func (m *MockRbacLogic) DeleteResource(c context.Context, in *rbacpb.DeleteResourceIn) error {
	args := m.Called(c, in)

	return args.Error(0)
}

func (m *MockRbacLogic) AssignRole(c context.Context, in *rbacpb.AssignRoleIn) error {
	args := m.Called(c, in)

	return args.Error(0)
}

func (m *MockRbacLogic) RevokeRole(c context.Context, in *rbacpb.RevokeRoleIn) error {
	args := m.Called(c, in)

	return args.Error(0)
}

func (m *MockRbacLogic) GrantPerm(c context.Context, in *rbacpb.GrantPermIn) error {
	args := m.Called(c, in)

	return args.Error(0)
}

func (m *MockRbacLogic) RevokePerm(c context.Context, in *rbacpb.RevokePermIn) error {
	args := m.Called(c, in)

	return args.Error(0)
}

// MockZanzibarLogic is a mock of ZanzibarLogic interface.
type MockZanzibarLogic struct {
	mock.Mock
}

func (m *MockZanzibarLogic) Create(c context.Context, tuple *authzpbv1.Tuple) error {
	args := m.Called(c, tuple)

	return args.Error(0)
}

func (m *MockZanzibarLogic) List(
	c context.Context,
	in *authzpbv1.ListTuplesIn,
) (*authzpbv1.ListTuplesOut, error) {
	args := m.Called(c, in)

	return args.Get(0).(*authzpbv1.ListTuplesOut), args.Error(1)
}

func (m *MockZanzibarLogic) Find(
	c context.Context,
	filter *authzpbv1.TupleFilter,
) ([]*authzpbv1.Tuple, error) {
	args := m.Called(c, filter)

	return args.Get(0).([]*authzpbv1.Tuple), args.Error(1)
}

func (m *MockZanzibarLogic) Delete(c context.Context, in *authzpbv1.DeleteTuplesIn) error {
	args := m.Called(c, in)

	return args.Error(0)
}

func (m *MockZanzibarLogic) GetPermissions(
	c context.Context,
	sbj *authzpbv1.Instance,
	nsType string,
) ([]*rbacpb.Permission, error) {
	args := m.Called(c, sbj, nsType)

	return args.Get(0).([]*rbacpb.Permission), args.Error(1)
}

func (m *MockZanzibarLogic) Check(
	c context.Context,
	in *authzpbv1.CheckIn,
) (*authzpbv1.CheckOut, error) {
	args := m.Called(c, in)

	return args.Get(0).(*authzpbv1.CheckOut), args.Error(1)
}

func TestHandler_ListUsers(t *testing.T) {
	mockRbac := new(MockRbacLogic)
	mockZanzibar := new(MockZanzibarLogic)
	handler := rbac.NewHandler(mockZanzibar, mockRbac)

	req := &connect.Request[rbacpb.ListUsersIn]{
		Msg: &rbacpb.ListUsersIn{},
	}
	expectedRes := &rbacpb.ListUsersOut{
		Users: []*rbacpb.User{
			{Id: 1, Name: "test"},
		},
	}

	mockRbac.On("ListUsers", mock.Anything, req.Msg).Return(expectedRes, nil)

	res, err := handler.ListUsers(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, expectedRes, res.Msg)
	mockRbac.AssertExpectations(t)
}

func TestHandler_UpdateUser(t *testing.T) {
	mockRbac := new(MockRbacLogic)
	mockZanzibar := new(MockZanzibarLogic)
	handler := rbac.NewHandler(mockZanzibar, mockRbac)

	req := &connect.Request[rbacpb.UpdateUserIn]{
		Msg: &rbacpb.UpdateUserIn{Id: 1, Name: stringPtr("new name")},
	}

	mockRbac.On("UpdateUser", mock.Anything, req.Msg).Return(nil)

	_, err := handler.UpdateUser(context.Background(), req)

	assert.NoError(t, err)
	mockRbac.AssertExpectations(t)
}

func TestHandler_DeleteUser(t *testing.T) {
	mockRbac := new(MockRbacLogic)
	mockZanzibar := new(MockZanzibarLogic)
	handler := rbac.NewHandler(mockZanzibar, mockRbac)

	req := &connect.Request[rbacpb.DeleteUserIn]{
		Msg: &rbacpb.DeleteUserIn{Id: 1},
	}

	mockRbac.On("DeleteUser", mock.Anything, req.Msg).Return(nil)

	_, err := handler.DeleteUser(context.Background(), req)

	assert.NoError(t, err)
	mockRbac.AssertExpectations(t)
}

func TestHandler_CreateRole(t *testing.T) {
	mockRbac := new(MockRbacLogic)
	mockZanzibar := new(MockZanzibarLogic)
	handler := rbac.NewHandler(mockZanzibar, mockRbac)

	req := &connect.Request[rbacpb.CreateRoleIn]{
		Msg: &rbacpb.CreateRoleIn{Name: "new role"},
	}

	mockRbac.On("CreateRole", mock.Anything, req.Msg).Return(nil)

	_, err := handler.CreateRole(context.Background(), req)

	assert.NoError(t, err)
	mockRbac.AssertExpectations(t)
}

func TestHandler_ListRoles(t *testing.T) {
	mockRbac := new(MockRbacLogic)
	mockZanzibar := new(MockZanzibarLogic)
	handler := rbac.NewHandler(mockZanzibar, mockRbac)

	req := &connect.Request[rbacpb.ListRolesIn]{
		Msg: &rbacpb.ListRolesIn{},
	}
	expectedRes := &rbacpb.ListRolesOut{
		Roles: []*rbacpb.Role{
			{Id: 1, Name: "test role"},
		},
	}

	mockRbac.On("ListRoles", mock.Anything, req.Msg).Return(expectedRes, nil)

	res, err := handler.ListRoles(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, expectedRes, res.Msg)
	mockRbac.AssertExpectations(t)
}

func TestHandler_UpdateRole(t *testing.T) {
	mockRbac := new(MockRbacLogic)
	mockZanzibar := new(MockZanzibarLogic)
	handler := rbac.NewHandler(mockZanzibar, mockRbac)

	req := &connect.Request[rbacpb.UpdateRoleIn]{
		Msg: &rbacpb.UpdateRoleIn{Id: 1, Name: "new name"},
	}

	mockRbac.On("UpdateRole", mock.Anything, req.Msg).Return(nil)

	_, err := handler.UpdateRole(context.Background(), req)

	assert.NoError(t, err)
	mockRbac.AssertExpectations(t)
}

func TestHandler_DeleteRole(t *testing.T) {
	mockRbac := new(MockRbacLogic)
	mockZanzibar := new(MockZanzibarLogic)
	handler := rbac.NewHandler(mockZanzibar, mockRbac)

	req := &connect.Request[rbacpb.DeleteRoleIn]{
		Msg: &rbacpb.DeleteRoleIn{Id: 1},
	}

	mockRbac.On("DeleteRole", mock.Anything, req.Msg).Return(nil)

	_, err := handler.DeleteRole(context.Background(), req)

	assert.NoError(t, err)
	mockRbac.AssertExpectations(t)
}

func TestHandler_CreateResource(t *testing.T) {
	mockRbac := new(MockRbacLogic)
	mockZanzibar := new(MockZanzibarLogic)
	handler := rbac.NewHandler(mockZanzibar, mockRbac)

	req := &connect.Request[rbacpb.CreateResourceIn]{
		Msg: &rbacpb.CreateResourceIn{Ns: "test", Name: "test_resource"},
	}

	mockRbac.On("CreateResource", mock.Anything, req.Msg).Return(nil)

	_, err := handler.CreateResource(context.Background(), req)

	assert.NoError(t, err)
	mockRbac.AssertExpectations(t)
}

func TestHandler_ListResources(t *testing.T) {
	mockRbac := new(MockRbacLogic)
	mockZanzibar := new(MockZanzibarLogic)
	handler := rbac.NewHandler(mockZanzibar, mockRbac)

	req := &connect.Request[rbacpb.ListResourcesIn]{
		Msg: &rbacpb.ListResourcesIn{},
	}
	expectedRes := &rbacpb.ListResourcesOut{
		Resources: []*rbacpb.Resource{
			{Ns: "test", Name: "test_resource"},
		},
	}

	mockRbac.On("ListResources", mock.Anything, req.Msg).Return(expectedRes, nil)

	res, err := handler.ListResources(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, expectedRes, res.Msg)
	mockRbac.AssertExpectations(t)
}

func TestHandler_DeleteResource(t *testing.T) {
	mockRbac := new(MockRbacLogic)
	mockZanzibar := new(MockZanzibarLogic)
	handler := rbac.NewHandler(mockZanzibar, mockRbac)

	req := &connect.Request[rbacpb.DeleteResourceIn]{
		Msg: &rbacpb.DeleteResourceIn{Id: 1},
	}

	mockRbac.On("DeleteResource", mock.Anything, req.Msg).Return(nil)

	_, err := handler.DeleteResource(context.Background(), req)

	assert.NoError(t, err)
	mockRbac.AssertExpectations(t)
}

func TestHandler_AssignRole(t *testing.T) {
	mockRbac := new(MockRbacLogic)
	mockZanzibar := new(MockZanzibarLogic)
	handler := rbac.NewHandler(mockZanzibar, mockRbac)

	req := &connect.Request[rbacpb.AssignRoleIn]{
		Msg: &rbacpb.AssignRoleIn{UserId: 1, RoleId: 1},
	}

	mockRbac.On("AssignRole", mock.Anything, req.Msg).Return(nil)

	_, err := handler.AssignRole(context.Background(), req)

	assert.NoError(t, err)
	mockRbac.AssertExpectations(t)
}

func TestHandler_RevokeRole(t *testing.T) {
	mockRbac := new(MockRbacLogic)
	mockZanzibar := new(MockZanzibarLogic)
	handler := rbac.NewHandler(mockZanzibar, mockRbac)

	req := &connect.Request[rbacpb.RevokeRoleIn]{
		Msg: &rbacpb.RevokeRoleIn{UserId: 1, RoleId: 1},
	}

	mockRbac.On("RevokeRole", mock.Anything, req.Msg).Return(nil)

	_, err := handler.RevokeRole(context.Background(), req)

	assert.NoError(t, err)
	mockRbac.AssertExpectations(t)
}

func TestHandler_GrantPerm(t *testing.T) {
	mockRbac := new(MockRbacLogic)
	mockZanzibar := new(MockZanzibarLogic)
	handler := rbac.NewHandler(mockZanzibar, mockRbac)

	req := &connect.Request[rbacpb.GrantPermIn]{
		Msg: &rbacpb.GrantPermIn{RoleId: 1, ResourceId: 1, Perm: "read"},
	}

	mockRbac.On("GrantPerm", mock.Anything, req.Msg).Return(nil)

	_, err := handler.GrantPerm(context.Background(), req)

	assert.NoError(t, err)
	mockRbac.AssertExpectations(t)
}

func TestHandler_RevokePerm(t *testing.T) {
	mockRbac := new(MockRbacLogic)
	mockZanzibar := new(MockZanzibarLogic)
	handler := rbac.NewHandler(mockZanzibar, mockRbac)

	req := &connect.Request[rbacpb.RevokePermIn]{
		Msg: &rbacpb.RevokePermIn{RoleId: 1, ResourceId: 1, Perm: "read"},
	}

	mockRbac.On("RevokePerm", mock.Anything, req.Msg).Return(nil)

	_, err := handler.RevokePerm(context.Background(), req)

	assert.NoError(t, err)
	mockRbac.AssertExpectations(t)
}

func TestHandler_ListTuples(t *testing.T) {
	mockRbac := new(MockRbacLogic)
	mockZanzibar := new(MockZanzibarLogic)
	handler := rbac.NewHandler(mockZanzibar, mockRbac)

	req := &connect.Request[authzpbv1.ListTuplesIn]{
		Msg: &authzpbv1.ListTuplesIn{},
	}
	expectedRes := &authzpbv1.ListTuplesOut{
		Tuples: []*authzpbv1.Tuple{
			{SbjNs: "user", SbjId: "1", Rel: "member", ObjNs: "group", ObjId: "1"},
		},
	}

	mockZanzibar.On("List", mock.Anything, req.Msg).Return(expectedRes, nil)

	res, err := handler.ListTuples(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, expectedRes, res.Msg)
	mockZanzibar.AssertExpectations(t)
}

func TestHandler_CreateTuple(t *testing.T) {
	mockRbac := new(MockRbacLogic)
	mockZanzibar := new(MockZanzibarLogic)
	handler := rbac.NewHandler(mockZanzibar, mockRbac)

	req := &connect.Request[authzpbv1.Tuple]{
		Msg: &authzpbv1.Tuple{SbjNs: "user", SbjId: "1", Rel: "member", ObjNs: "group", ObjId: "1"},
	}

	mockZanzibar.On("Create", mock.Anything, req.Msg).Return(nil)

	_, err := handler.CreateTuple(context.Background(), req)

	assert.NoError(t, err)
	mockZanzibar.AssertExpectations(t)
}

func TestHandler_DeleteTuples(t *testing.T) {
	mockRbac := new(MockRbacLogic)
	mockZanzibar := new(MockZanzibarLogic)
	handler := rbac.NewHandler(mockZanzibar, mockRbac)

	req := &connect.Request[authzpbv1.DeleteTuplesIn]{
		Msg: &authzpbv1.DeleteTuplesIn{},
	}

	mockZanzibar.On("Delete", mock.Anything, req.Msg).Return(nil)

	_, err := handler.DeleteTuples(context.Background(), req)

	assert.NoError(t, err)
	mockZanzibar.AssertExpectations(t)
}

func TestHandler_Check(t *testing.T) {
	mockRbac := new(MockRbacLogic)
	mockZanzibar := new(MockZanzibarLogic)
	handler := rbac.NewHandler(mockZanzibar, mockRbac)

	req := &connect.Request[authzpbv1.CheckIn]{
		Msg: &authzpbv1.CheckIn{
			SbjNs: "user",
			SbjId: "1",
			Rel:   "member",
			ObjNs: "group",
			ObjId: "1",
		},
	}
	expectedRes := &authzpbv1.CheckOut{IsAllowed: true}

	mockZanzibar.On("Check", mock.Anything, req.Msg).Return(expectedRes, nil)

	res, err := handler.Check(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, expectedRes, res.Msg)
	mockZanzibar.AssertExpectations(t)
}

func TestHandler_ListResourceType(t *testing.T) {
	mockRbac := new(MockRbacLogic)
	mockZanzibar := new(MockZanzibarLogic)
	handler := rbac.NewHandler(mockZanzibar, mockRbac)

	req := &connect.Request[emptypb.Empty]{
		Msg: &emptypb.Empty{},
	}
	expectedRes := &rbacpb.ListResourceTypeOut{Types: []string{"type1", "type2"}}

	mockRbac.On("ListResourceTypes", mock.Anything).Return(expectedRes, nil)

	res, err := handler.ListResourceType(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, expectedRes, res.Msg)
	mockRbac.AssertExpectations(t)
}

func TestHandler_ListResourcesByType(t *testing.T) {
	mockRbac := new(MockRbacLogic)
	mockZanzibar := new(MockZanzibarLogic)
	handler := rbac.NewHandler(mockZanzibar, mockRbac)

	req := &connect.Request[rbacpb.ListResourcesByTypeIn]{
		Msg: &rbacpb.ListResourcesByTypeIn{Type: "type1"},
	}
	expectedRes := &rbacpb.ListResourcesByTypeOut{
		Resources: []*rbacpb.ListResourcesByTypeData{
			{Id: 1, Name: "resource1"},
		},
	}

	mockRbac.On("ListResourcesByType", mock.Anything, req.Msg).Return(expectedRes, nil)

	res, err := handler.ListResourcesByType(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, expectedRes, res.Msg)
	mockRbac.AssertExpectations(t)
}

func TestHandler_ListPermissionByResource(t *testing.T) {
	mockRbac := new(MockRbacLogic)
	mockZanzibar := new(MockZanzibarLogic)
	handler := rbac.NewHandler(mockZanzibar, mockRbac)

	req := &connect.Request[rbacpb.ListPermissionByResourceIn]{
		Msg: &rbacpb.ListPermissionByResourceIn{ResourceNs: "ns1", ResourceId: "1"},
	}
	expectedRes := &rbacpb.ListPermissionByResourceOut{Permissions: []string{"read", "write"}}

	mockRbac.On("ListPermissionByResource", mock.Anything, req.Msg).Return(expectedRes, nil)

	res, err := handler.ListPermissionByResource(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, expectedRes, res.Msg)
	mockRbac.AssertExpectations(t)
}

func TestHandler_GetRole(t *testing.T) {
	mockRbac := new(MockRbacLogic)
	mockZanzibar := new(MockZanzibarLogic)
	handler := rbac.NewHandler(mockZanzibar, mockRbac)

	req := &connect.Request[rbacpb.GetRoleIn]{
		Msg: &rbacpb.GetRoleIn{Id: 1},
	}
	expectedRes := &rbacpb.GetRoleOut{Ns: "role", Name: "test role"}

	mockRbac.On("GetRole", mock.Anything, req.Msg).Return(expectedRes, nil)

	res, err := handler.GetRole(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, expectedRes, res.Msg)
	mockRbac.AssertExpectations(t)
}

func stringPtr(s string) *string {
	return &s
}

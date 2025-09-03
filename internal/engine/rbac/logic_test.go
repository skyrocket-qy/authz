package rbac_test

import (
	"context"
	"testing"

	"authz/internal/engine/rbac"
	"authz/internal/entity/model"
	"authz/internal/schema"
	"authz/internal/zanzibar"
	"github.com/skyrocket-qy/protos/gen/authzpb/rbacpb"
	authzpbv1 "github.com/skyrocket-qy/protos/gen/authzpb/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// mockZanzibarLogic is a mock implementation of the ZanzibarLogic interface.
type mockZanzibarLogic struct {
	mock.Mock
	zanzibar.ZanzibarLogic
}

func (m *mockZanzibarLogic) Check(
	ctx context.Context,
	in *authzpbv1.CheckIn,
) (*authzpbv1.CheckOut, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*authzpbv1.CheckOut), args.Error(1)
}

func (m *mockZanzibarLogic) Create(ctx context.Context, in *authzpbv1.Tuple) error {
	args := m.Called(ctx, in)

	return args.Error(0)
}

func (m *mockZanzibarLogic) Delete(ctx context.Context, in *authzpbv1.DeleteTuplesIn) error {
	args := m.Called(ctx, in)

	return args.Error(0)
}

func (m *mockZanzibarLogic) GetPermissions(
	ctx context.Context,
	in *authzpbv1.Instance,
	s string,
) ([]*rbacpb.Permission, error) {
	args := m.Called(ctx, in, s)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*rbacpb.Permission), args.Error(1)
}

func (m *mockZanzibarLogic) List(
	ctx context.Context,
	in *authzpbv1.ListTuplesIn,
) (*authzpbv1.ListTuplesOut, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*authzpbv1.ListTuplesOut), args.Error(1)
}

func setupTestDB(t *testing.T, s *schema.Schema) (*gorm.DB, *mockZanzibarLogic, rbac.RbacLogic) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(&rbac.User{}, &rbac.Role{}, &rbac.Resource{}, &rbac.UserAuth{}, &model.Tuple{})
	if err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	if s == nil {
		s = &schema.Schema{Namespaces: map[string]*schema.Namespace{}}
	}

	mockZanzibar := new(mockZanzibarLogic)
	logic := rbac.NewRbacLogic(mockZanzibar, db, s)

	return db, mockZanzibar, logic
}

func TestRbacLogicImpl_ListUsers(t *testing.T) {
	db, _, logic := setupTestDB(t, nil)

	// Create test users
	users := []*rbac.User{
		{Email: "user1@example.com", Name: "user1"},
		{Email: "user2@example.com", Name: "user2"},
	}
	for _, u := range users {
		if err := db.Create(u).Error; err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
	}

	// Call ListUsers
	out, err := logic.ListUsers(context.Background(), &rbacpb.ListUsersIn{})

	// Assert the results
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, int64(2), out.GetCount())
	assert.Len(t, out.GetUsers(), 2)

	// Check user data
	for i, u := range out.GetUsers() {
		assert.Equal(t, users[i].Email, u.GetEmail())
		assert.Equal(t, users[i].Name, u.GetName())
	}
}

func TestRbacLogicImpl_UpdateUser(t *testing.T) {
	t.Run("should update a user successfully", func(t *testing.T) {
		db, _, logic := setupTestDB(t, nil)

		// Create a test user
		user := &rbac.User{Email: "user@example.com", Name: "testuser"}
		if err := db.Create(user).Error; err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		// Update the user
		newName := "updateduser"
		isActive := false
		err := logic.UpdateUser(context.Background(), &rbacpb.UpdateUserIn{
			Id:       user.Id,
			Name:     &newName,
			IsActive: &isActive,
		})
		assert.NoError(t, err)

		// Verify the update
		var updatedUser rbac.User
		if err := db.First(&updatedUser, user.Id).Error; err != nil {
			t.Fatalf("failed to find user: %v", err)
		}

		assert.Equal(t, newName, updatedUser.Name)
		assert.Equal(t, isActive, updatedUser.IsActive)
	})

	t.Run("should not return an error when updating a non-existent user", func(t *testing.T) {
		_, _, logic := setupTestDB(t, nil)

		// Attempt to update a non-existent user
		newName := "updateduser"
		isActive := false
		err := logic.UpdateUser(context.Background(), &rbacpb.UpdateUserIn{
			Id:       999, // Non-existent ID
			Name:     &newName,
			IsActive: &isActive,
		})

		// Assert that no error is returned
		assert.NoError(t, err)
	})
}

func TestRbacLogicImpl_DeleteUser(t *testing.T) {
	t.Run("should delete a user successfully", func(t *testing.T) {
		db, mockZanzibar, logic := setupTestDB(t, nil)

		// Create a test user
		user := &rbac.User{Email: "user@example.com", Name: "testuser"}
		if err := db.Create(user).Error; err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		// Mock Zanzibar delete
		mockZanzibar.On("Delete", mock.Anything, mock.Anything).Return(nil)

		// Delete the user
		err := logic.DeleteUser(context.Background(), &rbacpb.DeleteUserIn{Id: user.Id})
		assert.NoError(t, err)

		// Verify the user is deleted
		var deletedUser rbac.User

		err = db.First(&deletedUser, user.Id).Error
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})

	t.Run("should rollback transaction if zanzibar delete fails", func(t *testing.T) {
		db, mockZanzibar, logic := setupTestDB(t, nil)

		// Create a test user
		user := &rbac.User{Email: "user@example.com", Name: "testuser"}
		if err := db.Create(user).Error; err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		// Mock Zanzibar delete to return an error
		mockZanzibar.On("Delete", mock.Anything, mock.Anything).Return(assert.AnError)

		// Attempt to delete the user
		err := logic.DeleteUser(context.Background(), &rbacpb.DeleteUserIn{Id: user.Id})
		assert.Error(t, err)

		// Verify the user is not deleted
		var notDeletedUser rbac.User
		err = db.First(&notDeletedUser, user.Id).Error
		assert.NoError(t, err)
	})
}

func TestRbacLogicImpl_CreateRole(t *testing.T) {
	t.Run("should create a role successfully", func(t *testing.T) {
		db, _, logic := setupTestDB(t, nil)

		// Create a role
		roleName := "admin"
		err := logic.CreateRole(context.Background(), &rbacpb.CreateRoleIn{Name: roleName})
		assert.NoError(t, err)

		// Verify the role is created
		var role rbac.Role
		if err := db.Where("name = ?", roleName).First(&role).Error; err != nil {
			t.Fatalf("failed to find role: %v", err)
		}

		assert.Equal(t, roleName, role.Name)
	})

	t.Run("should return an error when creating a role with a duplicate name", func(t *testing.T) {
		db, _, logic := setupTestDB(t, nil)

		// Create a role
		roleName := "admin"
		err := logic.CreateRole(context.Background(), &rbacpb.CreateRoleIn{Name: roleName})
		assert.NoError(t, err)

		// Attempt to create another role with the same name
		err = logic.CreateRole(context.Background(), &rbacpb.CreateRoleIn{Name: roleName})
		assert.Error(t, err)

		// Verify that only one role with that name exists
		var count int64
		err = db.Model(&rbac.Role{}).Where("name = ?", roleName).Count(&count).Error
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})
}

func TestRbacLogicImpl_ListRoles(t *testing.T) {
	db, _, logic := setupTestDB(t, nil)

	// Create test roles
	roles := []*rbac.Role{
		{Name: "admin"},
		{Name: "editor"},
	}
	for _, r := range roles {
		if err := db.Create(r).Error; err != nil {
			t.Fatalf("failed to create role: %v", err)
		}
	}

	// Call ListRoles
	out, err := logic.ListRoles(context.Background(), &rbacpb.ListRolesIn{})

	// Assert the results
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, int64(2), out.GetTotal())
	assert.Len(t, out.GetRoles(), 2)

	// Check role data
	for i, r := range out.GetRoles() {
		assert.Equal(t, roles[i].Name, r.GetName())
	}
}

func TestRbacLogicImpl_UpdateRole(t *testing.T) {
	db, _, logic := setupTestDB(t, nil)

	// Create a test role
	role := &rbac.Role{Name: "admin"}
	if err := db.Create(role).Error; err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	// Update the role
	newName := "superadmin"
	err := logic.UpdateRole(context.Background(), &rbacpb.UpdateRoleIn{
		Id:   role.Id,
		Name: newName,
	})
	assert.NoError(t, err)

	// Verify the update
	var updatedRole rbac.Role
	if err := db.First(&updatedRole, role.Id).Error; err != nil {
		t.Fatalf("failed to find role: %v", err)
	}

	assert.Equal(t, newName, updatedRole.Name)
}

func TestRbacLogicImpl_DeleteRole(t *testing.T) {
	t.Run("should delete a role successfully", func(t *testing.T) {
		db, mockZanzibar, logic := setupTestDB(t, nil)

		// Create a test role
		role := &rbac.Role{Name: "admin"}
		if err := db.Create(role).Error; err != nil {
			t.Fatalf("failed to create role: %v", err)
		}

		// Mock Zanzibar delete
		mockZanzibar.On("Delete", mock.Anything, mock.Anything).Return(nil)

		// Delete the role
		err := logic.DeleteRole(context.Background(), &rbacpb.DeleteRoleIn{Id: role.Id})
		assert.NoError(t, err)

		// Verify the role is deleted
		var deletedRole rbac.Role

		err = db.First(&deletedRole, role.Id).Error
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})

	t.Run("should rollback transaction if zanzibar delete fails", func(t *testing.T) {
		db, mockZanzibar, logic := setupTestDB(t, nil)

		// Create a test role
		role := &rbac.Role{Name: "admin"}
		if err := db.Create(role).Error; err != nil {
			t.Fatalf("failed to create role: %v", err)
		}

		// Mock Zanzibar delete to return an error
		mockZanzibar.On("Delete", mock.Anything, mock.Anything).Return(assert.AnError)

		// Attempt to delete the role
		err := logic.DeleteRole(context.Background(), &rbacpb.DeleteRoleIn{Id: role.Id})
		assert.Error(t, err)

		// Verify the role is not deleted
		var notDeletedRole rbac.Role
		err = db.First(&notDeletedRole, role.Id).Error
		assert.NoError(t, err)
	})
}

func TestRbacLogicImpl_CreateResource(t *testing.T) {
	db, _, logic := setupTestDB(t, nil)

	// Create a resource
	resourceNs := "document"
	resourceName := "doc1"
	err := logic.CreateResource(context.Background(), &rbacpb.CreateResourceIn{
		Ns:   resourceNs,
		Name: resourceName,
	})
	assert.NoError(t, err)

	// Verify the resource is created
	var resource rbac.Resource
	if err := db.Where("ns = ? AND name = ?", resourceNs, resourceName).First(&resource).Error; err != nil {
		t.Fatalf("failed to find resource: %v", err)
	}

	assert.Equal(t, resourceNs, resource.Ns)
	assert.Equal(t, resourceName, resource.Name)
}

func TestRbacLogicImpl_ListResources(t *testing.T) {
	db, _, logic := setupTestDB(t, nil)

	// Create test resources
	resources := []*rbac.Resource{
		{Ns: "document", Name: "doc1"},
		{Ns: "image", Name: "img1"},
	}
	for _, r := range resources {
		if err := db.Create(r).Error; err != nil {
			t.Fatalf("failed to create resource: %v", err)
		}
	}

	// Call ListResources
	out, err := logic.ListResources(context.Background(), &rbacpb.ListResourcesIn{})

	// Assert the results
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, int64(2), out.GetTotal())
	assert.Len(t, out.GetResources(), 2)

	// Check resource data
	for i, r := range out.GetResources() {
		assert.Equal(t, resources[i].Ns, r.GetNs())
		assert.Equal(t, resources[i].Name, r.GetName())
	}
}

func TestRbacLogicImpl_DeleteResource(t *testing.T) {
	db, mockZanzibar, logic := setupTestDB(t, nil)

	// Create a test resource
	resource := &rbac.Resource{Ns: "document", Name: "doc1"}
	if err := db.Create(resource).Error; err != nil {
		t.Fatalf("failed to create resource: %v", err)
	}

	// Mock Zanzibar delete
	mockZanzibar.On("Delete", mock.Anything, mock.Anything).Return(nil)

	// Delete the resource
	err := logic.DeleteResource(context.Background(), &rbacpb.DeleteResourceIn{Id: resource.Id})
	assert.NoError(t, err)

	// Verify the resource is deleted
	var deletedResource rbac.Resource

	err = db.First(&deletedResource, resource.Id).Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestRbacLogicImpl_AssignRole(t *testing.T) {
	t.Run("should assign a role successfully", func(t *testing.T) {
		_, mockZanzibar, logic := setupTestDB(t, nil)

		// Mock Zanzibar create
		mockZanzibar.On("Create", mock.Anything, mock.Anything).Return(nil)

		// Assign a role
		err := logic.AssignRole(context.Background(), &rbacpb.AssignRoleIn{
			UserId: 1,
			RoleId: 1,
		})
		assert.NoError(t, err)

		// Assert that Zanzibar's Create method was called
		mockZanzibar.AssertCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("should call zanzibar create even with non-existent user and role ids", func(t *testing.T) {
		_, mockZanzibar, logic := setupTestDB(t, nil)

		// Mock Zanzibar create
		mockZanzibar.On("Create", mock.Anything, mock.Anything).Return(nil)

		// Assign a role with non-existent IDs
		err := logic.AssignRole(context.Background(), &rbacpb.AssignRoleIn{
			UserId: 999,
			RoleId: 999,
		})
		assert.NoError(t, err)

		// Assert that Zanzibar's Create method was still called
		mockZanzibar.AssertCalled(t, "Create", mock.Anything, mock.Anything)
	})
}

func TestRbacLogicImpl_GetRole(t *testing.T) {
	db, mockZanzibar, logic := setupTestDB(t, nil)

	// Create a test role
	role := &rbac.Role{Id: 1, Name: "admin"}
	if err := db.Create(role).Error; err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	// Mock Zanzibar GetPermissions
	expectedPerms := []*rbacpb.Permission{{Permission: "read"}, {Permission: "write"}}
	mockZanzibar.On("GetPermissions", mock.Anything, mock.Anything, "resource").
		Return(expectedPerms, nil)

	// Get the role
	out, err := logic.GetRole(context.Background(), &rbacpb.GetRoleIn{Id: role.Id})
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, role.Name, out.GetName())
	assert.Equal(t, "role", out.GetNs())
	assert.Equal(t, expectedPerms, out.GetPermissions())
}

func TestRbacLogicImpl_ListResourcesByType(t *testing.T) {
	db, _, logic := setupTestDB(t, nil)

	// Create test resources
	resources := []*rbac.Resource{
		{Ns: "document", Name: "doc1"},
		{Ns: "document", Name: "doc2"},
		{Ns: "image", Name: "img1"},
	}
	for _, r := range resources {
		if err := db.Create(r).Error; err != nil {
			t.Fatalf("failed to create resource: %v", err)
		}
	}

	// Call ListResourcesByType
	out, err := logic.ListResourcesByType(
		context.Background(),
		&rbacpb.ListResourcesByTypeIn{Type: "document"},
	)

	// Assert the results
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Len(t, out.GetResources(), 2)
	assert.Equal(t, "doc1", out.GetResources()[0].GetName())
	assert.Equal(t, "doc2", out.GetResources()[1].GetName())
}

func TestRbacLogicImpl_ListResourceTypes(t *testing.T) {
	// Create a mock schema
	mockSchema := &schema.Schema{
		Namespaces: map[string]*schema.Namespace{
			"document": {Type: "resource"},
			"image":    {Type: "resource"},
			"user":     {Type: "user"},
		},
	}
	_, _, logic := setupTestDB(t, mockSchema)

	// Call ListResourceTypes
	out, err := logic.ListResourceTypes(context.Background())

	// Assert the results
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.ElementsMatch(t, []string{"document", "image"}, out.GetTypes())
}

func TestRbacLogicImpl_ListPermissionByResource(t *testing.T) {
	t.Run("should list available permissions for a resource", func(t *testing.T) {
		// Create a mock schema
		mockSchema := &schema.Schema{
			Namespaces: map[string]*schema.Namespace{
				"document": {
					Type: "resource",
					Relations: map[string]*schema.Relation{
						"read":  {},
						"write": {},
						"owner": {},
					},
				},
			},
		}
		db, _, logic := setupTestDB(t, mockSchema)

		// Create a test tuple
		tuple := &model.Tuple{
			SbjNs:    "role",
			SbjId:    "1",
			Relation: "read",
			ObjNs:    "document",
			ObjId:    "1",
		}

		if err := db.Create(tuple).Error; err != nil {
			t.Fatalf("failed to create tuple: %v", err)
		}

		// Call ListPermissionByResource
		out, err := logic.ListPermissionByResource(
			context.Background(),
			&rbacpb.ListPermissionByResourceIn{
				RoleId:     "1",
				ResourceNs: "document",
				ResourceId: "1",
			},
		)

		// Assert the results
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.ElementsMatch(t, []string{"write", "owner"}, out.GetPermissions())
	})

	t.Run("should not panic with non-existent namespace", func(t *testing.T) {
		_, _, logic := setupTestDB(t, &schema.Schema{
			Namespaces: map[string]*schema.Namespace{},
		})

		out, err := logic.ListPermissionByResource(
			context.Background(),
			&rbacpb.ListPermissionByResourceIn{
				RoleId:     "1",
				ResourceNs: "non-existent-namespace",
				ResourceId: "1",
			},
		)

		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.Empty(t, out.GetPermissions())
	})
}

func TestRbacLogicImpl_RevokeRole(t *testing.T) {
	_, mockZanzibar, logic := setupTestDB(t, nil)

	// Mock Zanzibar delete
	mockZanzibar.On("Delete", mock.Anything, mock.Anything).Return(nil)

	// Revoke a role
	err := logic.RevokeRole(context.Background(), &rbacpb.RevokeRoleIn{
		UserId: 1,
		RoleId: 1,
	})
	assert.NoError(t, err)
}

func TestRbacLogicImpl_GrantPerm(t *testing.T) {
	_, mockZanzibar, logic := setupTestDB(t, nil)

	// Mock Zanzibar create
	mockZanzibar.On("Create", mock.Anything, mock.Anything).Return(nil)

	// Grant a permission
	err := logic.GrantPerm(context.Background(), &rbacpb.GrantPermIn{
		RoleId:     1,
		ResourceId: 1,
		Perm:       "read",
	})
	assert.NoError(t, err)
}

func TestRbacLogicImpl_RevokePerm(t *testing.T) {
	t.Run("should revoke a permission successfully", func(t *testing.T) {
		db, mockZanzibar, logic := setupTestDB(t, nil)

		// Create a test resource
		resource := &rbac.Resource{Id: 1, Ns: "document", Name: "doc1"}
		if err := db.Create(resource).Error; err != nil {
			t.Fatalf("failed to create resource: %v", err)
		}

		// Mock Zanzibar delete
		mockZanzibar.On("Delete", mock.Anything, mock.Anything).Return(nil)

		// Revoke a permission
		err := logic.RevokePerm(context.Background(), &rbacpb.RevokePermIn{
			RoleId:     1,
			ResourceId: 1,
			Perm:       "read",
		})
		assert.NoError(t, err)
	})

	t.Run("should return an error when revoking a permission for a non-existent resource", func(t *testing.T) {
		_, _, logic := setupTestDB(t, nil)

		// Attempt to revoke a permission for a non-existent resource
		err := logic.RevokePerm(context.Background(), &rbacpb.RevokePermIn{
			RoleId:     1,
			ResourceId: 999, // Non-existent ID
			Perm:       "read",
		})

		// Assert that an error is returned
		assert.Error(t, err)
	})
}

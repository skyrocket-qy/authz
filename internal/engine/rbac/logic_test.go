package rbac

import (
	"context"
	"testing"

	"authz/internal/zanzibar"

	"github.com/skyrocket-qy/protos/gen/authzpb/rbacpb"
	authzpbv1 "github.com/skyrocket-qy/protos/gen/authzpb/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockZanzibarLogic is a mock implementation of the ZanzibarLogic interface.
type MockZanzibarLogic struct {
	mock.Mock
	zanzibar.ZanzibarLogic
}

func (m *MockZanzibarLogic) Check(ctx context.Context, in *authzpbv1.CheckIn) (*authzpbv1.CheckOut, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authzpbv1.CheckOut), args.Error(1)
}

func (m *MockZanzibarLogic) Create(ctx context.Context, in *authzpbv1.Tuple) error {
	args := m.Called(ctx, in)
	return args.Error(0)
}

func (m *MockZanzibarLogic) Delete(ctx context.Context, in *authzpbv1.DeleteTuplesIn) error {
	args := m.Called(ctx, in)
	return args.Error(0)
}

func (m *MockZanzibarLogic) GetPermissions(ctx context.Context, in *authzpbv1.Instance, s string) ([]*rbacpb.Permission, error) {
	args := m.Called(ctx, in, s)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*rbacpb.Permission), args.Error(1)
}

func (m *MockZanzibarLogic) List(ctx context.Context, in *authzpbv1.ListTuplesIn) (*authzpbv1.ListTuplesOut, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authzpbv1.ListTuplesOut), args.Error(1)
}

func setupTestDB(t *testing.T) (*gorm.DB, *MockZanzibarLogic, RbacLogic) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(&User{}, &Role{}, &Resource{}, &UserAuth{})
	if err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	mockZanzibar := new(MockZanzibarLogic)
	logic := NewRbacLogic(mockZanzibar, db, nil)

	return db, mockZanzibar, logic
}

func TestRbacLogicImpl_ListUsers(t *testing.T) {
	db, _, logic := setupTestDB(t)

	// Create test users
	users := []*User{
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
	assert.Equal(t, int64(2), out.Count)
	assert.Len(t, out.Users, 2)

	// Check user data
	for i, u := range out.Users {
		assert.Equal(t, users[i].Email, u.Email)
		assert.Equal(t, users[i].Name, u.Name)
	}
}

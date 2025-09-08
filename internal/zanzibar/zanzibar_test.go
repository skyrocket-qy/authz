package zanzibar_test

import (
	"context"
	"testing"

	"authz/internal/entity"
	"authz/internal/schema"
	"authz/internal/zanzibar"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	authzpbv1 "github.com/skyrocket-qy/protos/gen/authzpb/v1"
	pkgpbv1 "github.com/skyrocket-qy/protos/gen/pkgpb/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type mockZanzibarMemory struct {
	mock.Mock
}

func (m *mockZanzibarMemory) Check(
	c context.Context,
	sbj entity.Instance,
	rel string,
	obj entity.Instance,
) (bool, error) {
	args := m.Called(c, sbj, rel, obj)

	return args.Bool(0), args.Error(1)
}

func (m *mockZanzibarMemory) Lookup(
	c context.Context,
	sbj entity.Instance,
	rel string,
) ([]entity.Instance, error) {
	args := m.Called(c, sbj, rel)

	return args.Get(0).([]entity.Instance), args.Error(1)
}

func (m *mockZanzibarMemory) Expand(
	c context.Context,
	rel string,
	obj entity.Instance,
) ([]entity.Instance, error) {
	args := m.Called(c, rel, obj)

	return args.Get(0).([]entity.Instance), args.Error(1)
}

func newMockRedis(t *testing.T) (*redis.Client, redismock.ClientMock) {
	t.Helper()

	db, mock := redismock.NewClientMock()

	return db, mock
}

func TestZanzibarLogicImpl_Check(t *testing.T) {
	mockZm := new(mockZanzibarMemory)
	logic, err := zanzibar.NewZanzibarLogic(nil, mockZm, nil, nil)
	assert.NoError(t, err)

	ctx := context.Background()
	in := &authzpbv1.CheckIn{
		SbjNs: "user",
		SbjId: "1",
		Rel:   "viewer",
		ObjNs: "doc",
		ObjId: "1",
	}
	sbj := entity.Instance{Ns: "user", Id: "1"}
	obj := entity.Instance{Ns: "doc", Id: "1"}

	mockZm.On("Check", ctx, sbj, "viewer", obj).Return(true, nil)

	out, err := logic.Check(ctx, in)
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.True(t, out.GetIsAllowed())

	mockZm.AssertExpectations(t)
}

func newMockDb(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	gormDb, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	assert.NoError(t, err)

	return gormDb, mock
}

func TestZanzibarLogicImpl_Create(t *testing.T) {
	db, mock := newMockDb(t)
	logic, err := zanzibar.NewZanzibarLogic(db, nil, nil, nil)
	assert.NoError(t, err)

	ctx := context.Background()
	tuple := &authzpbv1.Tuple{
		SbjNs: "user",
		SbjId: "1",
		Rel:   "owner",
		ObjNs: "doc",
		ObjId: "1",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "tuples"`).
		WithArgs(tuple.GetSbjNs(), tuple.GetSbjId(), tuple.GetRel(), tuple.GetObjNs(), tuple.GetObjId()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err = logic.Create(ctx, tuple)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestZanzibarLogicImpl_Delete(t *testing.T) {
	db, mock := newMockDb(t)
	logic, err := zanzibar.NewZanzibarLogic(db, nil, nil, nil)
	assert.NoError(t, err)
	ctx := context.Background()

	// Test case 1: Delete by filter
	t.Run("DeleteByFilter", func(t *testing.T) {
		userNs := "user"
		userId := "1"
		filter := &authzpbv1.TupleFilter{
			SbjNs: &userNs,
			SbjId: &userId,
		}
		in := &authzpbv1.DeleteTuplesIn{
			Mode: &authzpbv1.DeleteTuplesIn_Filter{
				Filter: filter,
			},
		}

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM "tuples" WHERE sbj_ns = \$1 AND sbj_id = \$2`).
			WithArgs("user", "1").
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := logic.Delete(ctx, in)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// Test case 2: Delete by list of tuples
	t.Run("DeleteByTuples", func(t *testing.T) {
		tuples := []*authzpbv1.Tuple{
			{SbjNs: "user", SbjId: "1", Rel: "owner", ObjNs: "doc", ObjId: "1"},
			{SbjNs: "user", SbjId: "2", Rel: "editor", ObjNs: "doc", ObjId: "2"},
		}
		in := &authzpbv1.DeleteTuplesIn{
			Mode: &authzpbv1.DeleteTuplesIn_DeleteTuples{
				DeleteTuples: &authzpbv1.DeleteTuples{
					Tuples: tuples,
				},
			},
		}

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM "tuples" WHERE \(sbj_ns, sbj_id, rel, obj_ns, obj_id\) IN \(\(\$1,\$2,\$3,\$4,\$5\),\(\$6,\$7,\$8,\$9,\$10\)\)`).
			WithArgs("user", "1", "owner", "doc", "1", "user", "2", "editor", "doc", "2").
			WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectCommit()

		err := logic.Delete(ctx, in)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// Test case 3: Delete by list of tuple IDs
	t.Run("DeleteByIds", func(t *testing.T) {
		ids := []uint64{1, 2, 3}
		in := &authzpbv1.DeleteTuplesIn{
			Mode: &authzpbv1.DeleteTuplesIn_DeleteTupleIds{
				DeleteTupleIds: &authzpbv1.DeleteTupleIds{
					Ids: ids,
				},
			},
		}

		mock.ExpectBegin()
		mock.ExpectExec(`DELETE FROM "tuples" WHERE id IN \(\$1,\$2,\$3\)`).
			WithArgs(1, 2, 3).
			WillReturnResult(sqlmock.NewResult(0, 3))
		mock.ExpectCommit()

		err := logic.Delete(ctx, in)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestZanzibarLogicImpl_Find(t *testing.T) {
	db, mock := newMockDb(t)
	logic, err := zanzibar.NewZanzibarLogic(db, nil, nil, nil)
	assert.NoError(t, err)
	ctx := context.Background()
	userNs := "user"
	filter := &authzpbv1.TupleFilter{
		SbjNs: &userNs,
	}

	rows := sqlmock.NewRows([]string{"sbj_ns", "sbj_id", "relation", "obj_ns", "obj_id"}).
		AddRow("user", "1", "owner", "doc", "1").
		AddRow("user", "2", "editor", "doc", "2")

	mock.ExpectQuery(`SELECT \* FROM "tuples" WHERE sbj_ns = \$1`).
		WithArgs("user").
		WillReturnRows(rows)

	tuples, err := logic.Find(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, tuples, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestZanzibarLogicImpl_List(t *testing.T) {
	db, mock := newMockDb(t)
	logic, err := zanzibar.NewZanzibarLogic(db, nil, nil, nil)
	assert.NoError(t, err)
	ctx := context.Background()
	in := &authzpbv1.ListTuplesIn{
		Cursor: &pkgpbv1.Cursor{
			Size: 10,
		},
		Filters: []*pkgpbv1.Filter{
			{Field: "sbj_ns", Op: pkgpbv1.Operator_EQ, Values: []string{"user"}},
		},
	}

	rows := sqlmock.NewRows([]string{"sbj_ns", "sbj_id", "relation", "obj_ns", "obj_id"}).
		AddRow("user", "1", "owner", "doc", "1").
		AddRow("user", "2", "editor", "doc", "2")

	mock.ExpectQuery("SELECT \\* FROM \"tuples\" WHERE `sbj_ns` = \\$1 LIMIT \\$2").
		WithArgs("user", 10).
		WillReturnRows(rows)

	out, err := logic.List(ctx, in)
	assert.NoError(t, err)
	assert.Len(t, out.GetTuples(), 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestZanzibarLogicImpl_GetPermissions(t *testing.T) {
	db, mock := newMockDb(t)
	s := &schema.Schema{
		Namespaces: map[string]*schema.Namespace{
			"doc": {
				Type: "resource",
				Relations: map[string]*schema.Relation{
					"owner": {
						UsersetRewrite: schema.UsersetRewrite{
							ComputedUserSet: &schema.ComputedUserset{
								Relation: "owner",
							},
						},
					},
				},
			},
		},
	}
	logic, err := zanzibar.NewZanzibarLogic(db, nil, s, nil)
	assert.NoError(t, err)
	ctx := context.Background()
	sbj := &authzpbv1.Instance{
		Ns: "user",
		Id: "1",
	}

	rows := sqlmock.NewRows([]string{"obj_ns", "obj_id", "relation"}).
		AddRow("doc", "1", "owner")

	mock.ExpectQuery(`SELECT \* FROM "tuples" WHERE sbj_ns = \$1 AND sbj_id = \$2`).
		WithArgs("user", "1").
		WillReturnRows(rows)

	perms, err := logic.GetPermissions(ctx, sbj, "resource")
	assert.NoError(t, err)
	assert.Len(t, perms, 1)
	assert.Equal(t, "doc", perms[0].GetResourceNs())
	assert.Equal(t, "1", perms[0].GetResourceId())
	assert.Equal(t, "owner", perms[0].GetPermission())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestZanzibarLogicImpl_Lookup(t *testing.T) {
	mockZm := new(mockZanzibarMemory)
	logic, err := zanzibar.NewZanzibarLogic(nil, mockZm, nil, nil)
	assert.NoError(t, err)

	ctx := context.Background()
	user := entity.Instance{Ns: "user", Id: "1"}
	expectedObjs := []entity.Instance{
		{Ns: "doc", Id: "1"},
		{Ns: "doc", Id: "2"},
	}

	mockZm.On("Lookup", ctx, user, "owner").Return(expectedObjs, nil)

	objs, err := logic.Lookup(ctx, user, "owner")
	assert.NoError(t, err)
	assert.Equal(t, expectedObjs, objs)

	mockZm.AssertExpectations(t)
}

func TestZanzibarLogicImpl_Expand(t *testing.T) {
	mockZm := new(mockZanzibarMemory)
	logic, err := zanzibar.NewZanzibarLogic(nil, mockZm, nil, nil)
	assert.NoError(t, err)

	ctx := context.Background()
	doc := entity.Instance{Ns: "doc", Id: "1"}
	expectedUsers := []entity.Instance{
		{Ns: "user", Id: "1"},
		{Ns: "user", Id: "2"},
	}

	mockZm.On("Expand", ctx, "owner", doc).Return(expectedUsers, nil)

	users, err := logic.Expand(ctx, "owner", doc)
	assert.NoError(t, err)
	assert.Equal(t, expectedUsers, users)

	mockZm.AssertExpectations(t)
}

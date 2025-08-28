package zanzibar

import (
	"context"
	"testing"

	"authz/internal/entity"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"authz/internal/schema"
	authzpbv1 "github.com/skyrocket-qy/protos/gen/authzpb/v1"
	pkgpbv1 "github.com/skyrocket-qy/protos/gen/pkgpb/v1"
)

type mockZanzibarMemory struct {
	mock.Mock
}

func (m *mockZanzibarMemory) Check(c context.Context, sbj entity.Instance, rel string, obj entity.Instance) (bool, error) {
	args := m.Called(c, sbj, rel, obj)
	return args.Bool(0), args.Error(1)
}

func TestZanzibarLogicImpl_Check(t *testing.T) {
	mockZm := new(mockZanzibarMemory)
	logic := &ZanzibarLogicImpl{
		zm: mockZm,
	}

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
	assert.True(t, out.IsAllowed)

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
	logic := &ZanzibarLogicImpl{
		pgdb: db,
	}

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
		WithArgs(tuple.SbjNs, tuple.SbjId, tuple.Rel, tuple.ObjNs, tuple.ObjId).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := logic.Create(ctx, tuple)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestZanzibarLogicImpl_Delete(t *testing.T) {
	db, mock := newMockDb(t)
	logic := &ZanzibarLogicImpl{
		pgdb: db,
	}
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
	})

	// TODO: Fix the tests for DeleteByTuples and DeleteByIds. The proto message types are not correct.
	// // Test case 2: Delete by list of tuples
	// t.Run("DeleteByTuples", func(t *testing.T) {
	// 	db, mock := newMockDb(t)
	// 	logic := &ZanzibarLogicImpl{
	// 		pgdb: db,
	// 	}
	// 	tuples := []*authzpbv1.Tuple{
	// 		{SbjNs: "user", SbjId: "1", Rel: "owner", ObjNs: "doc", ObjId: "1"},
	// 		{SbjNs: "user", SbjId: "2", Rel: "editor", ObjNs: "doc", ObjId: "2"},
	// 	}
	// 	in := &authzpbv1.DeleteTuplesIn{
	// 		Mode: &authzpbv1.DeleteTuplesIn_DeleteTuples{
	// 			DeleteTuples: &authzpbv1.DeleteTuplesIn_DeleteTupleList{
	// 				Tuples: tuples,
	// 			},
	// 		},
	// 	}

	// 	mock.ExpectBegin()
	// 	mock.ExpectExec(`DELETE FROM "tuples" WHERE \(sbj_ns, sbj_id, rel, obj_ns, obj_id\) IN \(\(\\$1,\\$2,\\$3,\\$4,\\$5\),\(\\$6,\\$7,\\$8,\\$9,\\$10\)\)`).
	// 		WithArgs("user", "1", "owner", "doc", "1", "user", "2", "editor", "doc", "2").
	// 		WillReturnResult(sqlmock.NewResult(0, 2))
	// 	mock.ExpectCommit()

	// 	err := logic.Delete(ctx, in)
	// 	assert.NoError(t, err)
	// })

	// // Test case 3: Delete by list of tuple IDs
	// t.Run("DeleteByIds", func(t *testing.T) {
	// 	db, mock := newMockDb(t)
	// 	logic := &ZanzibarLogicImpl{
	// 		pgdb: db,
	// 	}
	// 	ids := []uint32{1, 2, 3}
	// 	in := &authzpbv1.DeleteTuplesIn{
	// 		Mode: &authzpbv1.DeleteTuplesIn_DeleteTupleIds{
	// 			DeleteTupleIds: &authzpbv1.DeleteTuplesIn_DeleteTupleIdList{
	// 				Ids: ids,
	// 			},
	// 		},
	// 	}

	// 	mock.ExpectBegin()
	// 	mock.ExpectExec(`DELETE FROM "tuples" WHERE id IN \(\\$1,\\$2,\\$3\)`).
	// 		WithArgs(1, 2, 3).
	// 		WillReturnResult(sqlmock.NewResult(0, 3))
	// 	mock.ExpectCommit()

	// 	err := logic.Delete(ctx, in)
	// 	assert.NoError(t, err)
	// })
}
func TestZanzibarLogicImpl_Find(t *testing.T) {
	db, mock := newMockDb(t)
	logic := &ZanzibarLogicImpl{
		pgdb: db,
	}
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
	logic := &ZanzibarLogicImpl{
		pgdb: db,
	}
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
	assert.Len(t, out.Tuples, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestZanzibarLogicImpl_GetPermissions(t *testing.T) {
	db, mock := newMockDb(t)
	s := &schema.Schema{
		Namespaces: map[string]*schema.Namespace{
			"doc": {
				Type: "resource",
				Relations: map[string]*schema.Relation{
					"owner": &schema.Relation{
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
	logic := &ZanzibarLogicImpl{
		pgdb:   db,
		schema: s,
	}
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
	assert.Equal(t, "doc", perms[0].ResourceNs)
	assert.Equal(t, "1", perms[0].ResourceId)
	assert.Equal(t, "owner", perms[0].Permission)
	assert.NoError(t, mock.ExpectationsWereMet())
}

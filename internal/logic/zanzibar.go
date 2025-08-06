package logic

import (
	"authz/internal/entity"
	"authz/internal/entity/model"
	"authz/internal/pkg"
	"context"
	"strings"

	"github.com/redis/go-redis/v9"
	authzpbv1 "github.com/skyrocket-qy/protos/gen/authzpb/v1"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

type ZanzibarLogic interface {
	Create(c context.Context, tuple *authzpbv1.Tuple) error
	List(c context.Context, in *authzpbv1.ListTuplesIn) (*authzpbv1.ListTuplesOut, error)
	Find(c context.Context, filter *authzpbv1.TupleFilter) ([]*authzpbv1.Tuple, error)
	Delete(c context.Context, filter *authzpbv1.TupleFilter) error

	Check(c context.Context, sbj *entity.Instance, rel string, obj *entity.Instance) (bool, error)
	Lookup(c context.Context, sbj *entity.Instance, rel string) ([]*entity.Instance, error)
	Expand(c context.Context, relation string, obj *entity.Instance) ([]entity.Instance, error)
}

var _ ZanzibarLogic = (*ZanzibarLogicImpl)(nil)

type ZanzibarLogicImpl struct {
	pgdb *gorm.DB
	rdb  *redis.Client
}

func NewZanzibarLogic(db *gorm.DB, rdb *redis.Client) *ZanzibarLogicImpl {
	return &ZanzibarLogicImpl{
		pgdb: db,
		rdb:  rdb,
	}
}

func (r *ZanzibarLogicImpl) db(c context.Context) *gorm.DB {
	return r.pgdb.WithContext(c)
}

func (r *ZanzibarLogicImpl) List(c context.Context, in *authzpbv1.ListTuplesIn) (
	out *authzpbv1.ListTuplesOut, err error,
) {
	validFilterFields := []string{"sbj_ns", "sbj_id", "relation", "obj_ns", "obj_id"}

	filterScope, err := pkg.ApplyFilter(in.Filters, validFilterFields, nil)
	if err != nil {
		return nil, err
	}

	cnt := int64(0)
	if err := r.db(c).Model(&model.Tuple{}).Scopes(filterScope).Count(&cnt).Error; err != nil {
		return nil, err
	}

	tupleModels := []*model.Tuple{}
	if err := r.db(c).
		Scopes(
			filterScope,
			pkg.ApplyPager(in.Pager),
			pkg.ApplySorter(in.Sorters),
		).
		Find(&tupleModels).Error; err != nil {
		return nil, err
	}

	tuples := make([]*authzpbv1.Tuple, len(tupleModels))
	for i, tuple := range tupleModels {
		tuples[i] = &authzpbv1.Tuple{
			SbjNs: tuple.SbjNs,
			SbjId: tuple.SbjId,
			Rel:   tuple.Relation,
			ObjNs: tuple.ObjNs,
			ObjId: tuple.ObjId,
		}
	}
	return &authzpbv1.ListTuplesOut{
		Tuples: tuples,
		Total:  cnt,
	}, nil
}

// TODO: pagination and limit , sorter
func (r *ZanzibarLogicImpl) Find(c context.Context, filter *authzpbv1.TupleFilter) (
	[]*authzpbv1.Tuple, error,
) {
	tuples := []*model.Tuple{}
	if err := r.db(c).Scopes(ApplyTupleFilter(filter)).Find(&tuples).Error; err != nil {
		return nil, err
	}

	tuplesProto := make([]*authzpbv1.Tuple, 0, len(tuples))
	for _, tuple := range tuples {
		tuplesProto = append(tuplesProto, &authzpbv1.Tuple{
			SbjNs: tuple.SbjNs,
			SbjId: tuple.SbjId,
			Rel:   tuple.Relation,
			ObjNs: tuple.ObjNs,
			ObjId: tuple.ObjId,
		})
	}

	return tuplesProto, nil
}

func (r *ZanzibarLogicImpl) Create(c context.Context, tuple *authzpbv1.Tuple) error {
	tupleModel := model.Tuple{
		SbjNs:    tuple.SbjNs,
		SbjId:    tuple.SbjId,
		Relation: tuple.Rel,
		ObjNs:    tuple.ObjNs,
		ObjId:    tuple.ObjId,
	}

	tupleBytes, err := proto.Marshal(tuple)
	if err != nil {
		return err
	}

	if err := r.db(c).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&tupleModel).Error; err != nil {
			return err
		}

		if err := tx.Create(&model.ChangeLog{
			Tuple: tupleBytes,
		}).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (r *ZanzibarLogicImpl) Delete(c context.Context, tuples []*authzpbv1.Tuple) error {
	changelogs := make([]model.ChangeLog, len(tuples))
	for i, tuple := range tuples {
		data, err := proto.Marshal(tuple)
		if err != nil {
			return err
		}

		changelogs[i] = model.ChangeLog{Tuple: data}
	}

	return r.db(c).Transaction(func(tx *gorm.DB) error {
		values := make([]any, 0, len(tuples)*5)
		placeholders := make([]string, len(tuples))

		for i, t := range tuples {
			placeholders[i] = "(?, ?, ?, ?, ?)"
			values = append(values, t.SbjNs, t.SbjId, t.Rel, t.ObjNs, t.ObjId)
		}

		query := `
			DELETE FROM tuples
			WHERE (sbj_ns, sbj_id, rel, obj_ns, obj_id) IN (` + strings.Join(placeholders, ",") + `)
		`
		if err := tx.Unscoped().Exec(query, values...).Error; err != nil {
			return err
		}

		if err := tx.Create(&changelogs).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *ZanzibarLogicImpl) Check(c context.Context, user *entity.Instance, perm string,
	obj *entity.Instance) (bool, error,
) {
	return false, nil
}

// What are all the objs that sbj has rel on
func (r *ZanzibarLogicImpl) Lookup(c context.Context, user *entity.Instance, perm string) (
	objs []*entity.Instance, err error,
) {

	return objs, nil
}

// func StartCleanJob(r *ZanzibarLogicImpl, cro *cron.Cron) error {
// 	return cro.AddFunc("0 0 * * *", func() {
// 		for {
// 			done, err := r.CleanUnused()
// 			if err != nil {
// 				logx.Error(err.Error())
// 				time.Sleep(15 * time.Minute)
// 				continue
// 			}

// 			if done {
// 				break
// 			}

// 			time.Sleep(15 * time.Minute) // TODO: retry exponentially
// 		}
// 	})
// }

func (r *ZanzibarLogicImpl) Expand(c context.Context, perm string, obj *entity.Instance) (
	users []entity.Instance, err error,
) {
	return nil, nil
}

// func (r *ZanzibarLogicImpl) CleanUnused() (done bool, err error) {
// 	ctx := context.Background()

// 	lockKey := "zanzibar:clean-daily:lock"
// 	lockValue := uuid.New().String()
// 	ok, err := r.rdb.SetNX(ctx, lockKey, lockValue, 24*time.Hour).Result()
// 	if err != nil {
// 		return false, err
// 	}
// 	if !ok {
// 		logx.Info("Another clean job is already running.")
// 		return false, nil
// 	}
// 	defer func() {
// 		script := `
// 			if redis.call("get", KEYS[1]) == ARGV[1] then
// 				return redis.call("del", KEYS[1])
// 			else
// 				return 0
// 			end
// 		`
// 		if _, err := r.rdb.Eval(ctx, script, []string{lockKey}, lockValue).Result(); err != nil {
// 			logx.Errorf("Failed to release lock: %v", err)
// 		}
// 	}()

// 	val, err := r.rdb.Get(ctx, "zanzibar:clean-daily:done").Result()
// 	if err != nil && err != redis.Nil {
// 		return false, err
// 	}
// 	if val != "" {
// 		logx.Info("Clean job already completed.")
// 		return true, nil
// 	}

// 	if err := r.cleanUnusedNamespaces(ctx); err != nil {
// 		return false, err
// 	}
// 	if err := r.cleanUnusedRelationDefs(ctx); err != nil {
// 		return false, err
// 	}
// 	if err := r.cleanUnusedInstances(ctx); err != nil {
// 		return false, err
// 	}

// 	if err := r.rdb.Set(ctx, "zanzibar:clean-daily:done",
// 		time.Now().Format(time.RFC3339), 24*time.Hour).Err(); err != nil {
// 		return true, err
// 	}

// 	return true, nil
// }

// func (r *ZanzibarLogicImpl) cleanUnusedNamespaces(c context.Context) error {
// 	subquery := `
// 		SELECT obj_ns_id AS id FROM relations
// 		UNION
// 		SELECT sbj_ns_id AS id FROM relations
// 	`
// 	if err := r.db.WithContext(ctx).Where("id NOT IN (" + subquery + ")").
// 		Delete(&model.Namespace{}).Error; err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (r *ZanzibarLogicImpl) cleanUnusedRelationDefs(c context.Context) error {
// 	subquery := `
// 		SELECT sbj_relation_id AS id FROM relations
// 		UNION
// 		SELECT relation_id AS id FROM relations
// 	`
// 	if err := r.db.WithContext(ctx).Where("id NOT IN (" + subquery + ")").
// 		Delete(&model.RelationDef{}).Error; err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (r *ZanzibarLogicImpl) cleanUnusedInstances(c context.Context) error {
// 	subquery := `
// 		SELECT obj_id AS id FROM relations
// 		UNION
// 		SELECT sbj_id AS id FROM relations
// 	`
// 	if err := r.db.WithContext(ctx).Where("id NOT IN (" + subquery + ")").
// 		Delete(&model.Instance{}).Error; err != nil {
// 		return err
// 	}

// 	return nil
// }

func ApplyTupleFilter(f *authzpbv1.TupleFilter) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if f.SbjNs != nil {
			db = db.Where("sbj_ns = ?", *f.SbjNs)
		}
		if f.SbjId != nil {
			db = db.Where("sbj_id = ?", *f.SbjId)
		}
		if f.Rel != nil {
			db = db.Where("relation = ?", *f.Rel)
		}
		if f.ObjNs != nil {
			db = db.Where("obj_ns = ?", *f.ObjNs)
		}
		if f.ObjId != nil {
			db = db.Where("obj_id = ?", *f.ObjId)
		}

		return db
	}
}

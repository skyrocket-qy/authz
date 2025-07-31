package logic

import (
	"authz/internal/entity"
	"authz/internal/entity/model"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron"
	"github.com/skyrocket-qy/gox/logx"
	authzpbv1 "github.com/skyrocket-qy/protos/gen/authzpb/v1"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ZanzibarLogic interface {
	Create(c context.Context, rel *entity.Tuple) error
	Find(c context.Context, filter *entity.Tuple, exact bool) ([]*entity.Tuple, error)
	Delete(c context.Context, filter *entity.Tuple) error

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

// TODO: pagination and limit , sorter
func (r *ZanzibarLogicImpl) Find(c context.Context, filter *entity.TupleFilter) (
	[]*model.Tuple, error,
) {
	tuples := []*model.Tuple{}
	if err := r.db(c).Scopes(filter.Apply()).Find(&tuples).Error; err != nil {
		return nil, err
	}

	return tuples, nil
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

func (r *ZanzibarLogicImpl) Delete(c context.Context, filter *entity.TupleFilter) error {
	return r.db(c).Transaction(func(tx *gorm.DB) error {
		var toDelete []model.Tuple

		if err := tx.Scopes(filter.Apply()).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Find(&toDelete).Error; err != nil {
			return err
		}

		var changelogs []model.ChangeLog
		for _, tuple := range toDelete {
			tupleProto := &authzpbv1.Tuple{
				SbjNs: tuple.SbjNs,
				SbjId: tuple.SbjId,
				Rel:   tuple.Relation,
				ObjNs: tuple.ObjNs,
				ObjId: tuple.ObjId,
			}

			data, err := proto.Marshal(tupleProto)
			if err != nil {
				return err
			}

			changelogs = append(changelogs, model.ChangeLog{Tuple: data})
		}

		if err := tx.Scopes(filter.Apply()).Unscoped().Delete(&model.Tuple{}).Error; err != nil {
			return err
		}

		if len(changelogs) > 0 {
			if err := tx.Create(&changelogs).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *ZanzibarLogicImpl) Check(c context.Context, user *entity.Instance, perm string,
	obj *entity.Instance) (bool, error,
) {
	return ok, nil
}

// What are all the objs that sbj has rel on
func (r *ZanzibarLogicImpl) Lookup(c context.Context, user *entity.Instance, perm string) (
	objs []*entity.Instance, err error,
) {

	return objs, nil
}

func StartCleanJob(r *ZanzibarLogicImpl, cro *cron.Cron) error {
	return cro.AddFunc("0 0 * * *", func() {
		for {
			done, err := r.CleanUnused()
			if err != nil {
				logx.Error(err.Error())
				time.Sleep(15 * time.Minute)
				continue
			}

			if done {
				break
			}

			time.Sleep(15 * time.Minute) // TODO: retry exponentially
		}
	})
}

func (r *ZanzibarLogicImpl) Expand(c context.Context, perm string, obj *entity.Instance) (
	users []entity.Instance, err error,
) {
	return results, nil
}

func (r *ZanzibarLogicImpl) CleanUnused() (done bool, err error) {
	ctx := context.Background()

	lockKey := "zanzibar:clean-daily:lock"
	lockValue := uuid.New().String()
	ok, err := r.rdb.SetNX(ctx, lockKey, lockValue, 24*time.Hour).Result()
	if err != nil {
		return false, err
	}
	if !ok {
		logx.Info("Another clean job is already running.")
		return false, nil
	}
	defer func() {
		script := `
			if redis.call("get", KEYS[1]) == ARGV[1] then
				return redis.call("del", KEYS[1])
			else
				return 0
			end
		`
		if _, err := r.rdb.Eval(ctx, script, []string{lockKey}, lockValue).Result(); err != nil {
			logx.Errorf("Failed to release lock: %v", err)
		}
	}()

	val, err := r.rdb.Get(ctx, "zanzibar:clean-daily:done").Result()
	if err != nil && err != redis.Nil {
		return false, err
	}
	if val != "" {
		logx.Info("Clean job already completed.")
		return true, nil
	}

	if err := r.cleanUnusedNamespaces(ctx); err != nil {
		return false, err
	}
	if err := r.cleanUnusedRelationDefs(ctx); err != nil {
		return false, err
	}
	if err := r.cleanUnusedInstances(ctx); err != nil {
		return false, err
	}

	if err := r.rdb.Set(ctx, "zanzibar:clean-daily:done",
		time.Now().Format(time.RFC3339), 24*time.Hour).Err(); err != nil {
		return true, err
	}

	return true, nil
}

func (r *ZanzibarLogicImpl) cleanUnusedNamespaces(ctx context.Context) error {
	subquery := `
		SELECT obj_ns_id AS id FROM relations
		UNION
		SELECT sbj_ns_id AS id FROM relations
	`
	if err := r.db.WithContext(ctx).Where("id NOT IN (" + subquery + ")").
		Delete(&model.Namespace{}).Error; err != nil {
		return err
	}

	return nil
}

func (r *ZanzibarLogicImpl) cleanUnusedRelationDefs(ctx context.Context) error {
	subquery := `
		SELECT sbj_relation_id AS id FROM relations
		UNION
		SELECT relation_id AS id FROM relations
	`
	if err := r.db.WithContext(ctx).Where("id NOT IN (" + subquery + ")").
		Delete(&model.RelationDef{}).Error; err != nil {
		return err
	}

	return nil
}

func (r *ZanzibarLogicImpl) cleanUnusedInstances(ctx context.Context) error {
	subquery := `
		SELECT obj_id AS id FROM relations
		UNION
		SELECT sbj_id AS id FROM relations
	`
	if err := r.db.WithContext(ctx).Where("id NOT IN (" + subquery + ")").
		Delete(&model.Instance{}).Error; err != nil {
		return err
	}

	return nil
}

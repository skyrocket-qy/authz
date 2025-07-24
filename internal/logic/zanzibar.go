package logic

import (
	"authz/internal/model"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron"
	"github.com/skyrocket-qy/gox/logx"
	authzpbv1 "github.com/skyrocket-qy/protos/gen/authzpb/v1"
	"gorm.io/gorm"
)

// file#read@(user:*#write)
// file#write@file#read

type ZanaibarLogic interface {
	Create(c context.Context, rel *authzpbv1.Tuple) error
	Find(c context.Context, filter *authzpbv1.Tuple, exact bool) ([]*authzpbv1.Tuple, error)
	Delete(c context.Context, filter *authzpbv1.Tuple, exact bool) error
	Check(c context.Context, filter *authzpbv1.Tuple) (bool, error)
	Lookup(c context.Context, sbj *authzpbv1.Sbj, rel string) ([]*authzpbv1.Obj, error)
	Expand(c context.Context, relation string, obj *authzpbv1.Obj) ([]authzpbv1.Sbj, error)
	Tree(c context.Context, sbj *authzpbv1.Sbj, maxDepth ...int) (*authzpbv1.TreeNode, error)

	// FindBySubject(c context.Context, subjNsID, sbj, rel string) ([]Relation, error)
	// FindByObject(c context.Context, objNsID, obj, rel string) ([]Relation, error)
	// Exists(c context.Context, rel *Relation) (bool, error)
	// BulkInsert(c context.Context, rels []Relation) error
	// DeleteBySubject(c context.Context, subjNs, sbj string) error
	// DeleteByObject(c context.Context, objNs, obj string) error
}

var _ ZanaibarLogic = (*ZanzibarLogicImpl)(nil)

type ZanzibarLogicImpl struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewZanzibarLogic(db *gorm.DB) *ZanzibarLogicImpl {
	return &ZanzibarLogicImpl{
		db: db,
	}
}

type CreateEdgeIn struct {
}

func (r *ZanzibarLogicImpl) Create(c context.Context, relation **authzpbv1.Tuple) error {
	var sbjNs, objNs model.Namespace
	var sbj, obj model.Instance
	var rel, sbjRel model.RelationDef

	if err := r.db.WithContext(c).Where("name = ?", relation.SbjNs).FirstOrCreate(&sbjNs).Error; err != nil {
		return err
	}
	if err := r.db.Where("name = ?", relation.ObjNs).FirstOrCreate(&objNs).Error; err != nil {
		return err
	}
	if err := r.db.Where("name = ?", relation.Sbj).FirstOrCreate(&sbj).Error; err != nil {
		return err
	}
	if err := r.db.Where("name = ?", relation.Obj).FirstOrCreate(&obj).Error; err != nil {
		return err
	}
	if err := r.db.Where("name = ?", relation.Relation).FirstOrCreate(&rel).Error; err != nil {
		return err
	}
	if relation.SbjRelation != "" {
		if err := r.db.Where("name = ?", relation.SbjRelation).FirstOrCreate(&sbjRel).Error; err != nil {
			return err
		}
	}

	tuple := model.Tuple{
		SbjNsId:    sbjNs.Id,
		SbjId:      sbj.Id,
		RelationId: rel.Id,
		ObjNsId:    objNs.Id,
		ObjId:      obj.Id,
	}

	if relation.SbjRelation != "" {
		tuple.SbjRelationId = &sbjRel.Id
	}

	return r.db.Create(&tuple).Error
}

func (r *ZanzibarLogicImpl) Delete(c context.Context, relation **authzpbv1.Tuple) error {
	var sbjNs, objNs model.Namespace
	var sbj, obj model.Instance
	var rel, sbjRel model.RelationDef

	if err := r.db.WithContext(c).Where("name = ?", relation.SbjNs).Take(&sbjNs).Error; err != nil {
		return err
	}
	if err := r.db.Where("name = ?", relation.ObjNs).Take(&objNs).Error; err != nil {
		return err
	}
	if err := r.db.Where("name = ?", relation.Sbj).Take(&sbj).Error; err != nil {
		return err
	}
	if err := r.db.Where("name = ?", relation.Obj).Take(&obj).Error; err != nil {
		return err
	}
	if err := r.db.Where("name = ?", relation.Relation).Take(&rel).Error; err != nil {
		return err
	}
	if relation.SbjRelation != "" {
		if err := r.db.Where("name = ?", relation.SbjRelation).Take(&sbjRel).Error; err != nil {
			return err
		}
	}

	tuple := model.Tuple{
		SbjNsId:    sbjNs.Id,
		SbjId:      sbj.Id,
		RelationId: rel.Id,
		ObjNsId:    objNs.Id,
		ObjId:      obj.Id,
	}

	if relation.SbjRelation != "" {
		tuple.SbjRelationId = &sbjRel.Id
	}

	return r.db.Delete(&tuple).Error
}

func (r *ZanzibarLogicImpl) Check()

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

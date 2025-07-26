package logic

import (
	"authz/internal/entity"
	"authz/internal/entity/model"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron"
	"github.com/skyrocket-qy/erx"
	"github.com/skyrocket-qy/gox/logx"
	"gorm.io/gorm"
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
	db  *gorm.DB
	rdb *redis.Client
}

func NewZanzibarLogic(db *gorm.DB, rdb *redis.Client) *ZanzibarLogicImpl {
	return &ZanzibarLogicImpl{
		db:  db,
		rdb: rdb,
	}
}

func (r *ZanzibarLogicImpl) Find(c context.Context, filter *entity.Tuple, exact bool) (
	[]*entity.Tuple, error,
) {
	filterModel, err := filter.ToFilterModel(c, r.db)
	if err != nil {
		return nil, erx.W(err)
	}

	if exact {
		var tuple entity.Tuple
		if err := r.db.WithContext(c).Scopes(filterModel.ToQuery()).Take(&tuple).Error; err != nil {
			return nil, err
		}
		return []*entity.Tuple{&tuple}, nil
	} else {
		var tuples []*entity.Tuple
		if err := r.db.WithContext(c).Where(filterModel).Find(&tuples).Error; err != nil {
			return nil, erx.W(err)
		}
		return tuples, nil
	}
}

func (r *ZanzibarLogicImpl) Create(c context.Context, tuple *entity.Tuple) error {
	var sbjNs, objNs model.Namespace
	var sbj, obj model.Instance
	var rel, sbjRel model.RelationDef

	if err := r.db.WithContext(c).Where("name = ?", tuple.Sbj.Ns).FirstOrCreate(&sbjNs).Error; err != nil {
		return err
	}
	if err := r.db.Where("name = ?", tuple.Obj.Ns).FirstOrCreate(&objNs).Error; err != nil {
		return err
	}
	if err := r.db.Where("name = ?", tuple.Sbj.Id).FirstOrCreate(&sbj).Error; err != nil {
		return err
	}
	if err := r.db.Where("name = ?", tuple.Obj.Id).FirstOrCreate(&obj).Error; err != nil {
		return err
	}
	if err := r.db.Where("name = ?", tuple.Rel).FirstOrCreate(&rel).Error; err != nil {
		return err
	}
	if *tuple.SbjRel != "" {
		if err := r.db.Where("name = ?", tuple.SbjRel).FirstOrCreate(&sbjRel).Error; err != nil {
			return err
		}
	}

	tupleModel := model.Tuple{
		SbjNsId:    sbjNs.Id,
		SbjId:      sbj.Id,
		RelationId: rel.Id,
		ObjNsId:    objNs.Id,
		ObjId:      obj.Id,
	}

	if *tuple.SbjRel != "" {
		tupleModel.SbjRelationId = &sbjRel.Id
	}

	return r.db.Create(&tupleModel).Error
}

func (r *ZanzibarLogicImpl) Delete(c context.Context, tuple *entity.Tuple) error {
	var sbjNs, objNs model.Namespace
	var sbj, obj model.Instance
	var rel, sbjRel model.RelationDef

	if err := r.db.WithContext(c).Where("name = ?", tuple.Sbj.Ns).Take(&sbjNs).Error; err != nil {
		return err
	}
	if err := r.db.Where("name = ?", tuple.Obj.Ns).Take(&objNs).Error; err != nil {
		return err
	}
	if err := r.db.Where("name = ?", tuple.Sbj.Id).Take(&sbj).Error; err != nil {
		return err
	}
	if err := r.db.Where("name = ?", tuple.Obj.Id).Take(&obj).Error; err != nil {
		return err
	}
	if err := r.db.Where("name = ?", tuple.Rel).Take(&rel).Error; err != nil {
		return err
	}
	if *tuple.SbjRel != "" {
		if err := r.db.Where("name = ?", tuple.SbjRel).Take(&sbjRel).Error; err != nil {
			return err
		}
	}

	tupleModel := model.Tuple{
		SbjNsId:    sbjNs.Id,
		SbjId:      sbj.Id,
		RelationId: rel.Id,
		ObjNsId:    objNs.Id,
		ObjId:      obj.Id,
	}

	if *tuple.SbjRel != "" {
		tupleModel.SbjRelationId = &sbjRel.Id
	}

	return r.db.Unscoped().Delete(&tupleModel).Error
}

func (r *ZanzibarLogicImpl) Check(c context.Context, user *entity.Instance, perm string,
	obj *entity.Instance) (bool, error,
) {
	var ok bool
	if err := r.db.WithContext(c).
		Raw("Check(?, ?, ?, ?, ?)", user.Ns, user.Id, perm, obj.Ns, obj.Id).
		Scan(&ok).Error; err != nil {
		return false, err
	}

	return ok, nil
}

// What are all the objs that sbj has rel on
func (r *ZanzibarLogicImpl) Lookup(c context.Context, user *entity.Instance, perm string) (
	objs []*entity.Instance, err error,
) {
	var results []struct {
		ObjNs   string
		ObjName string
	}

	// Call the stored procedure, mapping subject namespace, subject name, subject relation, relation
	err = r.db.Raw(`
        SELECT obj_ns, obj_name
        FROM lookup_objects_by_subject_and_relation(?, ?, ?)
    `, user.Ns, user.Id, perm).Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Map results into your entity.Obj slice
	objs = make([]*entity.Instance, 0, len(results))
	for _, res := range results {
		objs = append(objs, &entity.Instance{
			Ns: res.ObjNs,
			Id: res.ObjName,
		})
	}

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
	var results []entity.Instance

	err = r.db.WithContext(c).
		Raw(`
			SELECT ns, name
			FROM expand_permission_by_name(?, ?, ?)
		`, obj.Ns, obj.Id, perm).
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to expand permission: %w", err)
	}

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

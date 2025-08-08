package engine

import (
	"authz/internal/entity"
	"authz/internal/entity/model"
	"context"
	"fmt"
	"sync"

	"authz/internal/schema"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

type ZanzibarEngine interface {
	Check(c context.Context, sbj *entity.Instance, rel string, obj *entity.Instance) (bool, error)
	Lookup(c context.Context, sbj *entity.Instance, rel string) ([]*entity.Instance, error)
	Expand(c context.Context, rel string, obj *entity.Instance) ([]*entity.Instance, error)
}

var _ ZanzibarEngine = (*ZanzibarEngineImpl)(nil)

type ZanzibarEngineImpl struct {
	schema *schema.Schema
	kafkaR *kafka.Reader
	db     *gorm.DB
	rds    *redis.Client
	graph  map[entity.Instance]map[string]map[entity.Instance]struct{}

	instanceId map[entity.Instance]uint64
	relId      map[string]uint64
	idInstance map[uint64]entity.Instance
	idRel      map[uint64]string
	permId     map[string]uint64

	mutex sync.RWMutex
}

func NewZanzibarEngine(c context.Context, db *gorm.DB, kafkaR *kafka.Reader, rds *redis.Client) (
	*ZanzibarEngineImpl, error,
) {
	engine := ZanzibarEngineImpl{
		kafkaR:     kafkaR,
		db:         db,
		rds:        rds,
		graph:      make(map[entity.Instance]map[string]map[entity.Instance]struct{}),
		instanceId: make(map[entity.Instance]uint64),
		relId:      make(map[string]uint64),
	}

	if err := engine.build(c); err != nil {
		return nil, err
	}

	return &engine, nil
}

func (e *ZanzibarEngineImpl) Check(c context.Context, user *entity.Instance, perm string,
	obj *entity.Instance) (bool, error,
) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	if e.hasDirectTuple(user, perm, obj) {
		return true, nil
	}

	ns, ok := e.schema.Namespaces[obj.Ns]
	if !ok {
		return false, fmt.Errorf("unknown namespace: %s", obj.Ns)
	}
	relation, ok := ns.Relations[perm]
	if !ok {
		return false, fmt.Errorf("unknown relation: %s", perm)
	}

	if relation == nil {
		return false, nil
	}

	return e.evalUsersetRewrite(c, &relation.UsersetRewrite, user, obj), nil
}

func (e *ZanzibarEngineImpl) evalUsersetRewrite(c context.Context, rewrite *schema.UsersetRewrite,
	user *entity.Instance, obj *entity.Instance) bool {
	switch {

	case rewrite.ComputedUserSet != nil:
		return e.hasDirectTuple(user, rewrite.ComputedUserSet.Relation, obj)

	case rewrite.TupleToUserset != nil:
		// Find tuples that point to intermediate objects
		relSbj, ok := e.graph[*obj]
		if !ok {
			return false
		}
		sbjs, ok := relSbj[rewrite.TupleToUserset.Tupleset.Relation]
		if !ok {
			return false
		}

		for sbj := range sbjs {
			ok, _ := e.Check(c, user, rewrite.TupleToUserset.ComputedUserset.Relation, &sbj)
			if ok {
				return true
			}
		}

		return false

	case rewrite.Union != nil:
		for _, r := range rewrite.Union {
			if e.evalUsersetRewrite(c, r, user, obj) {
				return true
			}
		}
		return false

	case rewrite.Intersection != nil:
		for _, r := range rewrite.Intersection {
			if !e.evalUsersetRewrite(c, r, user, obj) {
				return false
			}
		}
		return true

	case rewrite.Exclusion != nil:
		return e.evalUsersetRewrite(c, rewrite.Exclusion.Base, user, obj) &&
			!e.evalUsersetRewrite(c, rewrite.Exclusion.Subtract, user, obj)
	}

	return false
}

func (e *ZanzibarEngineImpl) hasDirectTuple(user *entity.Instance, rel string, obj *entity.Instance) bool {
	if _, ok := e.graph[*obj][rel][*user]; ok {
		return true
	}
	return false
}

func (e *ZanzibarEngineImpl) Lookup(c context.Context, sbj *entity.Instance, rel string) (
	[]*entity.Instance, error) {
	return nil, nil
}

func (e *ZanzibarEngineImpl) Expand(c context.Context, perm string, obj *entity.Instance) (
	sbjs []*entity.Instance, err error,
) {

	return nil, nil
}

func (e *ZanzibarEngineImpl) build(c context.Context) error {
	tuples := []model.Tuple{}
	if err := e.db.WithContext(c).Find(&tuples).Error; err != nil {
		return err
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	for _, tuple := range tuples {
		obj := entity.Instance{
			Ns: tuple.ObjNs,
			Id: tuple.ObjId,
		}
		sbj := entity.Instance{
			Ns: tuple.SbjNs,
			Id: tuple.SbjId,
		}
		rel := tuple.Relation

		if _, exists := e.graph[obj]; !exists {
			e.graph[obj] = make(map[string]map[entity.Instance]struct{})
		}
		if _, exists := e.graph[obj][rel]; !exists {
			e.graph[obj][rel] = make(map[entity.Instance]struct{})
		}
		e.graph[obj][rel][sbj] = struct{}{}
	}

	return nil
}

package engine

import (
	"authz/internal/entity"
	"authz/internal/entity/model"
	"context"
	"errors"
	"fmt"
	"sync"

	"authz/internal/schema"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

type ZanzibarEngine interface {
	Check(c context.Context, sbj *entity.Instance, rel string, obj *entity.Instance) (bool, error)
	// Lookup(c context.Context, sbj *entity.Instance, rel string) ([]*entity.Instance, error)
	// Expand(c context.Context, rel string, obj *entity.Instance) ([]*entity.Instance, error)
}

var _ ZanzibarEngine = (*ZanzibarEngineImpl)(nil)

type ZanzibarEngineImpl struct {
	schema *schema.Schema
	kafkaR *kafka.Reader
	db     *gorm.DB
	rds    *redis.Client
	graph  map[string]map[string]map[string]struct{}

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
		kafkaR: kafkaR,
		db:     db,
		rds:    rds,
		// graph:      make(map[uint64]map[uint64]map[uint64]struct{}),
		instanceId: make(map[entity.Instance]uint64),
		relId:      make(map[string]uint64),
	}

	if err := engine.build(c); err != nil {
		return nil, err
	}

	return &engine, nil
}

func (e *ZanzibarEngineImpl) checkRel(c context.Context, sbj *entity.Instance, rel string,
	obj *entity.Instance) (
	bool, error,
) {

	return false, nil
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
	rewrite, ok := ns.Relations[perm]
	if !ok {
		return false, fmt.Errorf("unknown relation: %s", perm)
	}

	if rewrite == nil {
		return false, nil
	}

	return e.evalUsersetRewrite(rewrite, user, obj), nil
}

func (e *ZanzibarEngineImpl) evalUsersetRewrite(c context.Context, rewrite *schema.UsersetRewrite,
	user *entity.Instance, obj *entity.Instance) bool {
	if rewrite.ComputedUserSet != nil {
		return e.hasDirectTuple(user, rewrite.ComputedUserSet.Relation, obj)
	}

	if rewrite.TupleToUserset != nil {
		// Find tuples that point to intermediate objects
		var intermediate []entity.Instance
		relSbj, ok := e.graph[obj.Ns+":"+obj.Id]
		if !ok {
			return false
		}
		sbj, ok := relSbj[rewrite.TupleToUserset.Tupleset.Relation]
		if !ok {
			return false
		}

		intermediate = append(intermediate, sbj...)

		// For each intermediate, follow ComputedUserset
		for _, inst := range intermediate {
			ok, _ := e.Check(context.TODO(), user, rewrite.TupleToUserset.ComputedUserset.Relation, &inst)
			if ok {
				return true
			}
		}
		return false
	}

	if rewrite.Union != nil {
		for _, r := range rewrite.Union {
			if e.evalUsersetRewrite(r, user, obj) {
				return true
			}
		}
		return false
	}

	if rewrite.Intersection != nil {
		for _, r := range rewrite.Intersection {
			if !e.evalUsersetRewrite(r, user, obj) {
				return false
			}
		}
		return true
	}

	if rewrite.Exclusion != nil {
		return e.evalUsersetRewrite(rewrite.Exclusion.Base, user, obj) &&
			!e.evalUsersetRewrite(rewrite.Exclusion.Subtract, user, obj)
	}

	return false
}

func (e *ZanzibarEngineImpl) hasDirectTuple(user *entity.Instance, rel string, obj *entity.Instance) bool {
	objS := obj.Ns + ":" + obj.Id
	sbj := user.Ns + ":" + user.Id
	if _, ok := e.graph[objS][rel][sbj]; ok {
		return true
	}
	return false
}

func (e *ZanzibarEngineImpl) eval(c context.Context, sbj, obj *entity.Instance,
	expr *schema.PermissionExpr) (bool, error,
) {
	switch {
	case expr.Relation != "":
		return e.checkRel(c, sbj, expr.Relation, obj)

	case expr.Or != nil:
		for _, sub := range expr.Or {
			ok, err := e.evalExpr(c, sbj, obj, sub)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
		return false, nil

	case expr.And != nil:
		for _, sub := range expr.And {
			ok, err := e.evalExpr(c, sbj, obj, sub)
			if err != nil {
				return false, err
			}
			if !ok {
				return false, nil
			}
		}
		return true, nil

	case expr.Not != nil:
		if len(expr.Not) == 0 {
			return false, nil
		}
		ok, err := e.evalExpr(c, sbj, obj, expr.Not[0])
		if err != nil || !ok {
			return false, err
		}
		for _, sub := range expr.Not[1:] {
			ok, err := e.evalExpr(c, sbj, obj, sub)
			if err != nil {
				return false, err
			}
			if ok {
				return false, nil
			}
		}
		return true, nil

	default:
		return false, errors.New("invalid permission expression")
	}
}

func (e *ZanzibarEngineImpl) lookup(c context.Context, sbj *entity.Instance, perm string) (
	[]*entity.Instance, error,
) {
	return nil, nil
}

func (e *ZanzibarEngineImpl) expand(c context.Context, perm string, obj *entity.Instance) (
	sbjs []*entity.Instance, err error,
) {

	return nil, nil
}

func (e *ZanzibarEngineImpl) build(c context.Context) error {
	tuples := []model.Tuple{}
	if err := e.db.WithContext(c).Find(&tuples).Error; err != nil {
		return err
	}

	// vertices := []model.Vertex{}
	// if err := e.db.WithContext(c).Find(&vertices).Error; err != nil {
	// 	return err
	// }

	// edges := []model.Edge{}
	// if err := e.db.WithContext(c).Find(&edges).Error; err != nil {
	// 	return err
	// }

	e.mutex.Lock()
	defer e.mutex.Unlock()

	for _, tuple := range tuples {
		obj := tuple.ObjNs + ":" + tuple.ObjId
		sbj := tuple.SbjNs + ":" + tuple.SbjId
		rel := tuple.Relation

		if _, exists := e.graph[obj]; !exists {
			e.graph[obj] = make(map[string]map[string]struct{})
		}
		if _, exists := e.graph[obj][rel]; !exists {
			e.graph[obj][rel] = make(map[string]struct{})
		}
		e.graph[obj][rel][sbj] = struct{}{}
	}

	// for _, vertex := range vertices {
	// 	key := entity.Instance{
	// 		Ns: vertex.Namespace,
	// 		Id: vertex.Name,
	// 	}

	// 	if vertex.Relation != nil {
	// 		key.Rel = *vertex.Relation
	// 	}
	// 	e.instanceId[key] = vertex.Id
	// }

	// for _, edge := range edges {
	// 	e.relId[edge.Relation] = edge.Id
	// }

	return nil
}

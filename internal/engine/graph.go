package engine

import (
	"authz/internal/entity"
	"context"
	"errors"
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
	graph  map[entity.Instance]map[uint64]map[entity.Instance]struct{}

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
	visited := make(map[uint64]struct{})
	queue := []uint64{sbjId}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if _, seen := visited[current]; seen {
			continue
		}
		visited[current] = struct{}{}

		relMap, exists := e.graph[current]
		if !exists {
			continue
		}

		objSet, exists := relMap[relId]
		if !exists {
			continue
		}

		for next := range objSet {
			if next == objId {
				return true, nil
			}

			objInst, ok := e.idInstance[next]
			if !ok {
				continue // or log warning
			}
			objInst.Rel, ok = e.idRel[relId]
			if !ok {
				continue
			}
			nextId, ok := e.instanceId[objInst]
			if !ok {
				continue
			}
			queue = append(queue, nextId)
		}
	}

	return false, nil
}

func (e *ZanzibarEngineImpl) Check(c context.Context, sbj *entity.Instance, perm string,
	obj *entity.Instance) (bool, error,
) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	sbjs, err := e.expand(c, perm, obj)
	if err != nil {
		return false, err
	}

	for _, s := range sbjs {
		if *s == *sbj {
			return true, nil
		}
	}

	return false, nil
}

func (e *ZanzibarEngineImpl) evalExpr(c context.Context, sbj, obj *entity.Instance,
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
	// tuples := []model.Tuple{}
	// if err := e.db.WithContext(c).Find(&tuples).Error; err != nil {
	// 	return err
	// }

	// vertices := []model.Vertex{}
	// if err := e.db.WithContext(c).Find(&vertices).Error; err != nil {
	// 	return err
	// }

	// edges := []model.Edge{}
	// if err := e.db.WithContext(c).Find(&edges).Error; err != nil {
	// 	return err
	// }

	// e.mutex.Lock()
	// defer e.mutex.Unlock()

	// for _, tuple := range tuples {
	// 	if _, exists := e.graph[tuple.SbjId]; !exists {
	// 		e.graph[tuple.SbjId] = make(map[uint64]map[uint64]struct{})
	// 	}
	// 	if _, exists := e.graph[tuple.SbjId][tuple.RelationId]; !exists {
	// 		e.graph[tuple.SbjId][tuple.RelationId] = make(map[uint64]struct{})
	// 	}
	// 	e.graph[tuple.SbjId][tuple.RelationId][tuple.ObjId] = struct{}{}
	// }

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

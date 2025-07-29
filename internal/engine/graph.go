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
	schema     *schema.Schema
	kafkaR     *kafka.Reader
	db         *gorm.DB
	rds        *redis.Client
	graph      map[uint64]map[uint64]map[uint64]struct{}
	instanceId map[entity.Instance]uint64
	relId      map[string]uint64
	mutex      sync.RWMutex
}

func NewZanzibarEngine(c context.Context, db *gorm.DB, kafkaR *kafka.Reader, rds *redis.Client) (
	*ZanzibarEngineImpl, error,
) {
	engine := ZanzibarEngineImpl{
		kafkaR:     kafkaR,
		db:         db,
		rds:        rds,
		graph:      make(map[uint64]map[uint64]map[uint64]struct{}),
		instanceId: make(map[entity.Instance]uint64),
		relId:      make(map[string]uint64),
	}

	if err := engine.build(c); err != nil {
		return nil, err
	}

	return &engine, nil
}

func (e *ZanzibarEngineImpl) Check(c context.Context, sbj *entity.Instance, rel string,
	obj *entity.Instance) (bool, error,
) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	startID, ok := e.instanceId[*sbj]
	if !ok {
		return false, nil
	}
	targetID, ok := e.instanceId[*obj]
	if !ok {
		return false, nil
	}
	relID, ok := e.relId[rel]
	if !ok {
		return false, nil
	}

	nsSchema, ok := e.schema.Namespaces[obj.Ns]
	if !ok {
		return false, fmt.Errorf("namespace %s not found in schema", obj.Ns)
	}

	// Relation Check (direct)
	if _, ok := nsSchema.Relations[rel]; ok {
		if nexts, ok := e.graph[startID][relID]; ok {
			if _, exists := nexts[targetID]; exists {
				return true, nil
			}
		}
		return false, nil
	}

	// Permission Check (recursive)
	if permDef, ok := nsSchema.Permissions[rel]; ok {
		// Single relation
		if permDef.Relation != "" {
			return e.Check(c, sbj, permDef.Relation, obj)
		}

		// Union: any one true
		for _, subRel := range permDef.Union {
			ok, err := e.Check(c, sbj, subRel, obj)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}

		// Intersection: all must be true
		allTrue := true
		for _, subRel := range permDef.Intersection {
			ok, err := e.Check(c, sbj, subRel, obj)
			if err != nil || !ok {
				allTrue = false
				break
			}
		}
		if len(permDef.Intersection) > 0 {
			return allTrue, nil
		}

		// Exclusion: first true AND all rest false
		if len(permDef.Exclusion) > 0 {
			head := permDef.Exclusion[0]
			tail := permDef.Exclusion[1:]

			headOK, err := e.Check(c, sbj, head, obj)
			if err != nil || !headOK {
				return false, nil
			}
			for _, subRel := range tail {
				tailOK, err := e.Check(c, sbj, subRel, obj)
				if err != nil {
					return false, err
				}
				if tailOK {
					return false, nil
				}
			}
			return true, nil
		}
	}

	return false, nil
}

func (e *ZanzibarEngineImpl) Lookup(c context.Context, sbj *entity.Instance, rel string) (
	[]*entity.Instance, error,
) {
	return nil, nil
}

func (e *ZanzibarEngineImpl) Expand(c context.Context, rel string, obj *entity.Instance) (
	[]*entity.Instance, error,
) {
	return nil, nil
}

func (e *ZanzibarEngineImpl) build(c context.Context) error {
	tuples := []model.Tuple{}
	if err := e.db.WithContext(c).Find(&tuples).Error; err != nil {
		return err
	}

	vertices := []model.Vertex{}
	if err := e.db.WithContext(c).Find(&vertices).Error; err != nil {
		return err
	}

	edges := []model.Edge{}
	if err := e.db.WithContext(c).Find(&edges).Error; err != nil {
		return err
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	for _, tuple := range tuples {
		if _, exists := e.graph[tuple.SbjId]; !exists {
			e.graph[tuple.SbjId] = make(map[uint64]map[uint64]struct{})
		}
		if _, exists := e.graph[tuple.SbjId][tuple.RelationId]; !exists {
			e.graph[tuple.SbjId][tuple.RelationId] = make(map[uint64]struct{})
		}
		e.graph[tuple.SbjId][tuple.RelationId][tuple.ObjId] = struct{}{}
	}

	for _, vertex := range vertices {
		key := entity.Instance{
			Ns: vertex.Namespace,
			Id: vertex.Name,
		}

		if vertex.Relation != nil {
			key.Rel = *vertex.Relation
		}
		e.instanceId[key] = vertex.Id
	}

	for _, edge := range edges {
		e.relId[edge.Relation] = edge.Id
	}

	return nil
}

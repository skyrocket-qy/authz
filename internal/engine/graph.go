package engine

import (
	"authz/internal/entity"
	"context"
)

type ZanzibarEngine interface {
	Check(c context.Context, sbj *entity.Instance, rel string, obj *entity.Instance) (bool, error)
	Lookup(c context.Context, sbj *entity.Instance, rel string) ([]*entity.Instance, error)
	Expand(c context.Context, rel string, obj *entity.Instance) ([]*entity.Instance, error)
}

var _ ZanzibarEngine = (*ZanzibarEngineImpl)(nil)

type ZanzibarEngineImpl struct {
	graph      map[uint64]map[uint64]map[uint64]struct{}
	instanceId map[entity.Instance]uint64
	relId      map[string]uint64
}

func NewZanzibarEngine() *ZanzibarEngineImpl {
	return &ZanzibarEngineImpl{}
}

func (e *ZanzibarEngineImpl) Check(c context.Context, sbj *entity.Instance, rel string,
	obj *entity.Instance) (bool, error,
) {
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

	visited := make(map[uint64]struct{})
	queue := []uint64{startID}

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

		objSet, exists := relMap[relID]
		if !exists {
			continue
		}

		for next := range objSet {
			if next == targetID {
				return true, nil
			}
			queue = append(queue, next)
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

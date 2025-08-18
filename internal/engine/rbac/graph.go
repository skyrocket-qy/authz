package rbac

import (
	"authz/internal/entity"
	"strconv"
	"sync"
)

const NumShards = 8192

type Graph struct {
	Shards []*Shard
}

func NewGraph() *Graph {
	shards := make([]*Shard, NumShards)
	for i := range NumShards {
		shards[i] = &Shard{Graph: make(map[entity.Instance]*ObjEntry)}
	}
	return &Graph{Shards: shards}
}

type Shard struct {
	Graph map[entity.Instance]*ObjEntry // obj -> ObjEntry
	mu    sync.RWMutex
}

type ObjEntry struct {
	// We don't give each relation a lock because it's too expensive
	Relations map[string]map[entity.Instance]struct{} // rel -> sbj
	mu        sync.RWMutex
}

func (g *Graph) getShard(obj *entity.Instance) *Shard {
	sum, _ := strconv.Atoi(obj.Id) // TODO: use uint64 on obj
	return g.Shards[sum%NumShards]
}

// Read only locks the object
func (g *Graph) Exist(obj *entity.Instance, rel string, sbj *entity.Instance) bool {
	s := g.getShard(obj)
	s.mu.RLock()
	objEntry := s.Graph[*obj]
	s.mu.RUnlock()
	if objEntry == nil {
		return false
	}

	objEntry.mu.RLock()
	defer objEntry.mu.RUnlock()

	relEntry := objEntry.Relations[rel]
	if relEntry == nil {
		return false
	}

	_, ok := relEntry[*sbj]
	return ok
}

// Write only locks the object
func (g *Graph) Create(obj *entity.Instance, rel string, sbj *entity.Instance) {
	s := g.getShard(obj)
	s.mu.RLock()
	objEntry := s.Graph[*obj]
	s.mu.RUnlock()
	if objEntry == nil {
		s.mu.Lock()
		if s.Graph[*obj] == nil {
			s.Graph[*obj] = &ObjEntry{Relations: make(map[string]map[entity.Instance]struct{})}
		}
		objEntry = s.Graph[*obj]
		s.mu.Unlock()
	}

	objEntry.mu.Lock()
	defer objEntry.mu.Unlock()

	relEntry := objEntry.Relations[rel]
	if relEntry == nil {
		objEntry.Relations[rel] = make(map[entity.Instance]struct{})
		relEntry = objEntry.Relations[rel]
	}

	relEntry[*sbj] = struct{}{}
}

// Delete only locks the object
func (g *Graph) Delete(obj *entity.Instance, rel string, sbj *entity.Instance) {
	s := g.getShard(obj)

	s.mu.RLock()
	objEntry := s.Graph[*obj]
	s.mu.RUnlock()
	if objEntry == nil {
		return
	}

	objEntry.mu.Lock()
	defer objEntry.mu.Unlock()
	relEntry := objEntry.Relations[rel]
	if relEntry == nil {
		return
	}

	delete(relEntry, *sbj)
	if len(relEntry) == 0 {
		delete(objEntry.Relations, rel)

		if len(objEntry.Relations) == 0 {
			s.mu.Lock()
			if len(objEntry.Relations) == 0 {
				delete(s.Graph, *obj)
			}
			s.mu.Unlock()
		}
	}
}

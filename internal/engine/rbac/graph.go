package rbac

import (
	"authz/internal/entity"
	"strconv"
	"sync"
)

const NumShards = 8192

type Graph struct {
	shards []*Shard
}

func NewGraph() *Graph {
	shards := make([]*Shard, NumShards)
	for i := range NumShards {
		shards[i] = &Shard{graph: make(map[entity.Instance]*ObjEntry)}
	}
	return &Graph{shards: shards}
}

type Shard struct {
	graph map[entity.Instance]*ObjEntry // obj -> ObjEntry
	mu    sync.RWMutex
}

type ObjEntry struct {
	// We don't give each relation a lock because it's too expensive
	Relations map[string]map[entity.Instance]struct{} // rel -> sbj
	mu        sync.RWMutex
}

func (g *Graph) getShard(obj entity.Instance) *Shard {
	sum, _ := strconv.Atoi(obj.Id) // TODO: use uint64 on obj
	return g.shards[sum%NumShards]
}

// Read only locks the object
func (g *Graph) Exist(obj entity.Instance, rel string, sbj entity.Instance) bool {
	s := g.getShard(obj)
	s.mu.RLock()
	objEntry := s.graph[obj]
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

	_, ok := relEntry[sbj]
	return ok
}

// Write only locks the object
func (g *Graph) Create(obj entity.Instance, rel string, sbj entity.Instance) {
	s := g.getShard(obj)
	s.mu.RLock()
	objEntry := s.graph[obj]
	s.mu.RUnlock()
	if objEntry == nil {
		s.mu.Lock()
		if s.graph[obj] == nil {
			s.graph[obj] = &ObjEntry{Relations: make(map[string]map[entity.Instance]struct{})}
		}
		objEntry = s.graph[obj]
		s.mu.Unlock()
	}

	objEntry.mu.Lock()
	defer objEntry.mu.Unlock()

	relEntry := objEntry.Relations[rel]
	if relEntry == nil {
		objEntry.Relations[rel] = make(map[entity.Instance]struct{})
		relEntry = objEntry.Relations[rel]
	}

	relEntry[sbj] = struct{}{}
}

// Delete only locks the object
func (g *Graph) Delete(obj entity.Instance, rel string, sbj entity.Instance) {
	s := g.getShard(obj)

	s.mu.RLock()
	objEntry := s.graph[obj]
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

	delete(relEntry, sbj)
	if len(relEntry) == 0 {
		delete(objEntry.Relations, rel)

		if len(objEntry.Relations) == 0 {
			s.mu.Lock()
			if len(objEntry.Relations) == 0 {
				delete(s.graph, obj)
			}
			s.mu.Unlock()
		}
	}
}

func (g *Graph) GetSbjs(obj entity.Instance, rel string) map[entity.Instance]struct{} {
	s := g.getShard(obj)
	s.mu.RLock()
	objEntry := s.graph[obj]
	s.mu.RUnlock()
	if objEntry == nil {
		return nil
	}

	objEntry.mu.RLock()
	defer objEntry.mu.RUnlock()

	relEntry := objEntry.Relations[rel]
	if relEntry == nil {
		return nil
	}

	return relEntry
}

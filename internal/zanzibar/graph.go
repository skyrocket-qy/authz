package zanzibar

import (
	"strconv"
	"sync"

	"authz/internal/entity"
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
	Mu    sync.RWMutex
}

type ObjEntry struct {
	// We don't give each relation a lock because it's too expensive
	Relations map[string]map[entity.Instance]struct{} // rel -> sbj
	Mu        sync.RWMutex
}

func (g *Graph) GetShard(obj entity.Instance) *Shard {
	sum, _ := strconv.Atoi(obj.Id) // TODO: use uint64 on obj

	return g.Shards[sum%NumShards]
}

// TODO: wildcard support
// Read only locks the object.
func (g *Graph) Exist(obj entity.Instance, rel string, sbj entity.Instance) bool {
	s := g.GetShard(obj)
	s.Mu.RLock()
	objEntry := s.Graph[obj]
	s.Mu.RUnlock()

	if objEntry == nil {
		return false
	}

	objEntry.Mu.RLock()
	defer objEntry.Mu.RUnlock()

	relEntry := objEntry.Relations[rel]
	if relEntry == nil {
		return false
	}

	_, ok := relEntry[sbj]

	return ok
}

// Write only locks the object.
func (g *Graph) Create(obj entity.Instance, rel string, sbj entity.Instance) {
	s := g.GetShard(obj)
	s.Mu.RLock()
	objEntry := s.Graph[obj]
	s.Mu.RUnlock()

	if objEntry == nil {
		s.Mu.Lock()

		if s.Graph[obj] == nil {
			s.Graph[obj] = &ObjEntry{Relations: make(map[string]map[entity.Instance]struct{})}
		}

		objEntry = s.Graph[obj]
		s.Mu.Unlock()
	}

	objEntry.Mu.Lock()
	defer objEntry.Mu.Unlock()

	relEntry := objEntry.Relations[rel]
	if relEntry == nil {
		objEntry.Relations[rel] = make(map[entity.Instance]struct{})
		relEntry = objEntry.Relations[rel]
	}

	relEntry[sbj] = struct{}{}
}

// Delete only locks the object.
func (g *Graph) Delete(obj entity.Instance, rel string, sbj entity.Instance) {
	s := g.GetShard(obj)

	s.Mu.RLock()
	objEntry := s.Graph[obj]
	s.Mu.RUnlock()

	if objEntry == nil {
		return
	}

	objEntry.Mu.Lock()
	defer objEntry.Mu.Unlock()

	relEntry := objEntry.Relations[rel]
	if relEntry == nil {
		return
	}

	delete(relEntry, sbj)

	if len(relEntry) == 0 {
		delete(objEntry.Relations, rel)

		if len(objEntry.Relations) == 0 {
			s.Mu.Lock()

			if len(objEntry.Relations) == 0 {
				delete(s.Graph, obj)
			}

			s.Mu.Unlock()
		}
	}
}

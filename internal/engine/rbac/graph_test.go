package rbac_test

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"authz/internal/entity"
)

// numShards = min(2*numGoroutines, numObjects)

// very high case.
const (
	numObjects    = 1000000
	numRelations  = 8
	numSubjects   = 100
	readPercent   = 95
	numGoroutines = 500
	numOps        = 500000
	numShards     = 8192
)

// highly content
// const (
// 	numObjects    = 5
// 	numRelations  = 3
// 	numSubjects   = 5
// 	readPercent   = 90
// 	numGoroutines = 400
// 	numOps        = 50000
// 	numShards     = 5
// )

// test delete
// const (
// 	numObjects    = 5_000
// 	numRelations  = 5
// 	numSubjects   = 10
// 	readPercent   = 0
// 	numGoroutines = 50
// 	numOps        = 50_000
// 	numShards     = 16
// 	// Mix of writes: 50% create, 50% delete
// )

// General case
// const (
// 	numObjects    = 5000
// 	numRelations  = 5
// 	numSubjects   = 10
// 	readPercent   = 95
// 	numGoroutines = 50
// 	numOps        = 40000
// 	numShards     = 8192
// )

// -------------------- Per-Object Lock --------------------.
type ObjEntry struct {
	relations map[string]map[entity.Instance]struct{}
	mu        sync.RWMutex
}

type GraphPerObject struct {
	graph map[entity.Instance]*ObjEntry
	mu    sync.RWMutex // protects graph map itself
}

func NewGraphPerObject() *GraphPerObject {
	return &GraphPerObject{graph: make(map[entity.Instance]*ObjEntry)}
}

func (g *GraphPerObject) getObj(obj entity.Instance) *ObjEntry {
	g.mu.RLock()
	entry := g.graph[obj]
	g.mu.RUnlock()

	if entry == nil {
		g.mu.Lock()

		if g.graph[obj] == nil {
			g.graph[obj] = &ObjEntry{relations: make(map[string]map[entity.Instance]struct{})}
		}

		entry = g.graph[obj]
		g.mu.Unlock()
	}

	return entry
}

func (g *GraphPerObject) Read(obj entity.Instance, rel string, sbj entity.Instance) bool {
	entry := g.getObj(obj)

	entry.mu.RLock()
	defer entry.mu.RUnlock()

	return entry.relations[rel][sbj] != struct{}{}
}

func (g *GraphPerObject) Write(obj entity.Instance, rel string, sbj entity.Instance) {
	entry := g.getObj(obj)

	entry.mu.Lock()
	defer entry.mu.Unlock()

	if entry.relations[rel] == nil {
		entry.relations[rel] = make(map[entity.Instance]struct{})
	}

	entry.relations[rel][sbj] = struct{}{}
}

func (g *GraphPerObject) delete(obj entity.Instance, rel string, sbj entity.Instance) {
	// First: acquire the object entry under global lock for safe access
	g.mu.RLock()
	entry, exists := g.graph[obj]
	g.mu.RUnlock()

	if !exists {
		return
	}

	// Lock only the object entry
	entry.mu.Lock()

	sbjMap, ok := entry.relations[rel]
	if ok {
		delete(sbjMap, sbj)

		if len(sbjMap) == 0 {
			delete(entry.relations, rel)
		}
	}

	empty := len(entry.relations) == 0
	entry.mu.Unlock()

	// If object entry is empty, remove it from global map
	if empty {
		g.mu.Lock()
		// Recheck inside write lock to avoid race
		if entry2, ok := g.graph[obj]; ok && entry2 == entry && len(entry2.relations) == 0 {
			delete(g.graph, obj)
		}

		g.mu.Unlock()
	}
}

// -------------------- Object + Relation Lock --------------------.
type RelEntry struct {
	sbjs map[entity.Instance]struct{}
	mu   sync.RWMutex
}

type ObjREntry struct {
	relations map[string]*RelEntry
	mu        sync.RWMutex
}

type GraphPerObjRel struct {
	graph map[entity.Instance]*ObjREntry
	mu    sync.RWMutex // protects graph map itself
}

func NewGraphPerObjRel() *GraphPerObjRel {
	return &GraphPerObjRel{
		graph: make(map[entity.Instance]*ObjREntry),
	}
}

func (g *GraphPerObjRel) getObj(obj entity.Instance) *ObjREntry {
	g.mu.RLock()
	entry := g.graph[obj]
	g.mu.RUnlock()

	if entry == nil {
		g.mu.Lock()

		if g.graph[obj] == nil {
			g.graph[obj] = &ObjREntry{relations: make(map[string]*RelEntry)}
		}

		entry = g.graph[obj]
		g.mu.Unlock()
	}

	return entry
}

func (entry *ObjREntry) getRel(rel string) *RelEntry {
	entry.mu.RLock()
	r := entry.relations[rel]
	entry.mu.RUnlock()

	if r == nil {
		entry.mu.Lock()

		if entry.relations[rel] == nil {
			entry.relations[rel] = &RelEntry{sbjs: make(map[entity.Instance]struct{})}
		}

		r = entry.relations[rel]
		entry.mu.Unlock()
	}

	return r
}

func (g *GraphPerObjRel) Read(obj entity.Instance, rel string, sbj entity.Instance) bool {
	entry := g.getObj(obj)
	rEntry := entry.getRel(rel)

	rEntry.mu.RLock()
	defer rEntry.mu.RUnlock()

	_, exists := rEntry.sbjs[sbj]

	return exists
}

func (g *GraphPerObjRel) Write(obj entity.Instance, rel string, sbj entity.Instance) {
	entry := g.getObj(obj)
	rEntry := entry.getRel(rel)

	rEntry.mu.Lock()
	defer rEntry.mu.Unlock()

	rEntry.sbjs[sbj] = struct{}{}
}

func (g *GraphPerObjRel) delete(obj entity.Instance, rel string, sbj entity.Instance) {
	// Step 1: Find object entry
	g.mu.RLock()
	entry := g.graph[obj]
	g.mu.RUnlock()

	if entry == nil {
		return
	}

	// Step 2: Find relation entry
	entry.mu.RLock()
	rEntry := entry.relations[rel]
	entry.mu.RUnlock()

	if rEntry == nil {
		return
	}

	// Step 3: Delete subject
	rEntry.mu.Lock()
	delete(rEntry.sbjs, sbj)
	emptyRel := len(rEntry.sbjs) == 0
	rEntry.mu.Unlock()

	// Step 4: If relation is empty, delete it from object
	if emptyRel {
		entry.mu.Lock()
		delete(entry.relations, rel)
		emptyObj := len(entry.relations) == 0
		entry.mu.Unlock()

		// Step 5: If object is empty, delete it from graph
		if emptyObj {
			g.mu.Lock()
			delete(g.graph, obj)
			g.mu.Unlock()
		}
	}
}

// -------------------- Sharded Map --------------------.
type RelSDEntry struct {
	sbjs map[entity.Instance]struct{}
	mu   sync.RWMutex
}

type ObjSDEntry struct {
	relations map[string]*RelSDEntry
	mu        sync.RWMutex
}

type shard struct {
	graph map[entity.Instance]*ObjSDEntry
	mu    sync.RWMutex
}

type GraphSharded struct {
	shards []*shard
	n      int
}

func NewGraphSharded(numShards int) *GraphSharded {
	shards := make([]*shard, numShards)
	for i := range numShards {
		shards[i] = &shard{graph: make(map[entity.Instance]*ObjSDEntry)}
	}

	return &GraphSharded{shards: shards, n: numShards}
}

func (g *GraphSharded) getShard(obj entity.Instance) *shard {
	sum, _ := strconv.Atoi(obj.Id)

	return g.shards[sum%g.n]
}

func (g *GraphSharded) getObj(obj entity.Instance) *ObjSDEntry {
	s := g.getShard(obj)
	s.mu.RLock()
	entry := s.graph[obj]
	s.mu.RUnlock()

	if entry == nil {
		s.mu.Lock()

		if s.graph[obj] == nil {
			s.graph[obj] = &ObjSDEntry{relations: make(map[string]*RelSDEntry)}
		}

		entry = s.graph[obj]
		s.mu.Unlock()
	}

	return entry
}

func (entry *ObjSDEntry) getRel(rel string) *RelSDEntry {
	entry.mu.RLock()
	r := entry.relations[rel]
	entry.mu.RUnlock()

	if r == nil {
		entry.mu.Lock()

		if entry.relations[rel] == nil {
			entry.relations[rel] = &RelSDEntry{sbjs: make(map[entity.Instance]struct{})}
		}

		r = entry.relations[rel]
		entry.mu.Unlock()
	}

	return r
}

// Read returns true if the edge exists.
func (g *GraphSharded) Read(obj entity.Instance, rel string, sbj entity.Instance) bool {
	entry := g.getObj(obj)
	rEntry := entry.getRel(rel)

	rEntry.mu.RLock()
	defer rEntry.mu.RUnlock()

	_, exists := rEntry.sbjs[sbj]

	return exists
}

// Write adds an edge.
func (g *GraphSharded) Write(obj entity.Instance, rel string, sbj entity.Instance) {
	entry := g.getObj(obj)
	rEntry := entry.getRel(rel)

	rEntry.mu.Lock()
	defer rEntry.mu.Unlock()

	rEntry.sbjs[sbj] = struct{}{}
}

func (g *GraphSharded) delete(obj entity.Instance, rel string, sbj entity.Instance) {
	// Step 1: Get the shard for this object
	s := g.getShard(obj)

	// Step 2: Get the object entry
	s.mu.RLock()
	entry := s.graph[obj]
	s.mu.RUnlock()

	if entry == nil {
		return
	}

	// Step 3: Get the relation entry
	entry.mu.RLock()
	rEntry := entry.relations[rel]
	entry.mu.RUnlock()

	if rEntry == nil {
		return
	}

	// Step 4: Remove the subject from the relation
	rEntry.mu.Lock()
	delete(rEntry.sbjs, sbj)
	emptyRel := len(rEntry.sbjs) == 0
	rEntry.mu.Unlock()

	// Step 5: If the relation is now empty, remove it from the object
	if emptyRel {
		entry.mu.Lock()
		delete(entry.relations, rel)
		emptyObj := len(entry.relations) == 0
		entry.mu.Unlock()

		// Step 6: If the object is now empty, remove it from the shard
		if emptyObj {
			s.mu.Lock()
			delete(s.graph, obj)
			s.mu.Unlock()
		}
	}
}

// ----only shared and sync on obj------.
type ObjSOEntry struct {
	relations map[string]map[entity.Instance]struct{}
	mu        sync.RWMutex
}

type shardSO struct {
	graph map[entity.Instance]*ObjSOEntry
	mu    sync.RWMutex
}

type GraphSOSharded struct {
	shards []*shardSO
	n      int
}

func NewGraphSOSharded(numShards int) *GraphSOSharded {
	shards := make([]*shardSO, numShards)
	for i := range numShards {
		shards[i] = &shardSO{graph: make(map[entity.Instance]*ObjSOEntry)}
	}

	return &GraphSOSharded{shards: shards, n: numShards}
}

// simple shard by obj.Id uint64.
func (g *GraphSOSharded) getShard(obj entity.Instance) *shardSO {
	sum, _ := strconv.Atoi(obj.Id)

	return g.shards[sum%g.n]
}

// get or create object entry.
func (g *GraphSOSharded) getObj(obj entity.Instance) *ObjSOEntry {
	s := g.getShard(obj)
	s.mu.RLock()
	entry := s.graph[obj]
	s.mu.RUnlock()

	if entry == nil {
		s.mu.Lock()

		if s.graph[obj] == nil {
			s.graph[obj] = &ObjSOEntry{relations: make(map[string]map[entity.Instance]struct{})}
		}

		entry = s.graph[obj]
		s.mu.Unlock()
	}

	return entry
}

// Read only locks the object.
func (g *GraphSOSharded) Read(obj entity.Instance, rel string, sbj entity.Instance) bool {
	entry := g.getObj(obj)

	entry.mu.RLock()
	defer entry.mu.RUnlock()

	sbjs, ok := entry.relations[rel]
	if !ok {
		return false
	}

	_, exists := sbjs[sbj]

	return exists
}

// Write only locks the object.
func (g *GraphSOSharded) Write(obj entity.Instance, rel string, sbj entity.Instance) {
	entry := g.getObj(obj)

	entry.mu.Lock()
	defer entry.mu.Unlock()

	if entry.relations[rel] == nil {
		entry.relations[rel] = make(map[entity.Instance]struct{})
	}

	entry.relations[rel][sbj] = struct{}{}
}

// Delete only locks the object.
func (g *GraphSOSharded) delete(obj entity.Instance, rel string, sbj entity.Instance) {
	entry := g.getObj(obj)

	entry.mu.Lock()
	defer entry.mu.Unlock()

	sbjs, ok := entry.relations[rel]
	if !ok {
		return
	}

	delete(sbjs, sbj)

	if len(sbjs) == 0 {
		delete(entry.relations, rel)
	}

	// Optional: remove object from shard if empty
	if len(entry.relations) == 0 {
		s := g.getShard(obj)
		s.mu.Lock()
		delete(s.graph, obj)
		s.mu.Unlock()
	}
}

// -------------------- Benchmark --------------------

func benchmarkGraph(b *testing.B,
	readFunc func(obj entity.Instance, rel string, sbj entity.Instance) bool,
	writeFunc func(obj entity.Instance, rel string, sbj entity.Instance),
	deleteFunc func(obj entity.Instance, rel string, sbj entity.Instance),
) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup

		seeds := make([]int64, numGoroutines)
		for i := range numGoroutines {
			seeds[i] = time.Now().UnixNano() + int64(i)
		}

		for g := range numGoroutines {
			go func(seed int64) {
				rnd := rand.New(rand.NewSource(seed))
				for range numOps {
					x := rnd.Intn(1000)
					_ = x
				}
			}(seeds[g])
		}

		for g := range numGoroutines {
			wg.Add(1)

			go func() {
				defer wg.Done()

				rnd := rand.New(rand.NewSource(seeds[g]))
				for range numOps {
					obj := entity.Instance{
						Id: strconv.Itoa(rnd.Intn(numObjects)),
					}
					rel := fmt.Sprintf("r%d", rnd.Intn(numRelations))

					sbj := entity.Instance{
						Id: strconv.Itoa(rnd.Intn(numSubjects)),
					}
					if rnd.Intn(100) < readPercent {
						readFunc(obj, rel, sbj)
					} else {
						if rnd.Intn(100) < 95 {
							writeFunc(obj, rel, sbj)
						} else {
							deleteFunc(obj, rel, sbj)
						}
					}
				}
			}()
		}

		wg.Wait()
	}
}

// func BenchmarkPerObjectGraph(b *testing.B) {
// 	g := NewGraphPerObject()
// 	benchmarkGraph(b,
// 		g.Read,
// 		g.Write,
// 		g.delete,
// 	)
// }

// func BenchmarkPerObjRelGraph(b *testing.B) {
// 	g := NewGraphPerObjRel()
// 	benchmarkGraph(b,
// 		g.Read,
// 		g.Write,
// 		g.delete,
// 	)
// }

// func BenchmarkPerObjRelSbjGraph(b *testing.B) {
// 	g := NewGraphPerObjRelSbj()
// 	benchmarkGraph(b,
// 		g.Read,
// 		g.Write,
// 	)
// }

func BenchmarkShardedGraph(b *testing.B) {
	g := NewGraphSharded(numShards)
	benchmarkGraph(b,
		g.Read,
		g.Write,
		g.delete,
	)
}

func BenchmarkShardedSOGraph(b *testing.B) {
	g := NewGraphSOSharded(numShards)
	benchmarkGraph(b,
		g.Read,
		g.Write,
		g.delete,
	)
}

// func BenchmarkFormal(b *testing.B) {
// 	g := rbac.NewGraph()
// 	benchmarkGraph(b,
// 		g.Exist,
// 		g.Create,
// 		g.Delete,
// 	)
// }

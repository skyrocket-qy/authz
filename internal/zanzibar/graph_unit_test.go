package zanzibar

import (
	"strconv"
	"testing"

	"authz/internal/entity"

	"github.com/stretchr/testify/assert"
)

func TestNewGraph(t *testing.T) {
	g := NewGraph()
	assert.NotNil(t, g)
	assert.Len(t, g.Shards, NumShards)
	for _, shard := range g.Shards {
		assert.NotNil(t, shard)
		assert.NotNil(t, shard.Graph)
	}
}

func TestGetShard(t *testing.T) {
	g := NewGraph()
	obj := entity.Instance{Id: "123"}
	shard := g.getShard(obj)
	assert.NotNil(t, shard)

	sum, _ := strconv.Atoi(obj.Id)
	expectedShard := g.Shards[sum%NumShards]
	assert.Equal(t, expectedShard, shard)
}

func TestGraph_Create(t *testing.T) {
	g := NewGraph()
	obj := entity.Instance{Ns: "doc", Id: "1"}
	sbj := entity.Instance{Ns: "user", Id: "1"}
	rel := "owner"

	// Create a new relationship
	g.Create(obj, rel, sbj)
	assert.True(t, g.Exist(obj, rel, sbj))

	// Create a relationship that already exists
	g.Create(obj, rel, sbj)
	assert.True(t, g.Exist(obj, rel, sbj))

	// Create a relationship for a new object
	obj2 := entity.Instance{Ns: "doc", Id: "2"}
	g.Create(obj2, rel, sbj)
	assert.True(t, g.Exist(obj2, rel, sbj))
}

func TestGraph_Exist(t *testing.T) {
	g := NewGraph()
	obj := entity.Instance{Ns: "doc", Id: "1"}
	sbj := entity.Instance{Ns: "user", Id: "1"}
	rel := "owner"

	g.Create(obj, rel, sbj)

	// Check for an existing relationship
	assert.True(t, g.Exist(obj, rel, sbj))

	// Check for a non-existing relationship
	assert.False(t, g.Exist(obj, "viewer", sbj))

	// Check for a relationship with a non-existing object
	assert.False(t, g.Exist(entity.Instance{Ns: "doc", Id: "2"}, rel, sbj))
}

func TestGraph_Delete(t *testing.T) {
	g := NewGraph()
	obj := entity.Instance{Ns: "doc", Id: "1"}
	sbj1 := entity.Instance{Ns: "user", Id: "1"}
	sbj2 := entity.Instance{Ns: "user", Id: "2"}
	rel1 := "owner"
	rel2 := "editor"

	g.Create(obj, rel1, sbj1)
	g.Create(obj, rel1, sbj2)
	g.Create(obj, rel2, sbj1)

	// Delete an existing relationship
	g.Delete(obj, rel1, sbj1)
	assert.False(t, g.Exist(obj, rel1, sbj1))
	assert.True(t, g.Exist(obj, rel1, sbj2)) // Ensure other sbj is not deleted
	assert.True(t, g.Exist(obj, rel2, sbj1)) // Ensure other rel is not deleted

	// Delete a non-existing relationship
	g.Delete(obj, "viewer", sbj1)
	assert.False(t, g.Exist(obj, "viewer", sbj1))

	// Delete the last subject for a relation
	g.Delete(obj, rel1, sbj2)
	assert.False(t, g.Exist(obj, rel1, sbj2))
	s := g.getShard(obj)
	s.mu.RLock()
	objEntry := s.Graph[obj]
	s.mu.RUnlock()
	objEntry.mu.RLock()
	_, ok := objEntry.Relations[rel1]
	objEntry.mu.RUnlock()
	assert.False(t, ok, "relation should be deleted when it has no subjects")

	// Delete the last relation for an object
	g.Delete(obj, rel2, sbj1)
	assert.False(t, g.Exist(obj, rel2, sbj1))
	s.mu.RLock()
	_, ok = s.Graph[obj]
	s.mu.RUnlock()
	assert.False(t, ok, "object should be deleted when it has no relations")
}

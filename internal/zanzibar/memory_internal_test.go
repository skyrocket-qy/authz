package zanzibar

import (
	"context"
	"testing"

	"authz/internal/config"
	"authz/internal/entity"
	"authz/internal/schema"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

func setupZanzibarMemory(t *testing.T, s *schema.Schema) *ZanzibarMemoryImpl {
	t.Helper()
	config.Conf.MaxCheckNodes = 100
	return &ZanzibarMemoryImpl{
		Schema: s,
		Graph:  NewGraph(),
	}
}

func TestClose(t *testing.T) {
	cancelled := false
	cancel := func() {
		cancelled = true
	}
	zm := &ZanzibarMemoryImpl{
		Cancel: cancel,
	}
	err := zm.Close(context.Background())
	assert.NoError(t, err)
	assert.True(t, cancelled)
}

func TestCheckUnion(t *testing.T) {
	s := &schema.Schema{
		Namespaces: map[string]*schema.Namespace{
			"doc": {
				Relations: map[string]*schema.Relation{
					"viewer": {
						UsersetRewrite: schema.UsersetRewrite{
							Union: []*schema.UsersetRewrite{
								{ComputedUserSet: &schema.ComputedUserset{Relation: "owner"}},
								{ComputedUserSet: &schema.ComputedUserset{Relation: "editor"}},
							},
						},
					},
					"owner": {
						UsersetRewrite: schema.UsersetRewrite{
							ComputedUserSet: &schema.ComputedUserset{Relation: "owner"},
						},
					},
					"editor": {
						UsersetRewrite: schema.UsersetRewrite{
							ComputedUserSet: &schema.ComputedUserset{Relation: "editor"},
						},
					},
				},
			},
		},
	}

	zm := setupZanzibarMemory(t, s)
	user1 := entity.Instance{Ns: "user", Id: "1"}
	user2 := entity.Instance{Ns: "user", Id: "2"}
	doc := entity.Instance{Ns: "doc", Id: "1"}

	// user1 is owner
	zm.Graph.Create(doc, "owner", user1)
	ok, err := zm.Check(context.Background(), user1, "viewer", doc)
	assert.NoError(t, err)
	assert.True(t, ok)

	// user2 is editor
	zm.Graph.Create(doc, "editor", user2)
	ok, err = zm.Check(context.Background(), user2, "viewer", doc)
	assert.NoError(t, err)
	assert.True(t, ok)

	// user3 is neither
	ok, err = zm.Check(context.Background(), entity.Instance{Ns: "user", Id: "3"}, "viewer", doc)
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestCheckIntersection(t *testing.T) {
	s := &schema.Schema{
		Namespaces: map[string]*schema.Namespace{
			"doc": {
				Relations: map[string]*schema.Relation{
					"viewer": {
						UsersetRewrite: schema.UsersetRewrite{
							Intersection: []*schema.UsersetRewrite{
								{ComputedUserSet: &schema.ComputedUserset{Relation: "member"}},
								{ComputedUserSet: &schema.ComputedUserset{Relation: "sharable"}},
							},
						},
					},
					"member": {
						UsersetRewrite: schema.UsersetRewrite{
							ComputedUserSet: &schema.ComputedUserset{Relation: "member"},
						},
					},
					"sharable": {
						UsersetRewrite: schema.UsersetRewrite{
							ComputedUserSet: &schema.ComputedUserset{Relation: "sharable"},
						},
					},
				},
			},
		},
	}

	zm := setupZanzibarMemory(t, s)
	user1 := entity.Instance{Ns: "user", Id: "1"}
	user2 := entity.Instance{Ns: "user", Id: "2"}
	doc := entity.Instance{Ns: "doc", Id: "1"}

	// user1 is member and sharable
	zm.Graph.Create(doc, "member", user1)
	zm.Graph.Create(doc, "sharable", user1)
	ok, err := zm.Check(context.Background(), user1, "viewer", doc)
	assert.NoError(t, err)
	assert.True(t, ok)

	// user2 is only member
	zm.Graph.Create(doc, "member", user2)
	ok, err = zm.Check(context.Background(), user2, "viewer", doc)
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestCheckExclusion(t *testing.T) {
	s := &schema.Schema{
		Namespaces: map[string]*schema.Namespace{
			"doc": {
				Relations: map[string]*schema.Relation{
					"viewer": {
						UsersetRewrite: schema.UsersetRewrite{
							Exclusion: &schema.ExclusionNode{
								Base:     &schema.UsersetRewrite{ComputedUserSet: &schema.ComputedUserset{Relation: "member"}},
								Subtract: &schema.UsersetRewrite{ComputedUserSet: &schema.ComputedUserset{Relation: "banned"}},
							},
						},
					},
					"member": {
						UsersetRewrite: schema.UsersetRewrite{
							ComputedUserSet: &schema.ComputedUserset{Relation: "member"},
						},
					},
					"banned": {
						UsersetRewrite: schema.UsersetRewrite{
							ComputedUserSet: &schema.ComputedUserset{Relation: "banned"},
						},
					},
				},
			},
		},
	}

	zm := setupZanzibarMemory(t, s)
	user1 := entity.Instance{Ns: "user", Id: "1"}
	user2 := entity.Instance{Ns: "user", Id: "2"}
	doc := entity.Instance{Ns: "doc", Id: "1"}

	// user1 is member but not banned
	zm.Graph.Create(doc, "member", user1)
	ok, err := zm.Check(context.Background(), user1, "viewer", doc)
	assert.NoError(t, err)
	assert.True(t, ok)

	// user2 is member and banned
	zm.Graph.Create(doc, "member", user2)
	zm.Graph.Create(doc, "banned", user2)
	ok, err = zm.Check(context.Background(), user2, "viewer", doc)
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestCheckTupleToUserset(t *testing.T) {
	relation := "parent"
	s := &schema.Schema{
		Namespaces: map[string]*schema.Namespace{
			"doc": {
				Relations: map[string]*schema.Relation{
					"viewer": {
						UsersetRewrite: schema.UsersetRewrite{
							Union: []*schema.UsersetRewrite{
								{ComputedUserSet: &schema.ComputedUserset{Relation: "editor"}},
								{
									TupleToUserset: &schema.TupleToUserset{
										Tupleset:        &schema.Userset{Relation: &relation},
										ComputedUserset: &schema.ComputedUserset{Relation: "viewer"},
									},
								},
							},
						},
					},
					"editor": {
						UsersetRewrite: schema.UsersetRewrite{
							ComputedUserSet: &schema.ComputedUserset{Relation: "editor"},
						},
					},
					"parent": {
						UsersetRewrite: schema.UsersetRewrite{
							ComputedUserSet: &schema.ComputedUserset{Relation: "parent"},
						},
					},
				},
			},
			"folder": {
				Relations: map[string]*schema.Relation{
					"viewer": {
						UsersetRewrite: schema.UsersetRewrite{
							ComputedUserSet: &schema.ComputedUserset{Relation: "viewer"},
						},
					},
				},
			},
		},
	}

	zm := setupZanzibarMemory(t, s)
	user := entity.Instance{Ns: "user", Id: "1"}
	doc := entity.Instance{Ns: "doc", Id: "1"}
	folder := entity.Instance{Ns: "folder", Id: "1"}

	// doc's parent is folder
	zm.Graph.Create(doc, "parent", folder)
	// user is viewer of folder
	zm.Graph.Create(folder, "viewer", user)

	// user should be viewer of doc
	ok, err := zm.Check(context.Background(), user, "viewer", doc)
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestApplyMessage(t *testing.T) {
	zm := setupZanzibarMemory(t, nil)

	// Test create
	msgCreate := kafka.Message{
		Value:  []byte(`{"sbj_ns":"user","sbj_id":"1","relation":"owner","obj_ns":"doc","obj_id":"1","__op":"c"}`),
		Offset: 1,
	}
	err := zm.applyMessage(msgCreate)
	assert.NoError(t, err)
	assert.True(t, zm.Graph.Exist(
		entity.Instance{Ns: "doc", Id: "1"},
		"owner",
		entity.Instance{Ns: "user", Id: "1"},
	))
	assert.Equal(t, int64(1), zm.Offest)

	// Test delete
	msgDelete := kafka.Message{
		Value:  []byte(`{"sbj_ns":"user","sbj_id":"1","relation":"owner","obj_ns":"doc","obj_id":"1","__op":"d"}`),
		Offset: 2,
	}
	err = zm.applyMessage(msgDelete)
	assert.NoError(t, err)
	assert.False(t, zm.Graph.Exist(
		entity.Instance{Ns: "doc", Id: "1"},
		"owner",
		entity.Instance{Ns: "user", Id: "1"},
	))
	assert.Equal(t, int64(2), zm.Offest)
}

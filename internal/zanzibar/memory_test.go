package zanzibar_test

import (
	"context"
	"testing"

	"authz/internal/config"
	"authz/internal/entity"
	"authz/internal/schema"
	"authz/internal/zanzibar"
	"github.com/stretchr/testify/assert"
)

func setupZanzibarMemory(t *testing.T, s *schema.Schema) *zanzibar.ZanzibarMemoryImpl {
	t.Helper()
	config.Conf.MaxCheckNodes = 100
	return &zanzibar.ZanzibarMemoryImpl{
		Schema: s,
		Graph:  zanzibar.NewGraph(),
	}
}

func TestCheckDirect(t *testing.T) {
	s := &schema.Schema{
		Namespaces: map[string]*schema.Namespace{
			"doc": {
				Relations: map[string]*schema.Relation{
					"owner": {
						UsersetRewrite: schema.UsersetRewrite{
							ComputedUserSet: &schema.ComputedUserset{
								Relation: "owner",
							},
						},
					},
				},
			},
		},
	}

	zm := setupZanzibarMemory(t, s)
	user := entity.Instance{Ns: "user", Id: "1"}
	doc := entity.Instance{Ns: "doc", Id: "1"}
	zm.Graph.Create(doc, "owner", user)

	ok, err := zm.Check(context.Background(), user, "owner", doc)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = zm.Check(context.Background(), entity.Instance{Ns: "user", Id: "2"}, "owner", doc)
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestLookup(t *testing.T) {
	s := &schema.Schema{
		Namespaces: map[string]*schema.Namespace{
			"doc": {
				Relations: map[string]*schema.Relation{
					"owner": {
						UsersetRewrite: schema.UsersetRewrite{
							ComputedUserSet: &schema.ComputedUserset{
								Relation: "owner",
							},
						},
					},
				},
			},
		},
	}

	zm := setupZanzibarMemory(t, s)
	user := entity.Instance{Ns: "user", Id: "1"}
	doc1 := entity.Instance{Ns: "doc", Id: "1"}
	doc2 := entity.Instance{Ns: "doc", Id: "2"}
	zm.Graph.Create(doc1, "owner", user)
	zm.Graph.Create(doc2, "owner", user)

	objs, err := zm.Lookup(context.Background(), user, "owner")
	assert.NoError(t, err)
	assert.ElementsMatch(t, []entity.Instance{doc1, doc2}, objs)
}

func TestExpand(t *testing.T) {
	s := &schema.Schema{
		Namespaces: map[string]*schema.Namespace{
			"doc": {
				Relations: map[string]*schema.Relation{
					"owner": {},
				},
			},
		},
	}

	zm := setupZanzibarMemory(t, s)
	user1 := entity.Instance{Ns: "user", Id: "1"}
	user2 := entity.Instance{Ns: "user", Id: "2"}
	doc := entity.Instance{Ns: "doc", Id: "1"}
	zm.Graph.Create(doc, "owner", user1)
	zm.Graph.Create(doc, "owner", user2)

	users, err := zm.Expand(context.Background(), "owner", doc)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []entity.Instance{user1, user2}, users)
}

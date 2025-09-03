package zanzibar

import (
	"testing"

	"authz/internal/schema"
	"github.com/stretchr/testify/assert"
)

func TestRewriteIncludesDirect(t *testing.T) {
	t.Run("should return true if any child in union returns true", func(t *testing.T) {
		r := &ZanzibarLogicImpl{}
		permSet := map[permission]string{
			{resNs: "doc", resId: "1", perm: "owner"}: "direct",
		}
		ur := &schema.UsersetRewrite{
			Union: []*schema.UsersetRewrite{
				{ComputedUserSet: &schema.ComputedUserset{Relation: "viewer"}},
				{ComputedUserSet: &schema.ComputedUserset{Relation: "owner"}},
			},
		}
		assert.True(t, r.rewriteIncludesDirect("doc", "1", ur, permSet))
	})

	t.Run("should return true if all children in intersection return true", func(t *testing.T) {
		r := &ZanzibarLogicImpl{}
		permSet := map[permission]string{
			{resNs: "doc", resId: "1", perm: "owner"}:  "direct",
			{resNs: "doc", resId: "1", perm: "viewer"}: "direct",
		}
		ur := &schema.UsersetRewrite{
			Intersection: []*schema.UsersetRewrite{
				{ComputedUserSet: &schema.ComputedUserset{Relation: "viewer"}},
				{ComputedUserSet: &schema.ComputedUserset{Relation: "owner"}},
			},
		}
		assert.True(t, r.rewriteIncludesDirect("doc", "1", ur, permSet))
	})

	t.Run("should return false if any child in intersection returns false", func(t *testing.T) {
		r := &ZanzibarLogicImpl{}
		permSet := map[permission]string{
			{resNs: "doc", resId: "1", perm: "owner"}: "direct",
		}
		ur := &schema.UsersetRewrite{
			Intersection: []*schema.UsersetRewrite{
				{ComputedUserSet: &schema.ComputedUserset{Relation: "viewer"}},
				{ComputedUserSet: &schema.ComputedUserset{Relation: "owner"}},
			},
		}
		assert.False(t, r.rewriteIncludesDirect("doc", "1", ur, permSet))
	})

	t.Run("should return true if base is true and subtract is false in exclusion", func(t *testing.T) {
		r := &ZanzibarLogicImpl{}
		permSet := map[permission]string{
			{resNs: "doc", resId: "1", perm: "owner"}: "direct",
		}
		ur := &schema.UsersetRewrite{
			Exclusion: &schema.ExclusionNode{
				Base: &schema.UsersetRewrite{
					ComputedUserSet: &schema.ComputedUserset{Relation: "owner"},
				},
				Subtract: &schema.UsersetRewrite{
					ComputedUserSet: &schema.ComputedUserset{Relation: "viewer"},
				},
			},
		}
		assert.True(t, r.rewriteIncludesDirect("doc", "1", ur, permSet))
	})

	t.Run("should return false if base is false in exclusion", func(t *testing.T) {
		r := &ZanzibarLogicImpl{}
		permSet := map[permission]string{
			{resNs: "doc", resId: "1", perm: "viewer"}: "direct",
		}
		ur := &schema.UsersetRewrite{
			Exclusion: &schema.ExclusionNode{
				Base: &schema.UsersetRewrite{
					ComputedUserSet: &schema.ComputedUserset{Relation: "owner"},
				},
				Subtract: &schema.UsersetRewrite{
					ComputedUserSet: &schema.ComputedUserset{Relation: "viewer"},
				},
			},
		}
		assert.False(t, r.rewriteIncludesDirect("doc", "1", ur, permSet))
	})

	t.Run("should return false if subtract is true in exclusion", func(t *testing.T) {
		r := &ZanzibarLogicImpl{}
		permSet := map[permission]string{
			{resNs: "doc", resId: "1", perm: "owner"}:  "direct",
			{resNs: "doc", resId: "1", perm: "viewer"}: "direct",
		}
		ur := &schema.UsersetRewrite{
			Exclusion: &schema.ExclusionNode{
				Base: &schema.UsersetRewrite{
					ComputedUserSet: &schema.ComputedUserset{Relation: "owner"},
				},
				Subtract: &schema.UsersetRewrite{
					ComputedUserSet: &schema.ComputedUserset{Relation: "viewer"},
				},
			},
		}
		assert.False(t, r.rewriteIncludesDirect("doc", "1", ur, permSet))
	})

	t.Run("should return true if computed userset relation exists in permSet", func(t *testing.T) {
		r := &ZanzibarLogicImpl{}
		permSet := map[permission]string{
			{resNs: "doc", resId: "1", perm: "owner"}: "direct",
		}
		ur := &schema.UsersetRewrite{
			ComputedUserSet: &schema.ComputedUserset{Relation: "owner"},
		}
		assert.True(t, r.rewriteIncludesDirect("doc", "1", ur, permSet))
	})

	t.Run("should return false if computed userset relation does not exist in permSet", func(t *testing.T) {
		r := &ZanzibarLogicImpl{}
		permSet := map[permission]string{
			{resNs: "doc", resId: "1", perm: "viewer"}: "direct",
		}
		ur := &schema.UsersetRewrite{
			ComputedUserSet: &schema.ComputedUserset{Relation: "owner"},
		}
		assert.False(t, r.rewriteIncludesDirect("doc", "1", ur, permSet))
	})

	t.Run("should return true if tuple to userset is valid", func(t *testing.T) {
		r := &ZanzibarLogicImpl{}
		permSet := map[permission]string{
			{resNs: "doc", resId: "1", perm: "owner"}: "direct",
		}
		relation := "owner"
		ur := &schema.UsersetRewrite{
			TupleToUserset: &schema.TupleToUserset{
				Tupleset: &schema.Userset{
					Relation: &relation,
				},
				ComputedUserset: &schema.ComputedUserset{Relation: "owner"},
			},
		}
		assert.True(t, r.rewriteIncludesDirect("doc", "1", ur, permSet))
	})

	t.Run("should return false if tuple to userset is not valid", func(t *testing.T) {
		r := &ZanzibarLogicImpl{}
		permSet := map[permission]string{
			{resNs: "doc", resId: "1", perm: "viewer"}: "direct",
		}
		relation := "owner"
		ur := &schema.UsersetRewrite{
			TupleToUserset: &schema.TupleToUserset{
				Tupleset: &schema.Userset{
					Relation: &relation,
				},
				ComputedUserset: &schema.ComputedUserset{Relation: "owner"},
			},
		}
		assert.False(t, r.rewriteIncludesDirect("doc", "1", ur, permSet))
	})
}

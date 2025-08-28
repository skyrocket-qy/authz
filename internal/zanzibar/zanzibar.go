package zanzibar

import (
	"context"

	"authz/internal/entity"
	"authz/internal/entity/model"
	"authz/internal/schema"
	"authz/internal/util"

	"github.com/redis/go-redis/v9"
	"github.com/skyrocket-qy/erx"
	"github.com/skyrocket-qy/protos/gen/authzpb/rbacpb"
	authzpbv1 "github.com/skyrocket-qy/protos/gen/authzpb/v1"
	pkgpbv1 "github.com/skyrocket-qy/protos/gen/pkgpb/v1"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

type ZanzibarLogic interface {
	// Tuple
	Create(c context.Context, tuple *authzpbv1.Tuple) error
	List(c context.Context, in *authzpbv1.ListTuplesIn) (*authzpbv1.ListTuplesOut, error)
	Find(c context.Context, filter *authzpbv1.TupleFilter) ([]*authzpbv1.Tuple, error)
	Delete(c context.Context, in *authzpbv1.DeleteTuplesIn) error
	GetPermissions(c context.Context, sbj *authzpbv1.Instance, nsType string) (
		[]*rbacpb.Permission, error,
	)

	Check(c context.Context, in *authzpbv1.CheckIn) (*authzpbv1.CheckOut, error)
	// Lookup(c context.Context, sbj *entity.Instance, rel string) ([]*entity.Instance, error)
	// Expand(c context.Context, relation string, obj *entity.Instance) ([]entity.Instance, error)
}

var _ ZanzibarLogic = (*ZanzibarLogicImpl)(nil)

type ZanzibarLogicImpl struct {
	pgdb   *gorm.DB
	rdb    *redis.Client
	zm     ZanzibarMemory
	schema *schema.Schema
}

func NewZanzibarLogic(db *gorm.DB, rdb *redis.Client, zm ZanzibarMemory,
	s *schema.Schema,
) *ZanzibarLogicImpl {
	return &ZanzibarLogicImpl{
		pgdb:   db,
		rdb:    rdb,
		zm:     zm,
		schema: s,
	}
}

func (r *ZanzibarLogicImpl) db(c context.Context) *gorm.DB {
	if db, ok := c.Value(util.DbCtxKey{}).(*gorm.DB); ok && db != nil {
		return db
	}

	return r.pgdb.WithContext(c)
}

func (r *ZanzibarLogicImpl) List(c context.Context, in *authzpbv1.ListTuplesIn) (
	out *authzpbv1.ListTuplesOut, err error,
) {
	validFilterFields := []string{"sbj_ns", "sbj_id", "relation", "obj_ns", "obj_id"}

	if in.GetCursor() == nil {
		return nil, erx.New(util.ErrBadRequest, "cursor is nil")
	}

	pager := in.GetCursor()
	cursorVal := pager.GetVal()

	filterScope, err := util.ApplyFilter(in.GetFilters(), validFilterFields, nil)
	if err != nil {
		return nil, erx.W(err)
	}

	tupleModels := []*model.Tuple{}
	tx := r.db(c).
		Limit(int(pager.GetSize())).
		Scopes(
			filterScope,
			util.ApplySorter(in.GetSorters()),
		)

	if cursorVal != nil {
		nextCursorData := &pkgpbv1.CursorData{}
		if err := proto.Unmarshal(cursorVal, nextCursorData); err != nil {
			return nil, erx.W(err)
		}

		tx = tx.Scopes(util.ApplyCursor(nextCursorData))
	}

	if err := tx.
		Find(&tupleModels).Error; err != nil {
		return nil, erx.W(err)
	}

	tuples := make([]*authzpbv1.Tuple, len(tupleModels))
	for i, tuple := range tupleModels {
		tuples[i] = &authzpbv1.Tuple{
			SbjNs: tuple.SbjNs,
			SbjId: tuple.SbjId,
			Rel:   tuple.Relation,
			ObjNs: tuple.ObjNs,
			ObjId: tuple.ObjId,
		}
	}

	out = &authzpbv1.ListTuplesOut{
		Tuples: tuples,
	}

	if len(tuples) == 0 {
		return out, nil
	}

	lastTuple := tuples[len(tuples)-1]

	cursorData := &pkgpbv1.CursorData{
		Fields: make([]*pkgpbv1.Field, len(in.GetSorters())),
	}
	for i, sorter := range in.GetSorters() {
		field := &pkgpbv1.Field{
			Asc: sorter.GetAsc(),
			Col: sorter.GetField(),
		}
		switch sorter.GetField() {
		case "sbj_ns":
			field.Val = lastTuple.GetSbjNs()
		case "sbj_id":
			field.Val = lastTuple.GetSbjId()
		case "relation":
			field.Val = lastTuple.GetRel()
		case "obj_ns":
			field.Val = lastTuple.GetObjNs()
		case "obj_id":
			field.Val = lastTuple.GetObjId()
		}

		cursorData.Fields[i] = field
	}

	out.NextCursor, err = proto.Marshal(cursorData)
	if err != nil {
		return nil, erx.W(err)
	}

	return out, nil
}

func (r *ZanzibarLogicImpl) Find(c context.Context, filter *authzpbv1.TupleFilter) (
	[]*authzpbv1.Tuple, error,
) {
	tuples := []*model.Tuple{}
	if err := r.db(c).Scopes(ApplyTupleFilter(filter)).Find(&tuples).Error; err != nil {
		return nil, erx.W(err)
	}

	tuplesProto := make([]*authzpbv1.Tuple, 0, len(tuples))
	for _, tuple := range tuples {
		tuplesProto = append(tuplesProto, &authzpbv1.Tuple{
			SbjNs: tuple.SbjNs,
			SbjId: tuple.SbjId,
			Rel:   tuple.Relation,
			ObjNs: tuple.ObjNs,
			ObjId: tuple.ObjId,
		})
	}

	return tuplesProto, nil
}

func (r *ZanzibarLogicImpl) Create(c context.Context, tuple *authzpbv1.Tuple) error {
	return r.db(c).Create(&model.Tuple{
		SbjNs:    tuple.GetSbjNs(),
		SbjId:    tuple.GetSbjId(),
		Relation: tuple.GetRel(),
		ObjNs:    tuple.GetObjNs(),
		ObjId:    tuple.GetObjId(),
	}).Error
}

func (r *ZanzibarLogicImpl) Delete(c context.Context, in *authzpbv1.DeleteTuplesIn) error {
	switch req := in.GetMode().(type) {
	case *authzpbv1.DeleteTuplesIn_Filter:
		filter := req.Filter

		return r.db(c).Scopes(ApplyTupleFilter(filter)).Delete(&model.Tuple{}).Error

	case *authzpbv1.DeleteTuplesIn_DeleteTuples:
		tuples := req.DeleteTuples.GetTuples()

		values := make([][]any, len(tuples))
		for i, t := range tuples {
			values[i] = []any{t.GetSbjNs(), t.GetSbjId(), t.GetRel(), t.GetObjNs(), t.GetObjId()}
		}

		return r.db(c).
			Where("(sbj_ns, sbj_id, rel, obj_ns, obj_id) IN ?", values).
			Delete(&model.Tuple{}).Error

	case *authzpbv1.DeleteTuplesIn_DeleteTupleIds:
		if err := r.db(c).
			Where("id IN ?", req.DeleteTupleIds.GetIds()).
			Delete(&model.Tuple{}).Error; err != nil {
			return erx.W(err)
		}
	default:
		return erx.New(util.ErrBadRequest, "mode type error")
	}

	return nil
}

func (r *ZanzibarLogicImpl) Check(c context.Context, in *authzpbv1.CheckIn) (
	*authzpbv1.CheckOut, error,
) {
	user := entity.Instance{
		Ns: in.GetSbjNs(),
		Id: in.GetSbjId(),
	}

	obj := entity.Instance{
		Ns: in.GetObjNs(),
		Id: in.GetObjId(),
	}

	ok, err := r.zm.Check(c, user, in.GetRel(), obj)
	if err != nil {
		return nil, erx.W(err)
	}

	return &authzpbv1.CheckOut{IsAllowed: ok}, nil
}

// What are all the objs that sbj has rel on.
func (r *ZanzibarLogicImpl) Lookup(c context.Context, user *entity.Instance, perm string) (
	objs []*entity.Instance, err error,
) {
	return objs, nil
}

func (r *ZanzibarLogicImpl) Expand(c context.Context, perm string, obj *entity.Instance) (
	users []entity.Instance, err error,
) {
	return nil, nil
}

type permission struct {
	resNs string
	resId string
	perm  string
}

func (r *ZanzibarLogicImpl) GetPermissions(
	c context.Context,
	sbj *authzpbv1.Instance,
	nsType string,
) (
	perms []*rbacpb.Permission, err error,
) {
	// Step 1: Load direct tuples for subject
	var tuples []*model.Tuple
	if err := r.db(c).Where("sbj_ns = ? AND sbj_id = ?", sbj.GetNs(), sbj.GetId()).Find(&tuples).
		Error; err != nil {
		return nil, erx.W(err)
	}

	// Step 2: Group by namespace/object
	nss := map[string][]string{}
	permSet := map[permission]string{} // value = "direct" or "indirect"

	for _, tuple := range tuples {
		nss[tuple.ObjNs] = append(nss[tuple.ObjNs], tuple.ObjId)
		permSet[permission{tuple.ObjNs, tuple.ObjId, tuple.Relation}] = "direct"
	}

	// Step 3: For each namespace, traverse schema AST for indirect perms
	for ns, ids := range nss {
		nsT := r.schema.Namespaces[ns]
		if nsT.Type != nsType {
			continue
		}

		for rel, relT := range nsT.Relations {
			for _, id := range ids {
				if r.rewriteIncludesDirect(ns, id, &relT.UsersetRewrite, permSet) {
					if _, exists := permSet[permission{ns, id, rel}]; !exists {
						permSet[permission{ns, id, rel}] = "indirect"
					}
				}
			}
		}
	}

	// Step 4: Build final result
	for p, kind := range permSet {
		perm := &rbacpb.Permission{
			ResourceNs: p.resNs,
			ResourceId: p.resId,
			Permission: p.perm,
		}

		if kind == "indirect" {
			perm.Type = rbacpb.PermissionType_COMPUTE
		} else {
			perm.Type = rbacpb.PermissionType_TUPLE
		}

		perms = append(perms, perm)
	}

	return perms, nil
}

// Step 3 helper: Recursively check if rewrite matches any direct perms.
func (r *ZanzibarLogicImpl) rewriteIncludesDirect(resNs, resId string, ur *schema.UsersetRewrite,
	permSet map[permission]string,
) bool {
	// Check union
	if len(ur.Union) > 0 {
		for _, child := range ur.Union {
			if r.rewriteIncludesDirect(resNs, resId, child, permSet) {
				return true
			}
		}
	}

	// Check intersection
	if len(ur.Intersection) > 0 {
		for _, child := range ur.Intersection {
			if !r.rewriteIncludesDirect(resNs, resId, child, permSet) {
				return false
			}
		}

		return true
	}

	// Check exclusion
	if ur.Exclusion != nil {
		return r.rewriteIncludesDirect(resNs, resId, ur.Exclusion.Base, permSet) &&
			!r.rewriteIncludesDirect(resNs, resId, ur.Exclusion.Subtract, permSet)
	}

	// ComputedUserset
	if ur.ComputedUserSet != nil {
		rel := ur.ComputedUserSet.Relation
		if _, ok := permSet[permission{resNs, resId, rel}]; ok {
			return true
		}
	}

	// TupleToUserset
	if ur.TupleToUserset != nil {
		// Example: find tuples matching tupleset relation, then check computed userset
		// This would need a DB lookup or cached index
		// For now, assume you can check permSet directly
		tset := ur.TupleToUserset.Tupleset
		if tset != nil {
			// Check if subject has relation on intermediate object
			if _, ok := permSet[permission{resNs, resId, *tset.Relation}]; ok {
				// Now check computed userset of intermediate
				return ur.TupleToUserset.ComputedUserset != nil
			}
		}
	}

	return false
}

func ApplyTupleFilter(f *authzpbv1.TupleFilter) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if f.SbjNs != nil {
			db = db.Where("sbj_ns = ?", f.GetSbjNs())
		}

		if f.SbjId != nil {
			db = db.Where("sbj_id = ?", f.GetSbjId())
		}

		if f.Rel != nil {
			db = db.Where("relation = ?", f.GetRel())
		}

		if f.ObjNs != nil {
			db = db.Where("obj_ns = ?", f.GetObjNs())
		}

		if f.ObjId != nil {
			db = db.Where("obj_id = ?", f.GetObjId())
		}

		return db
	}
}

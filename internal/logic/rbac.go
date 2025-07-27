package logic

import (
	"authz/internal/entity"
	"context"

	mapset "github.com/deckarep/golang-set/v2"
)

/*
Only 4 type
User, Role(include Admin), Perm, resource
*/

type PermAssignment struct {
	Perm string
	Res  string
}

type RbacLogic interface {
	ListUsers(ctx context.Context, filter string, page, size int) ([]string, error)
	ListRoles(ctx context.Context, filter string, page, size int) ([]string, error)
	GetUsersWithRole(ctx context.Context, role string) ([]string, error)
	GetUsersWithPerm(ctx context.Context, perm string, res string) ([]string, error)

	DeleteUser(c context.Context, user string) error
	DeleteRole(c context.Context, role string) error
	DeletePerm(c context.Context, perm string) error
	DeleteRes(c context.Context, res string) error

	AssignRole(c context.Context, user string, role string) error
	RevokeRole(c context.Context, user string, role string) error
	AssignRoles(ctx context.Context, user string, roles []string) error
	RevokeRoles(ctx context.Context, user string, roles []string) error

	GrantPerm(c context.Context, role string, perm string, res string) error
	GrantPerms(ctx context.Context, role string, perms []*PermAssignment) error
	RevokePerm(c context.Context, role string, perm string, res string) error

	HasRole(c context.Context, user string, role string) (bool, error)
	HasPerm(c context.Context, user string, perm string, res string) (bool, error)

	GetRoles(c context.Context, user string) ([]string, error)
	GetPerms(c context.Context, role string) ([]string, error)
}

var _ RbacLogic = (*RbacLogicImpl)(nil)

type RbacLogicImpl struct {
	zbLogic ZanzibarLogic
}

func NewRbacLogic(zbLogic ZanzibarLogic) *RbacLogicImpl {
	return &RbacLogicImpl{zbLogic: zbLogic}
}

func (r *RbacLogicImpl) ListUser(c context.Context, filter string, page, size int) (
	[]string, error,
) {
	tp, err := r.zbLogic.Find(c, &entity.Tuple{Sbj: &entity.Instance{Ns: "user"}}, false)
	if err != nil {
		return nil, err
	}

	m := mapset.NewSet[string]()
	for _, t := range tp {
		m.Add(t.Sbj.Id)
	}
	return m.ToSlice(), nil
}

func (r *RbacLogicImpl) ListRole(c context.Context, filter string, page, size int) (
	[]string, error,
) {
	tp, err := r.zbLogic.Find(c, &entity.Tuple{Sbj: &entity.Instance{Ns: "role"}}, false)
	if err != nil {
		return nil, err
	}

	m := mapset.NewSet[string]()
	for _, t := range tp {
		m.Add(t.Sbj.Id)
	}
	return m.ToSlice(), nil
}

func (r *RbacLogicImpl) AssignRole(c context.Context, user string, role string) error {
	tuple := &entity.Tuple{
		Sbj: &entity.Instance{Ns: "user", Id: user},
		Rel: "member",
		Obj: &entity.Instance{Ns: "role", Id: role},
	}
	return r.zbLogic.Create(c, tuple)
}

func (r *RbacLogicImpl) RevokeRole(c context.Context, user string, role string) error {
	filter := &entity.Tuple{
		Sbj: &entity.Instance{Ns: "user", Id: user},
		Rel: "member",
		Obj: &entity.Instance{Ns: "role", Id: role},
	}
	return r.zbLogic.Delete(c, filter)
}

func (r *RbacLogicImpl) GrantPerm(c context.Context, role string, permission string, resource string) error {
	tuple := &entity.Tuple{
		Sbj: &entity.Instance{Ns: "role", Id: role},
		Rel: permission,
		Obj: &entity.Instance{Ns: "resource", Id: resource},
	}
	return r.zbLogic.Create(c, tuple)
}

func (r *RbacLogicImpl) RevokePerm(c context.Context, role string, permission string, resource string) error {
	filter := &entity.Tuple{
		Sbj: &entity.Instance{Ns: "role", Id: role},
		Rel: permission,
		Obj: &entity.Instance{Ns: "resource", Id: resource},
	}
	return r.zbLogic.Delete(c, filter)
}

func (r *RbacLogicImpl) HasRole(c context.Context, user string, role string) (bool, error) {
	return r.zbLogic.Check(c,
		&entity.Instance{Ns: "user", Id: user},
		"member",
		&entity.Instance{Ns: "role", Id: role},
	)
}

func (r *RbacLogicImpl) HasPerm(c context.Context, user string, permission string, resource string) (bool, error) {
	// This typically needs to check if user has the role that has the permission,
	// or directly assigned permission. Using zbLogic.Check with recursive logic.
	return r.zbLogic.Check(c,
		&entity.Instance{Ns: "user", Id: user},
		permission,
		&entity.Instance{Ns: "resource", Id: resource},
	)
}

func (r *RbacLogicImpl) GetRoles(c context.Context, user string) ([]string, error) {
	objs, err := r.zbLogic.Lookup(c, &entity.Instance{Ns: "user", Id: user}, "member")
	if err != nil {
		return nil, err
	}
	roles := make([]string, 0, len(objs))
	for _, obj := range objs {
		roles = append(roles, obj.Id)
	}
	return roles, nil
}

func (r *RbacLogicImpl) GetPerms(c context.Context, role string) ([]string, error) {
	objs, err := r.zbLogic.Lookup(c, &entity.Instance{Ns: "role", Id: role}, "")
	if err != nil {
		return nil, err
	}
	perms := make([]string, 0, len(objs))
	for _, obj := range objs {
		perms = append(perms, obj.Id)
	}
	return perms, nil
}

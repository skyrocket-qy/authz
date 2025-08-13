package rbac

import (
	"authz/internal/pkg"
	"authz/internal/schema"
	"context"
	"strconv"

	rbacpb "github.com/skyrocket-qy/protos/gen/authzpb/rbacpb"
	authzpbv1 "github.com/skyrocket-qy/protos/gen/authzpb/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type RbacLogic interface {
	ListUsers(c context.Context, in *rbacpb.ListUsersIn) (out *rbacpb.ListUsersOut, err error)
	UpdateUser(c context.Context, in *rbacpb.UpdateUserIn) error
	DeleteUser(c context.Context, in *rbacpb.DeleteUserIn) error

	GetRole(context.Context, *rbacpb.GetRoleIn) (*rbacpb.GetRoleOut, error)
	CreateRole(c context.Context, in *rbacpb.CreateRoleIn) error
	ListRoles(c context.Context, in *rbacpb.ListRolesIn) (*rbacpb.ListRolesOut, error)
	UpdateRole(c context.Context, in *rbacpb.UpdateRoleIn) error
	DeleteRole(c context.Context, in *rbacpb.DeleteRoleIn) error

	ListResourceTypes(context.Context) (
		*rbacpb.ListResourceTypeOut, error,
	)
	ListResourcesByType(context.Context, *rbacpb.ListResourcesByTypeIn) (
		*rbacpb.ListResourcesByTypeOut, error,
	)
	ListPermissionByResource(context.Context, *rbacpb.ListPermissionByResourceIn) (
		*rbacpb.ListPermissionByResourceOut, error,
	)
	CreateResource(c context.Context, in *rbacpb.CreateResourceIn) error
	ListResources(c context.Context, in *rbacpb.ListResourcesIn) (*rbacpb.ListResourcesOut, error)
	DeleteResource(c context.Context, in *rbacpb.DeleteResourceIn) error

	AssignRole(c context.Context, in *rbacpb.AssignRoleIn) error
	RevokeRole(c context.Context, in *rbacpb.RevokeRoleIn) error

	GrantPerm(c context.Context, in *rbacpb.GrantPermIn) error
	RevokePerm(c context.Context, in *rbacpb.RevokePermIn) error
}

var _ RbacLogic = (*RbacLogicImpl)(nil)

type RbacLogicImpl struct {
	pgdb    *gorm.DB
	zbLogic ZanzibarLogic
	schema  *schema.Schema
}

func NewRbacLogic(zbLogic ZanzibarLogic, pgdb *gorm.DB, schema *schema.Schema) *RbacLogicImpl {
	return &RbacLogicImpl{
		zbLogic: zbLogic,
		pgdb:    pgdb,
	}
}

func (r *RbacLogicImpl) db(c context.Context) *gorm.DB {
	return r.pgdb.WithContext(c)
}

func (r *RbacLogicImpl) ListUsers(c context.Context, in *rbacpb.ListUsersIn) (
	out *rbacpb.ListUsersOut, err error,
) {
	filterExprs := map[string][]string{
		"orgs.name": {
			"JOIN user_orgs ON user_orgs.user_id = users.id",
			"JOIN orgs ON orgs.id = user_orgs.org_id",
		},
		"user_auths.type": {
			"JOIN user_auths ON user_auths.user_id = users.id",
		},
	}

	validFilterFields := []string{"created_at", "email", "name", "is_email_confirmed", "is_active",
		"user_auths.type", "orgs.name"}

	filterScope, err := pkg.ApplyFilter(in.Filters, validFilterFields, filterExprs)
	if err != nil {
		return nil, err
	}

	cnt := int64(0)
	if err := r.db(c).Model(&User{}).Scopes(filterScope).Count(&cnt).Error; err != nil {
		return nil, err
	}

	userMds := []*User{}
	if err := r.db(c).
		Scopes(
			filterScope,
			pkg.ApplyPager(in.Pager),
			pkg.ApplySorter(in.Sorters),
		).
		Preload("Orgs").
		Preload("UserAuths").
		Find(&userMds).Error; err != nil {
		return nil, err
	}

	users := make([]*rbacpb.User, len(userMds))
	for i, userMd := range userMds {
		user := &rbacpb.User{
			Id:               userMd.Id,
			Name:             userMd.Name,
			Email:            userMd.Email,
			IsActive:         userMd.IsActive,
			IsEmailConfirmed: userMd.IsEmailConfirmed,
			CreatedAt:        timestamppb.New(userMd.CreatedAt),
		}

		for _, userAuth := range userMd.UserAuths {
			user.AuthTypes = append(user.AuthTypes, userAuth.Type)
		}

		users[i] = user
	}
	return &rbacpb.ListUsersOut{
		Users: users,
		Count: cnt,
	}, nil
}

func (r *RbacLogicImpl) UpdateUser(c context.Context, in *rbacpb.UpdateUserIn) error {
	updates := map[string]any{}
	if in.IsActive != nil {
		updates["is_active"] = in.IsActive
	}
	if in.Name != nil {
		updates["name"] = in.Name
	}

	if err := r.db(c).Model(&User{}).Where("id = ?", in.Id).Updates(updates).Error; err != nil {
		return err
	}

	return nil
}

func (r *RbacLogicImpl) DeleteUser(c context.Context, in *rbacpb.DeleteUserIn) error {
	if err := r.db(c).Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&User{}, in.Id).Error; err != nil {
			return err
		}

		return r.zbLogic.Delete(pkg.WithDB(c, tx), &authzpbv1.DeleteTuplesIn{
			Mode: &authzpbv1.DeleteTuplesIn_Filter{
				Filter: &authzpbv1.TupleFilter{
					SbjNs: pkg.Str("user"),
					SbjId: pkg.Str(strconv.FormatUint(in.Id, 10)),
				},
			},
		})
	}); err != nil {
		return err
	}

	return nil
}

func (r *RbacLogicImpl) CreateRole(c context.Context, in *rbacpb.CreateRoleIn) error {
	return r.db(c).Create(&Role{Name: in.Name}).Error
}

func (r *RbacLogicImpl) ListRoles(c context.Context, in *rbacpb.ListRolesIn) (
	out *rbacpb.ListRolesOut, err error,
) {
	validFilterFields := []string{"name"}
	filterScope, err := pkg.ApplyFilter(in.Filters, validFilterFields, nil)
	if err != nil {
		return nil, err
	}

	cnt := int64(0)
	if err := r.db(c).Model(&Role{}).Scopes(filterScope).Count(&cnt).Error; err != nil {
		return nil, err
	}

	roleMds := []*Role{}
	if err := r.db(c).
		Scopes(
			filterScope,
			pkg.ApplyPager(in.Pager),
			pkg.ApplySorter(in.Sorters),
		).
		Find(&roleMds).Error; err != nil {
		return nil, err
	}

	roles := make([]*rbacpb.Role, 0, len(roleMds))
	for _, roleMd := range roleMds {
		roles = append(roles, &rbacpb.Role{
			Id:   roleMd.Id,
			Name: roleMd.Name,
		})
	}

	return &rbacpb.ListRolesOut{Roles: roles, Total: cnt}, nil
}

func (r *RbacLogicImpl) UpdateRole(c context.Context, in *rbacpb.UpdateRoleIn) error {
	return r.db(c).Model(&Role{}).Where("id = ?", in.Id).Updates(map[string]any{
		"name": in.Name,
	}).Error
}

// TODO: wrap to transaction
func (r *RbacLogicImpl) DeleteRole(c context.Context, in *rbacpb.DeleteRoleIn) error {
	return r.db(c).Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&Role{}, in.Id).Error; err != nil {
			return err
		}

		cWithDb := pkg.WithDB(c, tx)

		if err := r.zbLogic.Delete(cWithDb,
			&authzpbv1.DeleteTuplesIn{
				Mode: &authzpbv1.DeleteTuplesIn_Filter{
					Filter: &authzpbv1.TupleFilter{
						SbjNs: pkg.Str("user"),
						Rel:   pkg.Str("member"),
						ObjNs: pkg.Str("role"),
						ObjId: pkg.Str(strconv.FormatUint(in.Id, 10)),
					},
				},
			},
		); err != nil {
			return err
		}

		if err := r.zbLogic.Delete(cWithDb,
			&authzpbv1.DeleteTuplesIn{
				Mode: &authzpbv1.DeleteTuplesIn_Filter{
					Filter: &authzpbv1.TupleFilter{
						SbjNs: pkg.Str("role"),
						Rel:   pkg.Str(strconv.FormatUint(in.Id, 10)),
					},
				},
			},
		); err != nil {
			return err
		}

		return nil
	})
}

func (r *RbacLogicImpl) CreateResource(c context.Context, in *rbacpb.CreateResourceIn) error {
	return r.db(c).Create(&Resource{Ns: in.Ns, Name: in.Name}).Error
}

// TODO: listOption handle
func (r *RbacLogicImpl) ListResources(c context.Context, in *rbacpb.ListResourcesIn) (
	out *rbacpb.ListResourcesOut, err error,
) {
	validFilterFields := []string{"ns", "name"}
	filterScope, err := pkg.ApplyFilter(in.Filters, validFilterFields, nil)
	if err != nil {
		return nil, err
	}

	cnt := int64(0)
	if err := r.db(c).Model(&Resource{}).Scopes(filterScope).Count(&cnt).Error; err != nil {
		return nil, err
	}

	resMds := []*Resource{}
	if err := r.db(c).
		Scopes(
			filterScope,
			pkg.ApplyPager(in.Pager),
			pkg.ApplySorter(in.Sorters),
		).
		Find(&resMds).Error; err != nil {
		return nil, err
	}

	resources := make([]*rbacpb.Resource, len(resMds))
	for _, resMd := range resMds {
		resources = append(resources, &rbacpb.Resource{
			Ns:   resMd.Ns,
			Name: resMd.Name,
		})
	}

	return &rbacpb.ListResourcesOut{Resources: resources, Total: cnt}, nil
}

func (r *RbacLogicImpl) DeleteResource(c context.Context, in *rbacpb.DeleteResourceIn) error {
	return r.db(c).Transaction(func(tx *gorm.DB) error {
		var res Resource
		if err := tx.Delete(&res, in.Id).Error; err != nil {
			return err
		}

		if err := r.zbLogic.Delete(pkg.WithDB(c, tx),
			&authzpbv1.DeleteTuplesIn{
				Mode: &authzpbv1.DeleteTuplesIn_Filter{
					Filter: &authzpbv1.TupleFilter{
						ObjNs: pkg.Str(res.Ns),
						ObjId: pkg.Str(strconv.FormatUint(in.Id, 10)),
					},
				},
			},
		); err != nil {
			return err
		}

		return nil
	})
}

// TODO: wrap to transaction
func (r *RbacLogicImpl) AssignRole(c context.Context, in *rbacpb.AssignRoleIn) error {
	return r.db(c).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&User{}).Where("id = ?", in.UserId).
			Take(&User{}).Error; err != nil {
			return err
		}

		if err := tx.Model(&Role{}).Where("id = ?", in.RoleId).
			Take(&Role{}).Error; err != nil {
			return err
		}

		return r.zbLogic.Create(pkg.WithDB(c, tx),
			&authzpbv1.Tuple{
				SbjNs: "user",
				SbjId: strconv.FormatUint(in.UserId, 10),
				Rel:   "member",
				ObjNs: "role",
				ObjId: strconv.FormatUint(in.RoleId, 10),
			},
		)
	})
}

func (r *RbacLogicImpl) RevokeRole(c context.Context, in *rbacpb.RevokeRoleIn) error {
	return r.zbLogic.Delete(c,
		&authzpbv1.DeleteTuplesIn{
			Mode: &authzpbv1.DeleteTuplesIn_DeleteTuples{
				DeleteTuples: &authzpbv1.DeleteTuples{
					Tuples: []*authzpbv1.Tuple{
						{
							SbjNs: "user",
							SbjId: strconv.FormatUint(in.UserId, 10),
							Rel:   "member",
							ObjNs: "role",
							ObjId: strconv.FormatUint(in.RoleId, 10),
						},
					},
				},
			},
		},
	)
}

func (r *RbacLogicImpl) GrantPerm(c context.Context, in *rbacpb.GrantPermIn) error {
	return r.db(c).Transaction(func(tx *gorm.DB) error {
		res := Resource{}
		if err := tx.Where("id = ?", in.ResourceId).Take(&res).Error; err != nil {
			return err
		}

		return r.zbLogic.Create(pkg.WithDB(c, tx),
			&authzpbv1.Tuple{
				SbjNs: "role",
				SbjId: strconv.FormatUint(in.RoleId, 10),
				Rel:   in.Perm,
				ObjNs: res.Ns,
				ObjId: strconv.FormatUint(in.ResourceId, 10),
			},
		)
	})
}

func (r *RbacLogicImpl) RevokePerm(c context.Context, in *rbacpb.RevokePermIn) error {
	res := Resource{}
	if err := r.db(c).Where("id = ?", in.ResourceId).Take(&res).Error; err != nil {
		return err
	}

	return r.zbLogic.Delete(c,
		&authzpbv1.DeleteTuplesIn{
			Mode: &authzpbv1.DeleteTuplesIn_DeleteTuples{
				DeleteTuples: &authzpbv1.DeleteTuples{
					Tuples: []*authzpbv1.Tuple{
						{
							SbjNs: "role",
							SbjId: strconv.FormatUint(in.RoleId, 10),
							Rel:   in.Perm,
							ObjNs: res.Ns,
							ObjId: strconv.FormatUint(in.ResourceId, 10),
						},
					},
				},
			},
		},
	)
}

func (r *RbacLogicImpl) GetRole(c context.Context, in *rbacpb.GetRoleIn) (*rbacpb.GetRoleOut, error) {
	var role Role
	if err := r.db(c).Where("id = ?", in.Id).Take(&role).Error; err != nil {
		return nil, err
	}

	out := &rbacpb.GetRoleOut{
		Ns:   "role",
		Name: role.Name,
	}

	perms, err := r.zbLogic.GetPermissions(c, &authzpbv1.Instance{
		Ns: "role",
		Id: strconv.FormatUint(in.Id, 10),
	}, "resource")
	if err != nil {
		return nil, err
	}
	out.Permissions = perms

	return out, nil
}

func (r *RbacLogicImpl) ListResourceTypes(context.Context) (
	*rbacpb.ListResourceTypeOut, error,
) {
	s := r.schema
	resNss := []string{}

	for ns, data := range s.Namespaces {
		if data.Type != "resource" {
			continue
		}
		resNss = append(resNss, ns)
	}

	return &rbacpb.ListResourceTypeOut{Types: resNss}, nil
}

func (r *RbacLogicImpl) ListResourcesByType(c context.Context, in *rbacpb.ListResourcesByTypeIn) (
	*rbacpb.ListResourcesByTypeOut, error,
) {
	out := &rbacpb.ListResourcesByTypeOut{}
	if err := r.db(c).Model(&Resource{}).Where("ns = ?", in.Type).Scan(&out.Resources).Error; err != nil {
		return nil, err
	}

	return out, nil
}

func (r *RbacLogicImpl) ListPermissionByResource(c context.Context, in *rbacpb.ListPermissionByResourceIn) (
	*rbacpb.ListPermissionByResourceOut, error,
) {
	tuples := []*Tuple{}
	if err := r.db(c).Where(
		"sbj_ns = role AND sbj_id = ? AND obj_ns = ? AND obj_id = ?",
		in.RoleId, in.ResourceNs, in.ResourceId).
		Find(&tuples).Error; err != nil {
		return nil, err
	}

	out := &rbacpb.ListPermissionByResourceOut{}
	relDatas := r.schema.Namespaces[in.ResourceNs].Relations
	existedRels := map[string]struct{}{}
	for _, tuple := range tuples {
		existedRels[tuple.Relation] = struct{}{}
	}

	rels := map[string]struct{}{}
	for rel := range relDatas {
		if _, ok := existedRels[rel]; ok {
			continue
		}

		rels[rel] = struct{}{}
	}

	for rel := range rels {
		out.Permissions = append(out.Permissions, rel)
	}

	return out, nil
}

package logic

import (
	"authz/internal/entity/model"
	"authz/internal/pkg"
	"context"
	"strconv"

	rbacpb "github.com/skyrocket-qy/protos/gen/authzpb/rbacpb"
	authzpbv1 "github.com/skyrocket-qy/protos/gen/authzpb/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

/*
Only 4 type
User, Role(include Admin), Perm, resource
*/

type RbacLogic interface {
	ListUsers(c context.Context, in *rbacpb.ListUsersIn) (out *rbacpb.ListUsersOut, err error)
	UpdateUser(c context.Context, in *rbacpb.UpdateUserIn) error
	DeleteUser(c context.Context, id *rbacpb.DeleteUserIn) error

	GetRole(context.Context, *rbacpb.GetRoleIn) (*rbacpb.GetRoleOut, error)
	CreateRole(c context.Context, in *rbacpb.CreateRoleIn) error
	ListRoles(c context.Context, in *rbacpb.ListRolesIn) (*rbacpb.ListRolesOut, error)
	UpdateRole(c context.Context, in *rbacpb.UpdateRoleIn) error
	DeleteRole(c context.Context, in *rbacpb.DeleteRoleIn) error

	ListResourceType(context.Context, *rbacpb.ListResourceTypeIn) (
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
}

func NewRbacLogic(zbLogic ZanzibarLogic, pgdb *gorm.DB) *RbacLogicImpl {
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
	if err := r.db(c).Model(&model.User{}).Scopes(filterScope).Count(&cnt).Error; err != nil {
		return nil, err
	}

	userMds := []*model.User{}
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

		for _, org := range userMd.Orgs {
			user.Orgs = append(user.Orgs, org.Name)
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

	if err := r.db(c).Model(&model.User{}).Where("id = ?", in.Id).Updates(updates).Error; err != nil {
		return err
	}

	return nil
}

func (r *RbacLogicImpl) DeleteUser(c context.Context, in *rbacpb.DeleteUserIn) error {
	return r.db(c).Delete(&model.User{}, in.Id).Error
}

func (r *RbacLogicImpl) CreateRole(c context.Context, in *rbacpb.CreateRoleIn) error {
	return r.db(c).Create(&model.Role{Name: in.Name}).Error
}

func (r *RbacLogicImpl) ListRoles(c context.Context, in *rbacpb.ListRolesIn) (
	out *rbacpb.ListRolesOut, err error,
) {
	roleMds := []*model.Role{}
	if err := r.db(c).Find(&roleMds).Error; err != nil {
		return nil, err
	}

	roles := make([]*rbacpb.Role, 0, len(roleMds))
	for _, roleMd := range roleMds {
		roles = append(roles, &rbacpb.Role{
			Id:   roleMd.Id,
			Name: roleMd.Name,
		})
	}

	return &rbacpb.ListRolesOut{Roles: roles}, nil
}

func (r *RbacLogicImpl) UpdateRole(c context.Context, in *rbacpb.UpdateRoleIn) error {
	return r.db(c).Model(&model.Role{}).Where("id = ?", in.Id).Updates(map[string]any{
		"name": in.Name,
	}).Error
}

func (r *RbacLogicImpl) DeleteRole(c context.Context, in *rbacpb.DeleteRoleIn) error {
	return r.db(c).Delete(&model.Role{}, in.Id).Error
}

func (r *RbacLogicImpl) CreateResource(c context.Context, in *rbacpb.CreateResourceIn) error {
	return r.db(c).Create(&model.Resource{Ns: in.Ns, Name: in.Name}).Error
}

func (r *RbacLogicImpl) ListResources(c context.Context, in *rbacpb.ListResourcesIn) (
	out *rbacpb.ListResourcesOut, err error,
) {
	resMds := []*model.Resource{}
	if err := r.db(c).Find(&resMds).Error; err != nil {
		return nil, err
	}

	resources := make([]*rbacpb.Resource, len(resMds))
	for _, resMd := range resMds {
		resources = append(resources, &rbacpb.Resource{
			Ns:   resMd.Ns,
			Name: resMd.Name,
		})
	}

	return &rbacpb.ListResourcesOut{Resources: resources}, nil
}

// func (r *RbacLogicImpl) UpdateResource(c context.Context, res *rbacpb.Resource) error {
// 	return r.db(c).Model(&model.Resource{}).Where("id = ?", res.Id).Updates(map[string]any{
// 		"ns":   res.Ns,
// 		"name": res.Name,
// 	}).Error
// }

func (r *RbacLogicImpl) DeleteResource(c context.Context, in *rbacpb.DeleteResourceIn) error {
	return r.db(c).Delete(&model.Resource{}, in.Id).Error
}

func (r *RbacLogicImpl) AssignRole(c context.Context, in *rbacpb.AssignRoleIn) error {
	return r.zbLogic.Create(c,
		&authzpbv1.Tuple{
			SbjNs: "user",
			SbjId: strconv.FormatUint(in.UserId, 10),
			Rel:   "member",
			ObjNs: "role",
			ObjId: strconv.FormatUint(in.RoleId, 10),
		},
	)
}

func (r *RbacLogicImpl) RevokeRole(c context.Context, in *rbacpb.RevokeRoleIn) error {
	return r.zbLogic.Delete(c,
		&authzpbv1.DeleteTuplesIn{
			Mode: &authzpbv1.DeleteTuplesIn_Tuples{
				Tuples: &authzpbv1.DeleteTuples{
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
	// res := model.Resource{}
	// if err := r.db(c).Where("id = ?", in.ResourceId).Take(&res).Error; err != nil {
	// 	return err
	// }

	return r.zbLogic.Create(c,
		&authzpbv1.Tuple{
			SbjNs: "role",
			SbjId: strconv.FormatUint(in.RoleId, 10),
			Rel:   in.Perm,
			ObjNs: "object",
			ObjId: strconv.FormatUint(in.ResourceId, 10),
		},
	)
}

func (r *RbacLogicImpl) RevokePerm(c context.Context, in *rbacpb.RevokePermIn) error {
	res := model.Resource{}
	if err := r.db(c).Where("id = ?", in.ResourceId).Take(&res).Error; err != nil {
		return err
	}

	return r.zbLogic.Delete(c,
		&authzpbv1.DeleteTuplesIn{
			Mode: &authzpbv1.DeleteTuplesIn_Tuples{
				Tuples: &authzpbv1.DeleteTuples{
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
	var role model.Role
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

func (r *RbacLogicImpl) ListResourceType(context.Context, *rbacpb.ListResourceTypeIn) (
	*rbacpb.ListResourceTypeOut, error,
) {
	return nil, nil
}

func (r *RbacLogicImpl) ListResourcesByType(context.Context, *rbacpb.ListResourcesByTypeIn) (
	*rbacpb.ListResourcesByTypeOut, error,
) {
	return nil, nil
}

func (r *RbacLogicImpl) ListPermissionByResource(context.Context, *rbacpb.ListPermissionByResourceIn) (
	*rbacpb.ListPermissionByResourceOut, error,
) {
	return nil, nil
}

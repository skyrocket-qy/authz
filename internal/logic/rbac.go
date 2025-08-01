package logic

import (
	"authz/internal/entity"
	"authz/internal/entity/model"
	"authz/internal/pkg"
	"context"
	"strconv"

	authzpbv1 "github.com/skyrocket-qy/protos/gen/authzpb/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

/*
Only 4 type
User, Role(include Admin), Perm, resource
*/

type RbacLogic interface {
	ListUsers(c context.Context, in *authzpbv1.ListUsersIn) (out *authzpbv1.ListUsersOut, err error)
	UpdateUser(c context.Context, in *authzpbv1.UpdateUserIn) error
	DeleteUser(c context.Context, id uint64) error

	CreateRole(c context.Context, name string) error
	ListRoles(c context.Context) ([]*authzpbv1.Role, error)
	UpdateRole(c context.Context, role *authzpbv1.Role) error
	DeleteRole(c context.Context, id uint64) error

	// GetUsersWithRole(ctx context.Context, role string) ([]string, error)
	// GetUsersWithPerm(ctx context.Context, perm string, res string) ([]string, error)

	// ListPerms(c context.Context) ([]*authzpbv1.Perm, error)

	CreateResource(c context.Context, ns, name string) error
	ListResources(c context.Context) ([]*authzpbv1.Resource, error)
	// UpdateResource(c context.Context, in *authzpbv1.Resource) error
	DeleteResource(c context.Context, id uint64) error

	AssignRole(c context.Context, userId uint64, roleId uint64) error
	RevokeRole(c context.Context, userId uint64, roleId uint64) error

	GrantPerm(c context.Context, roleId uint64, perm string, resId uint64) error
	RevokePerm(c context.Context, roleId uint64, perm string, resId uint64) error
}

var _ RbacLogic = (*RbacLogicImpl)(nil)

type RbacLogicImpl struct {
	pgdb    *gorm.DB
	zbLogic ZanzibarLogic
}

func NewRbacLogic(zbLogic ZanzibarLogic) *RbacLogicImpl {
	return &RbacLogicImpl{zbLogic: zbLogic}
}

func (r *RbacLogicImpl) db(c context.Context) *gorm.DB {
	return r.pgdb.WithContext(c)
}

func (r *RbacLogicImpl) ListUsers(c context.Context, in *authzpbv1.ListUsersIn) (
	out *authzpbv1.ListUsersOut, err error,
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

	users := make([]*authzpbv1.User, 0, len(userMds))
	for i, userMd := range userMds {
		user := &authzpbv1.User{
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
	return &authzpbv1.ListUsersOut{
		Users: users,
		Count: cnt,
	}, nil
}

func (r *RbacLogicImpl) UpdateUser(c context.Context, in *authzpbv1.UpdateUserIn) error {
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

func (r *RbacLogicImpl) DeleteUser(c context.Context, id uint64) error {
	return r.db(c).Delete(&model.User{}, id).Error
}

func (r *RbacLogicImpl) CreateRole(c context.Context, name string) error {
	return r.db(c).Create(&model.Role{Name: name}).Error
}

func (r *RbacLogicImpl) ListRoles(c context.Context) ([]*authzpbv1.Role, error) {
	roleMds := []*model.Role{}
	if err := r.db(c).Find(&roleMds).Error; err != nil {
		return nil, err
	}

	roles := make([]*authzpbv1.Role, 0, len(roleMds))
	for _, roleMd := range roleMds {
		roles = append(roles, &authzpbv1.Role{
			Id:   roleMd.Id,
			Name: roleMd.Name,
		})
	}

	return roles, nil
}

func (r *RbacLogicImpl) UpdateRole(c context.Context, in *authzpbv1.Role) error {
	return r.db(c).Model(&model.Role{}).Where("id = ?", in.Id).Updates(map[string]any{
		"name": in.Name,
	}).Error
}

func (r *RbacLogicImpl) DeleteRole(c context.Context, id uint64) error {
	return r.db(c).Delete(&model.Role{}, id).Error
}

func (r *RbacLogicImpl) CreateResource(c context.Context, ns, name string) error {
	return r.db(c).Create(&model.Resource{Ns: ns, Name: name}).Error
}

func (r *RbacLogicImpl) ListResources(c context.Context) ([]*authzpbv1.Resource, error) {
	resMds := []*model.Resource{}
	if err := r.db(c).Find(&resMds).Error; err != nil {
		return nil, err
	}

	resources := make([]*authzpbv1.Resource, 0, len(resMds))
	for _, resMd := range resMds {
		resources = append(resources, &authzpbv1.Resource{
			Ns:   resMd.Ns,
			Name: resMd.Name,
		})
	}

	return resources, nil
}

// func (r *RbacLogicImpl) UpdateResource(c context.Context, res *authzpbv1.Resource) error {
// 	return r.db(c).Model(&model.Resource{}).Where("id = ?", res.Id).Updates(map[string]any{
// 		"ns":   res.Ns,
// 		"name": res.Name,
// 	}).Error
// }

func (r *RbacLogicImpl) DeleteResource(c context.Context, id uint64) error {
	return r.db(c).Delete(&model.Resource{}, id).Error
}

func (r *RbacLogicImpl) AssignRole(c context.Context, userId uint64, roleId uint64) error {
	return r.zbLogic.Create(c,
		&authzpbv1.Tuple{
			SbjNs: "user",
			SbjId: strconv.FormatUint(userId, 64),
			Rel:   "member",
			ObjNs: "role",
			ObjId: strconv.FormatUint(roleId, 64),
		},
	)
}

func (r *RbacLogicImpl) RevokeRole(c context.Context, userId uint64, roleId uint64) error {
	return r.zbLogic.Delete(c,
		&entity.TupleFilter{
			SbjNs:    pkg.Str("user"),
			SbjId:    pkg.Str(strconv.FormatUint(userId, 64)),
			Relation: pkg.Str("member"),
			ObjNs:    pkg.Str("role"),
			ObjId:    pkg.Str(strconv.FormatUint(roleId, 64)),
		},
	)
}

func (r *RbacLogicImpl) GrantPerm(c context.Context, roleId uint64, perm string, resId uint64) error {
	res := model.Resource{}
	if err := r.db(c).Where("id = ?", resId).Take(&res).Error; err != nil {
		return err
	}

	return r.zbLogic.Create(c,
		&authzpbv1.Tuple{
			SbjNs: "role",
			SbjId: strconv.FormatUint(roleId, 64),
			Rel:   perm,
			ObjNs: res.Ns,
			ObjId: strconv.FormatUint(resId, 64),
		},
	)
}

func (r *RbacLogicImpl) RevokePerm(c context.Context, roleId uint64, perm string, resId uint64) error {
	res := model.Resource{}
	if err := r.db(c).Where("id = ?", resId).Take(&res).Error; err != nil {
		return err
	}

	return r.zbLogic.Delete(c,
		&entity.TupleFilter{
			SbjNs:    pkg.Str("role"),
			SbjId:    pkg.Str(strconv.FormatUint(roleId, 64)),
			Relation: &perm,
			ObjNs:    &res.Ns,
			ObjId:    pkg.Str(strconv.FormatUint(resId, 64)),
		},
	)
}

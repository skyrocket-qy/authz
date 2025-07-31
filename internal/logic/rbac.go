package logic

import (
	"authz/internal/entity"
	"authz/internal/entity/model"
	"context"

	authzpbv1 "github.com/skyrocket-qy/protos/gen/authzpb/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
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
	ListUsers(c context.Context) ([]*authzpbv1.User, error)
	UpdateUser(c context.Context, in *authzpbv1.UpdateUserIn) error
	DeleteUser(c context.Context, id uint64) error

	CreateRole(c context.Context, name string) error
	ListRoles(c context.Context) ([]*authzpbv1.Role, error)
	UpdateRole(c context.Context, role *authzpbv1.Role) error
	DeleteRole(c context.Context, id uint64) error

	GetUsersWithRole(ctx context.Context, role string) ([]string, error)
	GetUsersWithPerm(ctx context.Context, perm string, res string) ([]string, error)

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
	pgdb    *gorm.DB
	zbLogic ZanzibarLogic
}

func NewRbacLogic(zbLogic ZanzibarLogic) *RbacLogicImpl {
	return &RbacLogicImpl{zbLogic: zbLogic}
}

func (r *RbacLogicImpl) db(c context.Context) *gorm.DB {
	return r.pgdb.WithContext(c)
}

// TODO: add listoptions
func (r *RbacLogicImpl) ListUsers(c context.Context) ([]*authzpbv1.User, error) {
	userMds := []*model.User{}
	if err := r.db(c).Preload("Orgs").Preload("UserAuths").Find(&userMds).Error; err != nil {
		return nil, err
	}

	users := make([]*authzpbv1.User, 0, len(userMds))
	for _, userMd := range userMds {
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
			user.AuthTypes = append(user.AuthTypes, userAuth.AuthType)
		}
	}
	return users, nil
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

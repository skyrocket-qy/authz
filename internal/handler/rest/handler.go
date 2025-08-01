package handler

import (
	"authz/internal/logic"
	"net/http"

	"github.com/gin-gonic/gin"
	authzpbv1 "github.com/skyrocket-qy/protos/gen/authzpb/v1"
)

type Handler struct {
	zLogic    logic.ZanzibarLogic
	rbacLogic logic.RbacLogic
}

func NewHandler(zLogic logic.ZanzibarLogic, rbacLogic logic.RbacLogic) *Handler {
	return &Handler{
		zLogic:    zLogic,
		rbacLogic: rbacLogic,
	}
}

// @Success	200
// @Router		/v1/ping [get]
// @Tags		general.
func (h *Handler) Ping(c *gin.Context) {
	c.Status(http.StatusOK)
}

// @Tags		general
// @Success	200	{object}	erx.APIError
// @Router		/v1/healthy [get].
func (d *Handler) Healthy(c *gin.Context) {
	// err := d.logic.Healthy(c)
	// if err != nil {
	// 	pkg.Bind(c, erx.W(err))
	// 	return
	// }

	c.Status(http.StatusOK)
}

func (d *Handler) ListUsers(c *gin.Context) ([]*authzpbv1.User, error) {
	return d.rbacLogic.ListUsers(c)
}

func (d *Handler) UpdateUser(c *gin.Context, in *authzpbv1.UpdateUserIn) error {
	return d.rbacLogic.UpdateUser(c, in)
}

func (d *Handler) DeleteUser(c *gin.Context) error {
	type Req struct {
		Id uint64 `json:"id"`
	}

	var req Req
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}

	return d.rbacLogic.DeleteUser(c, req.Id)
}

func (d *Handler) CreateRole(c *gin.Context, name string) error {
	return d.rbacLogic.CreateRole(c, name)
}

func (d *Handler) ListRoles(c *gin.Context) ([]*authzpbv1.Role, error) {
	return d.rbacLogic.ListRoles(c)
}

func (d *Handler) UpdateRole(c *gin.Context, role *authzpbv1.Role) error {
	return d.rbacLogic.UpdateRole(c, role)
}

func (d *Handler) DeleteRole(c *gin.Context, id uint64) error {
	return d.rbacLogic.DeleteRole(c, id)
}

func (d *Handler) CreateResource(c *gin.Context, ns, name string) error {
	return d.rbacLogic.CreateResource(c, ns, name)
}

func (d *Handler) ListResources(c *gin.Context) ([]*authzpbv1.Resource, error) {
	return d.rbacLogic.ListResources(c)
}

func (d *Handler) DeleteResource(c *gin.Context, id uint64) error {
	return d.rbacLogic.DeleteResource(c, id)
}

func (d *Handler) AssignRole(c *gin.Context, userId uint64, roleId uint64) error {
	return d.rbacLogic.AssignRole(c, userId, roleId)
}

func (d *Handler) RevokeRole(c *gin.Context, userId uint64, roleId uint64) error {
	return d.rbacLogic.RevokeRole(c, userId, roleId)
}

func (d *Handler) GrantPerm(c *gin.Context, roleId uint64, perm string, resId uint64) error {
	return d.rbacLogic.GrantPerm(c, roleId, perm, resId)
}

func (d *Handler) RevokePerm(c *gin.Context, roleId uint64, perm string, resId uint64) error {
	return d.rbacLogic.RevokePerm(c, roleId, perm, resId)
}

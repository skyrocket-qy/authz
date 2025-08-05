package rest

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

func (d *Handler) ListUsers(c *gin.Context, in *authzpbv1.ListUsersIn) (*authzpbv1.ListUsersOut, error) {
	return d.rbacLogic.ListUsers(c, in)
}

func (d *Handler) UpdateUser(c *gin.Context, in *authzpbv1.UpdateUserIn) error {
	return d.rbacLogic.UpdateUser(c, in)
}

func (d *Handler) DeleteUser(c *gin.Context, in *authzpbv1.DeleteUserIn) error {
	return d.rbacLogic.DeleteUser(c, in)
}

func (d *Handler) CreateRole(c *gin.Context, in *authzpbv1.CreateRoleIn) error {
	return d.rbacLogic.CreateRole(c, in)
}

func (d *Handler) ListRoles(c *gin.Context, in *authzpbv1.ListRolesIn) (*authzpbv1.ListRolesOut, error) {
	return d.rbacLogic.ListRoles(c, in)
}

func (d *Handler) UpdateRole(c *gin.Context, in *authzpbv1.UpdateRoleIn) error {
	return d.rbacLogic.UpdateRole(c, in)
}

func (d *Handler) DeleteRole(c *gin.Context, in *authzpbv1.DeleteRoleIn) error {
	return d.rbacLogic.DeleteRole(c, in)
}

func (d *Handler) CreateResource(c *gin.Context, in *authzpbv1.CreateResourceIn) error {
	return d.rbacLogic.CreateResource(c, in)
}

func (d *Handler) ListResources(c *gin.Context, in *authzpbv1.ListResourcesIn) (
	*authzpbv1.ListResourcesOut, error,
) {
	return d.rbacLogic.ListResources(c, in)
}

func (d *Handler) DeleteResource(c *gin.Context, in *authzpbv1.DeleteResourceIn) error {
	return d.rbacLogic.DeleteResource(c, in)
}

func (d *Handler) AssignRole(c *gin.Context, in *authzpbv1.AssignRoleIn) error {
	return d.rbacLogic.AssignRole(c, in)
}

func (d *Handler) RevokeRole(c *gin.Context, in *authzpbv1.RevokeRoleIn) error {
	return d.rbacLogic.RevokeRole(c, in)
}

func (d *Handler) GrantPerm(c *gin.Context, in *authzpbv1.GrantPermIn) error {
	return d.rbacLogic.GrantPerm(c, in)
}

func (d *Handler) RevokePerm(c *gin.Context, in *authzpbv1.RevokePermIn) error {
	return d.rbacLogic.RevokePerm(c, in)
}

func (d *Handler) CreateTuple(c *gin.Context, tuple *authzpbv1.Tuple) error {
	return d.zLogic.Create(c, tuple)
}

func (d *Handler) FindTuples(c *gin.Context, filter *authzpbv1.TupleFilter) (
	[]*authzpbv1.Tuple, error,
) {
	return d.zLogic.Find(c, filter)
}

func (d *Handler) ListTuples(c *gin.Context, in *authzpbv1.ListTuplesIn) (
	*authzpbv1.ListTuplesOut, error,
) {
	return d.zLogic.List(c, in)
}

func (d *Handler) DeleteTuples(c *gin.Context, filter *authzpbv1.TupleFilter) error {
	return d.zLogic.Delete(c, filter)
}

package api

import (
	"authz/internal/handler/rest"
	"authz/internal/handler/rest/middleware"
	"authz/internal/pkg"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/skyrocket-qy/erx"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func RegisterAPIHandlers(r *gin.Engine, h *rest.Handler, checkAuth gin.HandlerFunc) {
	r.Use(middleware.Cors())
	r.Use(middleware.ErrorHttp)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	vr := r.Group("/v1")
	{
		vr.GET("/ping", h.Ping)
		vr.GET("/healthy", h.Healthy)
	}
	pR := r.Group("/")
	// pR.Use(checkAuth)
	vpr := pR.Group("/v1")
	{
		vpr.POST("/list-users", Hdl(h.ListUsers))
		vpr.PUT("/user", HdlNoOut(h.UpdateUser))
		vpr.DELETE("/user", HdlNoOut(h.DeleteUser))
		vpr.POST("/list-roles", Hdl(h.ListRoles))
		vpr.POST("/role", HdlNoOut(h.CreateRole))
		vpr.PUT("/role", HdlNoOut(h.UpdateRole))
		vpr.DELETE("/role", HdlNoOut(h.DeleteRole))
		vpr.POST("/list-resources", Hdl(h.ListResources))
		vpr.POST("/resource", HdlNoOut(h.CreateResource))
		vpr.DELETE("/resource", HdlNoOut(h.DeleteResource))
		vpr.POST("/assign-role", HdlNoOut(h.AssignRole))
		vpr.POST("/revoke-role", HdlNoOut(h.RevokeRole))
		vpr.POST("/grant-perm", HdlNoOut(h.GrantPerm))
		vpr.POST("/revoke-perm", HdlNoOut(h.RevokePerm))

		tr := vpr.Group("/tuples")
		{
			tr.POST("/list", Hdl(h.ListTuples))
			tr.POST("/", HdlNoOut(h.CreateTuple))
			tr.DELETE("/", HdlNoOut(h.DeleteTuples))
		}
	}
}

func Hdl[Req, Resp proto.Message](a func(c *gin.Context, req Req) (resp Resp, err error)) func(
	*gin.Context,
) {
	return func(c *gin.Context) {
		var req Req
		// TODO: this method is not performant, but for generic wrapper, it seems the only way
		req = reflect.New(reflect.TypeOf(req).Elem()).Interface().(Req)

		if !ShouldBindProto(c, req) {
			return
		}

		out, err := a(c, req)
		if err != nil {
			pkg.Bind(c, err)
			return
		}

		jsonBytes, err := protojson.Marshal(out)
		if err != nil {
			pkg.Bind(c, err)
			return
		}

		c.Status(http.StatusOK)
		c.Writer.Write(jsonBytes)
	}
}

func HdlNoOut[Req proto.Message](a func(c *gin.Context, req Req) error) func(*gin.Context) {
	return func(c *gin.Context) {
		req := any(new(Req)).(Req)
		if !ShouldBindProto(c, req) {
			return
		}

		if err := a(c, req); err != nil {
			pkg.Bind(c, err)
			return
		}

		c.Status(http.StatusOK)
	}

}

func ShouldBindProto(c *gin.Context, req proto.Message) bool {
	data, err := c.GetRawData()
	if err != nil {
		pkg.Bind(c, erx.W(err).SetCode(pkg.ErrBadRequest))
		return false
	}

	if err := protojson.Unmarshal(data, req); err != nil {
		pkg.Bind(c, erx.W(err).SetCode(pkg.ErrBadRequest))
		return false
	}

	return true
}

package api

import (
	"authz/internal/handler/rest/middleware"
	"authz/internal/pkg"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/skyrocket-qy/erx"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	pR.Use(checkAuth)

	vpr := pR.Group("/v1")
	{
	}
}

func Hdl[Req, Resp any](a func(c *gin.Context, req Req) (resp Resp, err error)) func(*gin.Context) {
	return func(c *gin.Context) {
		var req Req
		if !ShouldBindJson(c, &req) {
			return
		}

		out, err := a(c, req)
		if err != nil {
			pkg.Bind(c, err)
			return
		}

		c.JSON(http.StatusOK, out)
	}

}

func HdlNoOut[Req any](a func(c *gin.Context, req Req) error) func(*gin.Context) {
	return func(c *gin.Context) {
		var req Req
		if !ShouldBindJson(c, &req) {
			return
		}

		if err := a(c, req); err != nil {
			pkg.Bind(c, err)
			return
		}

		c.Status(http.StatusOK)
	}

}

func ShouldBindJson(c *gin.Context, req any) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		pkg.Bind(c, erx.W(err).SetCode(pkg.ErrBadRequest))
		return false
	}
	return true
}

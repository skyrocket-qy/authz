package handler

import (
	"net/http"

	"srv/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/skyrocket-qy/erx"
)

// @Tags		general
// @Success	200	{object}	erx.APIError
// @Router		/v1/healthy [get].
func (d *Handler) Healthy(c *gin.Context) {
	err := d.logic.Healthy(c)
	if err != nil {
		pkg.Bind(c, erx.W(err))
		return
	}

	c.Status(http.StatusOK)
}

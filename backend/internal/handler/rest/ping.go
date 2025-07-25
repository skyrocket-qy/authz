package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Success	200
// @Router		/v1/ping [get]
// @Tags		general.
func (h *Handler) Ping(c *gin.Context) {
	c.Status(http.StatusOK)
}

package handler

import (
	"authz/internal/logic"
)

type Handler struct {
	logic logic.ZanzibarLogic
}

func NewHandler(logic logic.ZanzibarLogic) *Handler {
	return &Handler{
		logic: logic,
	}
}

// @Tags		auth
// @Param		user	body		logic.LoginIn	true	"request body"
// @Success	200		{object}	logic.LoginOut
// @Router		/v1/login [post].
// func (h *Handler) Login(c *gin.Context, req *logic.LoginIn) (*logic.LoginOut, error) {
// 	return h.logic.Login(c.Request.Context(), req)
// }

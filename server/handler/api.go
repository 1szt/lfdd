package handler

import (
	"github.com/gin-gonic/gin"
)

type ApiHandler struct {
}

// API API 根路径探活
func (h *ApiHandler) API(c *gin.Context) {
	c.String(200, "Hello, API!")
}

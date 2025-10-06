package router

import (
	"github.com/cloudwego/hertz/pkg/app/server"
)

func Register(h *server.Hertz) {
	h.POST("/register", nil)
	h.POST("/login", nil)
	h.GET("/key", nil)
	h.GET("/keys", nil)
	h.DELETE("/key", nil)
}

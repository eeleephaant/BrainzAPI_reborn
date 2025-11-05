package router

import (
	"brainz/developersapi/internal/handler"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func Register(h *server.Hertz) {
	h.POST("/register", handler.Register)
	h.POST("/login", handler.Login)
	h.GET("/key", nil)
	h.GET("/keys", nil)
	h.DELETE("/key", nil)

}

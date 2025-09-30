package router

import (
	"brainz-api/internal/handler"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func Register(h *server.Hertz) {
	h.GET("/lessons", handler.GetLessons)
	h.GET("/lessons/count", handler.GetLessonsCount)
	h.GET("/institution", handler.GetInstitutions)
	h.POST("/institution", handler.AddInstitution)
	h.GET("/group", handler.GetGroups)
}

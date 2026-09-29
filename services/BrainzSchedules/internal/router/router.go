package router

import (
	"brainz-api/internal/handler"
	"brainz-api/internal/middleware"
	"context"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func Register(ctx context.Context,
	h *server.Hertz,
	sh *handler.ScheduleHandler,
	ih *handler.InstitutionHandler,
	gh *handler.GroupHandler,
) {
	h.Use(middleware.RecoveryMiddleware())

	// REST
	h.GET("/lessons", sh.GetLessons)
	h.POST("/lessons", sh.AddLessons)
	h.PUT("/lessons/timings", sh.UpdateLessonTimingsForDate)
	h.GET("/institution", ih.GetInstitutions)
	h.POST("/institution", ih.CreateInstitution)
	h.GET("/group", gh.GetGroupsForInst)

	// WEB-SOCKET
	h.GET("/info-stream", handler.InfoStreamHandler)
}

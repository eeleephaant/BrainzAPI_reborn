package router

import (
	"brainz-api/internal/handler"
	"context"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func Register(ctx context.Context,
	h *server.Hertz,
	sh *handler.ScheduleHandler,
	ih *handler.InstitutionHandler,
	gh *handler.GroupHandler,
) {
	h.GET("/lessons", sh.GetLessons)
	h.GET("/institution", ih.GetInstitutions)
	h.GET("/group", gh.GetGroupsForInst)
	h.GET("/ws/info_stream", handler.InfoStreamHandler)
}

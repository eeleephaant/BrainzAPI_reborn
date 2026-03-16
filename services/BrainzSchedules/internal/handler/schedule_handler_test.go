package handler

import (
	"brainz-api/internal/dtos"
	"brainz/common/permissions"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/ut"
)

type mockLessonService struct {
	deleteLessonFn          func(ctx context.Context, lessonID int) error
	addLessonsFn            func(ctx context.Context, dto dtos.LessonsCreateDTO) error
	getForDateInstitutionFn func(ctx context.Context, date time.Time, institutionID uint64) ([]dtos.LessonDto, error)
}

func (m *mockLessonService) DeleteLesson(ctx context.Context, lessonID int) error {
	return m.deleteLessonFn(ctx, lessonID)
}

func (m *mockLessonService) AddLessons(ctx context.Context, dto dtos.LessonsCreateDTO) error {
	return m.addLessonsFn(ctx, dto)
}

func (m *mockLessonService) GetForDateAndInstitution(ctx context.Context, date time.Time, institutionID uint64) ([]dtos.LessonDto, error) {
	return m.getForDateInstitutionFn(ctx, date, institutionID)
}

func TestScheduleHandler_GetLessons(t *testing.T) {
	ctx := context.Background()

	t.Run("missing_institution_id", func(t *testing.T) {
		h := &ScheduleHandler{}
		c := ut.CreateUtRequestContext("GET", "/lessons?date=2026-03-14", nil)
		h.GetLessons(ctx, c)
		if c.Response.StatusCode() != 400 {
			t.Fatalf("status = %d, want 400", c.Response.StatusCode())
		}
	})

	t.Run("invalid_institution_id", func(t *testing.T) {
		h := &ScheduleHandler{}
		c := ut.CreateUtRequestContext("GET", "/lessons?institution_id=bad&date=2026-03-14", nil)
		h.GetLessons(ctx, c)
		if c.Response.StatusCode() != 400 {
			t.Fatalf("status = %d, want 400", c.Response.StatusCode())
		}
	})

	t.Run("authorize_error", func(t *testing.T) {
		h := &ScheduleHandler{
			as: &mockAuthService{
				authorizeFn: func(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
					return false, errors.New("auth failed")
				},
			},
		}
		c := ut.CreateUtRequestContext("GET", "/lessons?institution_id=1&date=2026-03-14", nil)
		h.GetLessons(ctx, c)
		if c.Response.StatusCode() != 500 {
			t.Fatalf("status = %d, want 500", c.Response.StatusCode())
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		h := &ScheduleHandler{
			as: &mockAuthService{
				authorizeFn: func(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
					return false, nil
				},
			},
		}
		c := ut.CreateUtRequestContext("GET", "/lessons?institution_id=1&date=2026-03-14", nil)
		h.GetLessons(ctx, c)
		if c.Response.StatusCode() != 403 {
			t.Fatalf("status = %d, want 403", c.Response.StatusCode())
		}
	})

	t.Run("invalid_date", func(t *testing.T) {
		h := &ScheduleHandler{
			as: &mockAuthService{
				authorizeFn: func(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
					return true, nil
				},
			},
		}
		c := ut.CreateUtRequestContext("GET", "/lessons?institution_id=1&date=bad-date", nil)
		h.GetLessons(ctx, c)
		if c.Response.StatusCode() != 400 {
			t.Fatalf("status = %d, want 400", c.Response.StatusCode())
		}
	})

	t.Run("service_error", func(t *testing.T) {
		h := &ScheduleHandler{
			as: &mockAuthService{
				authorizeFn: func(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
					return true, nil
				},
			},
			ls: &mockLessonService{
				getForDateInstitutionFn: func(ctx context.Context, date time.Time, institutionID uint64) ([]dtos.LessonDto, error) {
					return nil, errors.New("list failed")
				},
			},
		}
		c := ut.CreateUtRequestContext("GET", "/lessons?institution_id=1&date=2026-03-14", nil)
		h.GetLessons(ctx, c)
		if c.Response.StatusCode() != 500 {
			t.Fatalf("status = %d, want 500", c.Response.StatusCode())
		}
	})

	t.Run("success", func(t *testing.T) {
		h := &ScheduleHandler{
			as: &mockAuthService{
				authorizeFn: func(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
					return true, nil
				},
			},
			ls: &mockLessonService{
				getForDateInstitutionFn: func(ctx context.Context, date time.Time, institutionID uint64) ([]dtos.LessonDto, error) {
					return []dtos.LessonDto{{Id: 1, Name: "Math"}}, nil
				},
			},
		}
		c := ut.CreateUtRequestContext("GET", "/lessons?institution_id=1&date=2026-03-14", nil)
		h.GetLessons(ctx, c)
		if c.Response.StatusCode() != 200 {
			t.Fatalf("status = %d, want 200", c.Response.StatusCode())
		}
		if !bytes.Contains(c.Response.Body(), []byte("Math")) {
			t.Fatalf("body = %s, want Math", c.Response.Body())
		}
	})
}

func TestScheduleHandler_AddLessons(t *testing.T) {
	ctx := context.Background()
	start := time.Date(2026, 3, 14, 9, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)

	makeCtx := func(body any) *app.RequestContext {
		payload, _ := json.Marshal(body)
		return ut.CreateUtRequestContext("POST", "/lessons", &ut.Body{Body: bytes.NewReader(payload), Len: len(payload)},
			ut.Header{Key: "Content-Type", Value: "application/json"})
	}

	validBody := map[string]any{
		"institutionID": 1,
		"lessons": []map[string]any{
			{
				"name":         "Math",
				"cab_num":      "101",
				"teacher_name": "Teacher",
				"start_time":   start.Format(time.RFC3339),
				"end_time":     end.Format(time.RFC3339),
				"num":          1,
				"group_id":     1,
			},
		},
	}

	t.Run("invalid_body", func(t *testing.T) {
		h := &ScheduleHandler{}
		c := makeCtx(map[string]any{})
		h.AddLessons(ctx, c)
		if c.Response.StatusCode() != 400 {
			t.Fatalf("status = %d, want 400", c.Response.StatusCode())
		}
	})

	t.Run("authorize_error", func(t *testing.T) {
		h := &ScheduleHandler{
			as: &mockAuthService{
				authorizeFn: func(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
					return false, errors.New("auth failed")
				},
			},
		}
		c := makeCtx(validBody)
		h.AddLessons(ctx, c)
		if c.Response.StatusCode() != 500 {
			t.Fatalf("status = %d, want 500", c.Response.StatusCode())
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		h := &ScheduleHandler{
			as: &mockAuthService{
				authorizeFn: func(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
					return false, nil
				},
			},
		}
		c := makeCtx(validBody)
		h.AddLessons(ctx, c)
		if c.Response.StatusCode() != 403 {
			t.Fatalf("status = %d, want 403", c.Response.StatusCode())
		}
	})

	t.Run("service_error", func(t *testing.T) {
		h := &ScheduleHandler{
			as: &mockAuthService{
				authorizeFn: func(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
					return true, nil
				},
			},
			ls: &mockLessonService{
				addLessonsFn: func(ctx context.Context, dto dtos.LessonsCreateDTO) error {
					return errors.New("add failed")
				},
			},
		}
		c := makeCtx(validBody)
		h.AddLessons(ctx, c)
		if c.Response.StatusCode() != 500 {
			t.Fatalf("status = %d, want 500", c.Response.StatusCode())
		}
	})

	t.Run("success", func(t *testing.T) {
		h := &ScheduleHandler{
			as: &mockAuthService{
				authorizeFn: func(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
					return true, nil
				},
			},
			ls: &mockLessonService{
				addLessonsFn: func(ctx context.Context, dto dtos.LessonsCreateDTO) error {
					return nil
				},
			},
		}
		c := makeCtx(validBody)
		h.AddLessons(ctx, c)
		if c.Response.StatusCode() != 201 {
			t.Fatalf("status = %d, want 201", c.Response.StatusCode())
		}
	})
}

func TestScheduleHandler_DeleteLesson(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid_lesson_id", func(t *testing.T) {
		h := &ScheduleHandler{}
		c := ut.CreateUtRequestContext("DELETE", "/lessons?lesson_id=bad", nil)
		h.DeleteLesson(ctx, c)
		if c.Response.StatusCode() != 400 {
			t.Fatalf("status = %d, want 400", c.Response.StatusCode())
		}
	})

	t.Run("authorize_error", func(t *testing.T) {
		h := &ScheduleHandler{
			as: &mockAuthService{
				authorizeFn: func(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
					return false, errors.New("auth failed")
				},
			},
		}
		c := ut.CreateUtRequestContext("DELETE", "/lessons?lesson_id=1", nil)
		h.DeleteLesson(ctx, c)
		if c.Response.StatusCode() != 500 {
			t.Fatalf("status = %d, want 500", c.Response.StatusCode())
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		h := &ScheduleHandler{
			as: &mockAuthService{
				authorizeFn: func(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
					return false, nil
				},
			},
		}
		c := ut.CreateUtRequestContext("DELETE", "/lessons?lesson_id=1", nil)
		h.DeleteLesson(ctx, c)
		if c.Response.StatusCode() != 403 {
			t.Fatalf("status = %d, want 403", c.Response.StatusCode())
		}
	})

	t.Run("service_error", func(t *testing.T) {
		h := &ScheduleHandler{
			as: &mockAuthService{
				authorizeFn: func(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
					return true, nil
				},
			},
			ls: &mockLessonService{
				deleteLessonFn: func(ctx context.Context, lessonID int) error {
					return errors.New("delete failed")
				},
			},
		}
		c := ut.CreateUtRequestContext("DELETE", "/lessons?lesson_id=1", nil)
		h.DeleteLesson(ctx, c)
		if c.Response.StatusCode() != 500 {
			t.Fatalf("status = %d, want 500", c.Response.StatusCode())
		}
	})

	t.Run("success", func(t *testing.T) {
		h := &ScheduleHandler{
			as: &mockAuthService{
				authorizeFn: func(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
					return true, nil
				},
			},
			ls: &mockLessonService{
				deleteLessonFn: func(ctx context.Context, lessonID int) error {
					return nil
				},
			},
		}
		c := ut.CreateUtRequestContext("DELETE", "/lessons?lesson_id=1", nil)
		h.DeleteLesson(ctx, c)
		if c.Response.StatusCode() != 200 {
			t.Fatalf("status = %d, want 200", c.Response.StatusCode())
		}
	})
}

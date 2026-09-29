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
	getLessonInstitutionFn  func(ctx context.Context, lessonID int64) (int64, error)
	updateTimingsFn         func(ctx context.Context, dto dtos.DateTimeChangeDTO) error
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

func (m *mockLessonService) GetLessonInstitutionID(ctx context.Context, lessonID int64) (int64, error) {
	return m.getLessonInstitutionFn(ctx, lessonID)
}

func (m *mockLessonService) UpdateTimingsForDateAndInstitution(ctx context.Context, dto dtos.DateTimeChangeDTO) error {
	return m.updateTimingsFn(ctx, dto)
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

	t.Run("empty_lessons", func(t *testing.T) {
		h := &ScheduleHandler{
			as: &mockAuthService{
				authorizeFn: func(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
					t.Fatal("authorize should not be called for invalid payload")
					return false, nil
				},
			},
			ls: &mockLessonService{
				addLessonsFn: func(ctx context.Context, dto dtos.LessonsCreateDTO) error {
					t.Fatal("lesson service should not be called for invalid payload")
					return nil
				},
			},
		}
		c := makeCtx(map[string]any{
			"institutionID": 1,
			"lessons":       []map[string]any{},
		})
		h.AddLessons(ctx, c)
		if c.Response.StatusCode() != 400 {
			t.Fatalf("status = %d, want 400", c.Response.StatusCode())
		}
	})

	t.Run("overlapping_lessons_same_group", func(t *testing.T) {
		h := &ScheduleHandler{
			as: &mockAuthService{
				authorizeFn: func(ctx context.Context, reqCtx *app.RequestContext, apiKey string, perm *permissions.Permission) (bool, error) {
					t.Fatal("authorize should not be called for invalid payload")
					return false, nil
				},
			},
			ls: &mockLessonService{
				addLessonsFn: func(ctx context.Context, dto dtos.LessonsCreateDTO) error {
					t.Fatal("lesson service should not be called for invalid payload")
					return nil
				},
			},
		}

		overlapBody := map[string]any{
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
				{
					"name":         "Physics",
					"cab_num":      "102",
					"teacher_name": "Teacher",
					"start_time":   start.Add(30 * time.Minute).Format(time.RFC3339),
					"end_time":     end.Add(30 * time.Minute).Format(time.RFC3339),
					"num":          2,
					"group_id":     1,
				},
			},
		}

		c := makeCtx(overlapBody)
		h.AddLessons(ctx, c)
		if c.Response.StatusCode() != 400 {
			t.Fatalf("status = %d, want 400", c.Response.StatusCode())
		}
		if !bytes.Contains(c.Response.Body(), []byte("overlap")) {
			t.Fatalf("body = %s, want overlap error", c.Response.Body())
		}
	})

	t.Run("adjacent_lessons_same_group_ok", func(t *testing.T) {
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

		adjacentBody := map[string]any{
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
				{
					"name":         "Physics",
					"cab_num":      "102",
					"teacher_name": "Teacher",
					"start_time":   end.Format(time.RFC3339),
					"end_time":     end.Add(time.Hour).Format(time.RFC3339),
					"num":          2,
					"group_id":     1,
				},
			},
		}

		c := makeCtx(adjacentBody)
		h.AddLessons(ctx, c)
		if c.Response.StatusCode() != 201 {
			t.Fatalf("status = %d, want 201", c.Response.StatusCode())
		}
	})

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
			ls: &mockLessonService{
				getLessonInstitutionFn: func(ctx context.Context, lessonID int64) (int64, error) {
					return 10, nil
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
			ls: &mockLessonService{
				getLessonInstitutionFn: func(ctx context.Context, lessonID int64) (int64, error) {
					return 10, nil
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
				getLessonInstitutionFn: func(ctx context.Context, lessonID int64) (int64, error) {
					return 10, nil
				},
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
				getLessonInstitutionFn: func(ctx context.Context, lessonID int64) (int64, error) {
					return 10, nil
				},
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

func TestScheduleHandler_UpdateLessonTimingsForDate(t *testing.T) {
	ctx := context.Background()
	makeCtx := func(body any) *app.RequestContext {
		payload, _ := json.Marshal(body)
		return ut.CreateUtRequestContext("PUT", "/lessons/timings", &ut.Body{Body: bytes.NewReader(payload), Len: len(payload)},
			ut.Header{Key: "Content-Type", Value: "application/json"})
	}

	validBody := map[string]any{
		"institution_id": 1,
		"date":           "2026-03-20T00:00:00Z",
		"lesson_timings": map[string]any{
			"lesson_1": map[string]any{"start_time": "0001-01-01T08:30:00Z", "end_time": "0001-01-01T10:00:00Z"},
			"lesson_2": map[string]any{"start_time": "0001-01-01T10:10:00Z", "end_time": "0001-01-01T11:40:00Z"},
			"lesson_3": map[string]any{"start_time": "0001-01-01T12:00:00Z", "end_time": "0001-01-01T13:30:00Z"},
			"lesson_4": map[string]any{"start_time": "0001-01-01T13:40:00Z", "end_time": "0001-01-01T15:10:00Z"},
			"lesson_5": map[string]any{"start_time": "0001-01-01T15:20:00Z", "end_time": "0001-01-01T16:50:00Z"},
			"lesson_6": map[string]any{"start_time": "0001-01-01T17:00:00Z", "end_time": "0001-01-01T18:30:00Z"},
		},
	}

	t.Run("invalid_body", func(t *testing.T) {
		h := &ScheduleHandler{}
		c := makeCtx(map[string]any{})
		h.UpdateLessonTimingsForDate(ctx, c)
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
			ls: &mockLessonService{
				updateTimingsFn: func(ctx context.Context, dto dtos.DateTimeChangeDTO) error { return nil },
			},
		}
		c := makeCtx(validBody)
		h.UpdateLessonTimingsForDate(ctx, c)
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
			ls: &mockLessonService{
				updateTimingsFn: func(ctx context.Context, dto dtos.DateTimeChangeDTO) error { return nil },
			},
		}
		c := makeCtx(validBody)
		h.UpdateLessonTimingsForDate(ctx, c)
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
				updateTimingsFn: func(ctx context.Context, dto dtos.DateTimeChangeDTO) error {
					return errors.New("validation failed")
				},
			},
		}
		c := makeCtx(validBody)
		h.UpdateLessonTimingsForDate(ctx, c)
		if c.Response.StatusCode() != 400 {
			t.Fatalf("status = %d, want 400", c.Response.StatusCode())
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
				updateTimingsFn: func(ctx context.Context, dto dtos.DateTimeChangeDTO) error { return nil },
			},
		}
		c := makeCtx(validBody)
		h.UpdateLessonTimingsForDate(ctx, c)
		if c.Response.StatusCode() != 200 {
			t.Fatalf("status = %d, want 200", c.Response.StatusCode())
		}
	})
}

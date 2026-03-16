package services

import (
	"brainz-api/internal/dtos"
	"brainz-api/internal/models"
	"context"
	"errors"
	"testing"
	"time"
)

type mockLessonRepository struct {
	deleteFn                func(ctx context.Context, id int64) error
	batchCreateFn           func(ctx context.Context, lessons []*models.Lesson) ([]int64, error)
	listForDayInstitutionFn func(ctx context.Context, date time.Time, institutionID uint64) ([]*models.Lesson, error)
}

func (m *mockLessonRepository) Create(ctx context.Context, lesson *models.Lesson) (int64, error) {
	return 0, nil
}

func (m *mockLessonRepository) GetByID(ctx context.Context, id int64) (*models.Lesson, error) {
	return nil, nil
}

func (m *mockLessonRepository) Update(ctx context.Context, lesson *models.Lesson) error {
	return nil
}

func (m *mockLessonRepository) Delete(ctx context.Context, id int64) error {
	return m.deleteFn(ctx, id)
}

func (m *mockLessonRepository) BatchCreate(ctx context.Context, lessons []*models.Lesson) ([]int64, error) {
	return m.batchCreateFn(ctx, lessons)
}

func (m *mockLessonRepository) ListForDayAndInstitution(ctx context.Context, date time.Time, institutionID uint64) ([]*models.Lesson, error) {
	return m.listForDayInstitutionFn(ctx, date, institutionID)
}

func TestLessonService_DeleteLesson(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var deletedID int64
		svc := &LessonService{
			lr: &mockLessonRepository{
				deleteFn: func(ctx context.Context, id int64) error {
					deletedID = id
					return nil
				},
				batchCreateFn: func(ctx context.Context, lessons []*models.Lesson) ([]int64, error) { return nil, nil },
				listForDayInstitutionFn: func(ctx context.Context, date time.Time, institutionID uint64) ([]*models.Lesson, error) {
					return nil, nil
				},
			},
		}

		if err := svc.DeleteLesson(context.Background(), 42); err != nil {
			t.Fatalf("DeleteLesson error = %v", err)
		}
		if deletedID != 42 {
			t.Fatalf("deleted id = %d, want 42", deletedID)
		}
	})

	t.Run("repository_error", func(t *testing.T) {
		wantErr := errors.New("delete failed")
		svc := &LessonService{
			lr: &mockLessonRepository{
				deleteFn:      func(ctx context.Context, id int64) error { return wantErr },
				batchCreateFn: func(ctx context.Context, lessons []*models.Lesson) ([]int64, error) { return nil, nil },
				listForDayInstitutionFn: func(ctx context.Context, date time.Time, institutionID uint64) ([]*models.Lesson, error) {
					return nil, nil
				},
			},
		}

		if err := svc.DeleteLesson(context.Background(), 42); !errors.Is(err, wantErr) {
			t.Fatalf("error = %v, want %v", err, wantErr)
		}
	})
}

func TestLessonService_AddLessons(t *testing.T) {
	start := time.Date(2026, 3, 14, 9, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	dto := dtos.LessonsCreateDTO{
		InstitutionID: 10,
		Lessons: []dtos.LessonCreateDTO{
			{
				Name:        "Math",
				CabNum:      "101",
				TeacherName: "Teacher",
				StartTime:   start,
				EndTime:     end,
				Num:         1,
				GroupID:     7,
			},
		},
	}

	t.Run("success", func(t *testing.T) {
		var created []*models.Lesson
		svc := &LessonService{
			gr: &mockGroupRepository{
				getListForInstitutionFn: func(ctx context.Context, instID int64) ([]models.Group, error) {
					return []models.Group{{ID: 7}}, nil
				},
			},
			lr: &mockLessonRepository{
				deleteFn: func(ctx context.Context, id int64) error { return nil },
				batchCreateFn: func(ctx context.Context, lessons []*models.Lesson) ([]int64, error) {
					created = lessons
					return []int64{1}, nil
				},
				listForDayInstitutionFn: func(ctx context.Context, date time.Time, institutionID uint64) ([]*models.Lesson, error) {
					return nil, nil
				},
			},
		}

		if err := svc.AddLessons(context.Background(), dto); err != nil {
			t.Fatalf("AddLessons error = %v", err)
		}
		if len(created) != 1 {
			t.Fatalf("created lessons = %d, want 1", len(created))
		}
		if created[0].InstitutionID != 10 || created[0].GroupID != 7 {
			t.Fatalf("unexpected created lesson: %+v", created[0])
		}
	})

	t.Run("group_repository_error", func(t *testing.T) {
		wantErr := errors.New("groups failed")
		svc := &LessonService{
			gr: &mockGroupRepository{
				getListForInstitutionFn: func(ctx context.Context, instID int64) ([]models.Group, error) {
					return nil, wantErr
				},
			},
			lr: &mockLessonRepository{
				deleteFn: func(ctx context.Context, id int64) error { return nil },
				batchCreateFn: func(ctx context.Context, lessons []*models.Lesson) ([]int64, error) {
					return nil, nil
				},
				listForDayInstitutionFn: func(ctx context.Context, date time.Time, institutionID uint64) ([]*models.Lesson, error) {
					return nil, nil
				},
			},
		}

		if err := svc.AddLessons(context.Background(), dto); !errors.Is(err, wantErr) {
			t.Fatalf("error = %v, want %v", err, wantErr)
		}
	})

	t.Run("group_not_belongs_to_institution", func(t *testing.T) {
		svc := &LessonService{
			gr: &mockGroupRepository{
				getListForInstitutionFn: func(ctx context.Context, instID int64) ([]models.Group, error) {
					return []models.Group{{ID: 100}}, nil
				},
			},
			lr: &mockLessonRepository{
				deleteFn: func(ctx context.Context, id int64) error { return nil },
				batchCreateFn: func(ctx context.Context, lessons []*models.Lesson) ([]int64, error) {
					return nil, nil
				},
				listForDayInstitutionFn: func(ctx context.Context, date time.Time, institutionID uint64) ([]*models.Lesson, error) {
					return nil, nil
				},
			},
		}

		if err := svc.AddLessons(context.Background(), dto); err == nil {
			t.Fatal("expected error when group is not allowed")
		}
	})

	t.Run("batch_create_error", func(t *testing.T) {
		wantErr := errors.New("batch create failed")
		svc := &LessonService{
			gr: &mockGroupRepository{
				getListForInstitutionFn: func(ctx context.Context, instID int64) ([]models.Group, error) {
					return []models.Group{{ID: 7}}, nil
				},
			},
			lr: &mockLessonRepository{
				deleteFn: func(ctx context.Context, id int64) error { return nil },
				batchCreateFn: func(ctx context.Context, lessons []*models.Lesson) ([]int64, error) {
					return nil, wantErr
				},
				listForDayInstitutionFn: func(ctx context.Context, date time.Time, institutionID uint64) ([]*models.Lesson, error) {
					return nil, nil
				},
			},
		}

		if err := svc.AddLessons(context.Background(), dto); !errors.Is(err, wantErr) {
			t.Fatalf("error = %v, want %v", err, wantErr)
		}
	})
}

func TestLessonService_GetForDateAndInstitution(t *testing.T) {
	start := time.Date(2026, 3, 14, 9, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	cab := "101"

	t.Run("success_handles_nil_cabnum", func(t *testing.T) {
		svc := &LessonService{
			lr: &mockLessonRepository{
				deleteFn:      func(ctx context.Context, id int64) error { return nil },
				batchCreateFn: func(ctx context.Context, lessons []*models.Lesson) ([]int64, error) { return nil, nil },
				listForDayInstitutionFn: func(ctx context.Context, date time.Time, institutionID uint64) ([]*models.Lesson, error) {
					return []*models.Lesson{
						{ID: 1, Name: "Math", CabNum: &cab, TeacherName: "Teach", StartTime: start, EndTime: end, Num: 1, GroupID: 7, InstitutionID: 10},
						{ID: 2, Name: "Physics", CabNum: nil, TeacherName: "Teach2", StartTime: start, EndTime: end, Num: 2, GroupID: 8, InstitutionID: 10},
					}, nil
				},
			},
		}

		got, err := svc.GetForDateAndInstitution(context.Background(), start, 10)
		if err != nil {
			t.Fatalf("GetForDateAndInstitution error = %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("len(lessons) = %d, want 2", len(got))
		}
		if got[0].CabNum != "101" {
			t.Fatalf("cab num = %q, want %q", got[0].CabNum, "101")
		}
		if got[1].CabNum != "" {
			t.Fatalf("cab num for nil = %q, want empty string", got[1].CabNum)
		}
	})

	t.Run("repository_error", func(t *testing.T) {
		wantErr := errors.New("list failed")
		svc := &LessonService{
			lr: &mockLessonRepository{
				deleteFn:      func(ctx context.Context, id int64) error { return nil },
				batchCreateFn: func(ctx context.Context, lessons []*models.Lesson) ([]int64, error) { return nil, nil },
				listForDayInstitutionFn: func(ctx context.Context, date time.Time, institutionID uint64) ([]*models.Lesson, error) {
					return nil, wantErr
				},
			},
		}

		_, err := svc.GetForDateAndInstitution(context.Background(), start, 10)
		if !errors.Is(err, wantErr) {
			t.Fatalf("error = %v, want %v", err, wantErr)
		}
	})
}

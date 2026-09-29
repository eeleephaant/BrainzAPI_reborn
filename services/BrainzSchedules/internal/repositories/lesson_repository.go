package repositories

import (
	"brainz-api/internal/models"
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LessonRepositoryInterface interface {
	Create(ctx context.Context, lesson *models.Lesson) (int64, error)
	GetByID(ctx context.Context, id int64) (*models.Lesson, error)
	Update(ctx context.Context, lesson *models.Lesson) error
	Delete(ctx context.Context, id int64) error
	BatchCreate(ctx context.Context, lessons []*models.Lesson) ([]int64, error)
	ListForDayAndInstitution(ctx context.Context, date time.Time, institutionID uint64) ([]*models.Lesson, error)
}

type LessonRepository struct {
	db *pgxpool.Pool
}

func NewLessonRepository(db *pgxpool.Pool) *LessonRepository {
	return &LessonRepository{db}
}

func (r *LessonRepository) ListForDayAndInstitution(ctx context.Context, date time.Time, institutionID uint64) ([]*models.Lesson, error) {
	startOfDay := time.Date(
		date.Year(), date.Month(), date.Day(),
		0, 0, 0, 0,
		date.Location(),
	)

	endOfDay := startOfDay.Add(24 * time.Hour)

	query, args, err := sq.
		Select("id", "name", "cab_num", "teacher_name", "start_time", "end_time", "num", "group_id", "institution_id", "created_at", "updated_at", "deleted_at").
		From("lessons").
		Where(sq.And{
			sq.GtOrEq{"start_time": startOfDay},
			sq.Lt{"start_time": endOfDay},
			sq.Eq{"institution_id": institutionID},
		}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lessons []*models.Lesson
	for rows.Next() {
		var lesson models.Lesson
		if err := rows.Scan(
			&lesson.ID,
			&lesson.Name,
			&lesson.CabNum,
			&lesson.TeacherName,
			&lesson.StartTime,
			&lesson.EndTime,
			&lesson.Num,
			&lesson.GroupID,
			&lesson.InstitutionID,
			&lesson.CreatedAt,
			&lesson.UpdatedAt,
			&lesson.DeletedAt,
		); err != nil {
			return nil, err
		}
		lessons = append(lessons, &lesson)
	}
	return lessons, nil
}

// Create inserts a new lesson and returns its ID.
func (r *LessonRepository) Create(ctx context.Context, lesson *models.Lesson) (int64, error) {
	query, args, err := sq.
		Insert("lessons").
		PlaceholderFormat(sq.Dollar).
		Columns(
			"name", "cab_num", "teacher_name", "start_time", "end_time", "num",
			"group_id", "institution_id",
		).
		Values(
			lesson.Name, lesson.CabNum, lesson.TeacherName, lesson.StartTime, lesson.EndTime, lesson.Num,
			lesson.GroupID, lesson.InstitutionID,
		).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return 0, err
	}
	var id int64
	err = r.db.QueryRow(ctx, query, args...).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *LessonRepository) BatchCreate(ctx context.Context, lessons []*models.Lesson) ([]int64, error) {
	if len(lessons) == 0 {
		return nil, nil
	}

	builder := sq.
		Insert("lessons").
		PlaceholderFormat(sq.Dollar).
		Columns(
			"name", "cab_num", "teacher_name",
			"start_time", "end_time", "num",
			"group_id", "institution_id",
		).
		Suffix("RETURNING id")

	for _, lesson := range lessons {
		builder = builder.Values(
			lesson.Name,
			lesson.CabNum,
			lesson.TeacherName,
			lesson.StartTime,
			lesson.EndTime,
			lesson.Num,
			lesson.GroupID,
			lesson.InstitutionID,
		)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]int64, 0, len(lessons))

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(ids) != len(lessons) {
		return nil, fmt.Errorf("expected %d ids, got %d", len(lessons), len(ids))
	}

	return ids, nil
}

// GetByID returns a lesson by its ID.
func (r *LessonRepository) GetByID(ctx context.Context, id int64) (*models.Lesson, error) {
	query, args, err := sq.
		Select(
			"id", "created_at", "updated_at", "deleted_at", "name", "cab_num",
			"teacher_name", "start_time", "end_time", "num", "group_id", "institution_id",
		).
		From("lessons").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}
	var lesson models.Lesson
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&lesson.ID,
		&lesson.CreatedAt,
		&lesson.UpdatedAt,
		&lesson.DeletedAt,
		&lesson.Name,
		&lesson.CabNum,
		&lesson.TeacherName,
		&lesson.StartTime,
		&lesson.EndTime,
		&lesson.Num,
		&lesson.GroupID,
		&lesson.InstitutionID,
	)
	if err != nil {
		return nil, err
	}
	return &lesson, nil
}

// Update updates lesson fields by ID.
func (r *LessonRepository) Update(ctx context.Context, lesson *models.Lesson) error {
	query, args, err := sq.
		Update("lessons").
		Set("name", lesson.Name).
		Set("cab_num", lesson.CabNum).
		Set("teacher_name", lesson.TeacherName).
		Set("start_time", lesson.StartTime).
		Set("end_time", lesson.EndTime).
		Set("num", lesson.Num).
		Set("group_id", lesson.GroupID).
		Set("institution_id", lesson.InstitutionID).
		Set("updated_at", lesson.UpdatedAt).
		Where(sq.Eq{"id": lesson.ID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}
	cmd, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("lesson not found")
	}
	return nil
}

// Delete removes a lesson by ID.
func (r *LessonRepository) Delete(ctx context.Context, id int64) error {
	query, args, err := sq.
		Delete("lessons").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}
	cmd, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("lesson not found")
	}
	return nil
}

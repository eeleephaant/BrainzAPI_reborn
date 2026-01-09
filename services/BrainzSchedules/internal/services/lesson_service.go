package services

import (
	"brainz-api/internal/models"
	"brainz-api/internal/repositories"
	"time"
)

type LessonService struct {
	repo repositories.LessonRepositoryInterface
}

func NewLessonService(repo repositories.LessonRepositoryInterface) *LessonService {
	return &LessonService{repo}
}

func (ls *LessonService) GetForDate(ctx context.Context, date time.Time) (*models.Lesson, error) {
	return ls.repo.GetByDate(ctx, date)
}

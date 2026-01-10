package services

import (
	"brainz-api/internal/dtos"
	"brainz-api/internal/repositories"
	"context"
	"time"
)

type LessonService struct {
	repo repositories.LessonRepositoryInterface
}

func NewLessonService(repo repositories.LessonRepositoryInterface) *LessonService {
	return &LessonService{repo}
}

func (ls *LessonService) UpdateTimingsForDateAndInstitution(ctx context.Context, date dtos.LessonsTimings, institutionID uint) error {
	return nil
}

func (ls *LessonService) GetForDateAndInstitution(ctx context.Context, date time.Time, institutionID uint64) ([]dtos.LessonDto, error) {
	lsModels, err := ls.repo.ListForDayAndInstitution(ctx, date, institutionID)
	if err != nil {
		return nil, err
	}

	var lsDtos []dtos.LessonDto
	for _, l := range lsModels {
		lsDtos = append(lsDtos, dtos.LessonDto{
			Id:            l.ID,
			Name:          l.Name,
			CabNum:        *l.CabNum,
			TeacherName:   l.TeacherName,
			StartTime:     l.StartTime,
			EndTime:       l.EndTime,
			Num:           l.Num,
			GroupID:       l.GroupID,
			InstitutionID: l.InstitutionID,
		})
	}
	return lsDtos, nil
}

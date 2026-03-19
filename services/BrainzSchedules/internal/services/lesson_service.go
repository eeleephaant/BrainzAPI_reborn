package services

import (
	"brainz-api/internal/dtos"
	"brainz-api/internal/models"
	"brainz-api/internal/repositories"
	"context"
	"fmt"
	"time"
)

type LessonService struct {
	lr repositories.LessonRepositoryInterface
	gr GroupRepository
}

func NewLessonService(lr repositories.LessonRepositoryInterface, gr *repositories.GroupRepository) *LessonService {
	return &LessonService{lr, gr}
}

func (ls *LessonService) UpdateTimingsForDateAndInstitution(ctx context.Context, dto dtos.DateTimeChangeDTO) error {
	if dto.InstitutionID == 0 {
		return fmt.Errorf("institution_id must be positive")
	}

	timingsByNum := map[uint8]dtos.LessonTiming{
		1: dto.LessonTimings.Lesson1,
		2: dto.LessonTimings.Lesson2,
		3: dto.LessonTimings.Lesson3,
		4: dto.LessonTimings.Lesson4,
		5: dto.LessonTimings.Lesson5,
		6: dto.LessonTimings.Lesson6,
	}

	for num, timing := range timingsByNum {
		if !timing.StartTime.Before(timing.EndTime) {
			return fmt.Errorf("lesson_%d start_time must be before end_time", num)
		}
	}

	lsModels, err := ls.lr.ListForDayAndInstitution(ctx, dto.Date, dto.InstitutionID)
	if err != nil {
		return err
	}

	for _, l := range lsModels {
		timing, ok := timingsByNum[l.Num]
		if !ok {
			continue
		}
		l.StartTime = mergeDateAndClock(dto.Date, timing.StartTime)
		l.EndTime = mergeDateAndClock(dto.Date, timing.EndTime)
		if err := ls.lr.Update(ctx, l); err != nil {
			return err
		}
	}

	return nil
}

func (ls *LessonService) DeleteLesson(ctx context.Context, lessonID int) error {
	err := ls.lr.Delete(ctx, int64(lessonID))
	if err != nil {
		return err
	}
	return nil
}

func (ls *LessonService) GetLessonInstitutionID(ctx context.Context, lessonID int64) (int64, error) {
	lesson, err := ls.lr.GetByID(ctx, lessonID)
	if err != nil {
		return 0, err
	}
	return int64(lesson.InstitutionID), nil
}

func (ls *LessonService) AddLessons(ctx context.Context, dto dtos.LessonsCreateDTO) error {
	if err := ValidateLessonsCreate(dto); err != nil {
		return err
	}

	instGroups, err := ls.gr.GetListForInstitution(ctx, int64(dto.InstitutionID))
	if err != nil {
		return err
	}

	allowed := make(map[int64]struct{}, len(instGroups))
	for _, g := range instGroups {
		allowed[g.ID] = struct{}{}
	}
	var lessonsModels []*models.Lesson

	for i, lesson := range dto.Lessons {
		if _, ok := allowed[int64(lesson.GroupID)]; !ok {
			return fmt.Errorf("group ID %d not found for institution (lesson index %d)",
				lesson.GroupID, i)
		}
		lessonsModels = append(lessonsModels, &models.Lesson{
			Name:          lesson.Name,
			CabNum:        &lesson.CabNum,
			TeacherName:   lesson.TeacherName,
			StartTime:     lesson.StartTime,
			EndTime:       lesson.EndTime,
			Num:           lesson.Num,
			GroupID:       lesson.GroupID,
			InstitutionID: uint(dto.InstitutionID),
		})
	}
	_, err = ls.lr.BatchCreate(ctx, lessonsModels)
	if err != nil {
		return err
	}
	return nil
}

func (ls *LessonService) GetForDateAndInstitution(ctx context.Context, date time.Time, institutionID uint64) ([]dtos.LessonDto, error) {
	lsModels, err := ls.lr.ListForDayAndInstitution(ctx, date, institutionID)
	if err != nil {
		return nil, err
	}

	var listDTOs []dtos.LessonDto
	for _, l := range lsModels {
		cabNum := ""
		if l.CabNum != nil {
			cabNum = *l.CabNum
		}
		listDTOs = append(listDTOs, dtos.LessonDto{
			Id:            l.ID,
			Name:          l.Name,
			CabNum:        cabNum,
			TeacherName:   l.TeacherName,
			StartTime:     l.StartTime,
			EndTime:       l.EndTime,
			Num:           l.Num,
			GroupID:       l.GroupID,
			InstitutionID: l.InstitutionID,
		})
	}
	return listDTOs, nil
}

func mergeDateAndClock(date time.Time, clock time.Time) time.Time {
	return time.Date(
		date.Year(), date.Month(), date.Day(),
		clock.Hour(), clock.Minute(), clock.Second(), clock.Nanosecond(),
		date.Location(),
	)
}

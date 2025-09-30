package mappers

import (
	"brainz-api/internal/models"
	"brainz-api/internal/models/dtos"
)

func LessonToDTO(lesson models.Lesson) dtos.LessonDto {
	return dtos.LessonDto{
		Id:            lesson.ID,
		Name:          lesson.Name,
		CabNum:        lesson.CabNum,
		TeacherName:   lesson.TeacherName,
		StartTime:     lesson.StartTime,
		EndTime:       lesson.EndTime,
		Num:           lesson.Num,
		GroupID:       lesson.GroupID,
		InstitutionID: lesson.InstitutionID,
		GroupName:     lesson.Group.Name,
	}
}

func LessonsToDTOs(lessons []*models.Lesson) []dtos.LessonDto {
	dtosSlice := make([]dtos.LessonDto, len(lessons))
	for i, lesson := range lessons {
		dtosSlice[i] = LessonToDTO(*lesson)
	}
	return dtosSlice
}

func GroupToDTO(group models.Group) dtos.GroupDto {
	return dtos.GroupDto{
		Id:   group.ID,
		Name: group.Name,
	}
}

func GroupsToDTOs(groups []*models.Group) []dtos.GroupDto {
	dtosSlice := make([]dtos.GroupDto, len(groups))
	for i, group := range groups {
		dtosSlice[i] = GroupToDTO(*group)
	}
	return dtosSlice
}

func InstitutionToDTO(institution models.Institution) dtos.InstitutionDto {
	return dtos.InstitutionDto{
		Id:   institution.ID,
		Name: institution.Name,
		Site: institution.Site,
	}
}

func InstitutionsToDTOs(institutions []*models.Institution) []dtos.InstitutionDto {
	dtosSlice := make([]dtos.InstitutionDto, len(institutions))
	for i, institution := range institutions {
		dtosSlice[i] = InstitutionToDTO(*institution)
	}
	return dtosSlice
}

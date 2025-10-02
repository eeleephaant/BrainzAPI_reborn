package mappers

import (
	"brainz-api/internal/models"
	"brainz-api/internal/models/dtos"
)

func LessonToDTO(lesson *models.Lesson) dtos.LessonDto {
	groupName := ""
	if lesson.Group.ID != 0 {
		groupName = lesson.Group.Name
	}

	return dtos.LessonDto{
		Id:            lesson.ID,
		Name:          lesson.Name,
		CabNum:        lesson.CabNum,
		TeacherName:   lesson.TeacherName,
		StartTime:     lesson.StartTime,
		EndTime:       lesson.EndTime,
		Num:           uint8(lesson.Num), // если Num стал uint8
		GroupID:       lesson.GroupID,
		InstitutionID: lesson.InstitutionID,
		GroupName:     groupName,
	}
}

func LessonsToDTOs(lessons []*models.Lesson) []dtos.LessonDto {
	dtosSlice := make([]dtos.LessonDto, 0, len(lessons))
	for _, l := range lessons {
		dtosSlice = append(dtosSlice, LessonToDTO(l))
	}
	return dtosSlice
}

func GroupToDTO(group *models.Group) dtos.GroupDto {
	return dtos.GroupDto{
		Id:   group.ID,
		Name: group.Name,
	}
}

func GroupsToDTOs(groups []*models.Group) []dtos.GroupDto {
	dtosSlice := make([]dtos.GroupDto, 0, len(groups))
	for _, g := range groups {
		dtosSlice = append(dtosSlice, GroupToDTO(g))
	}
	return dtosSlice
}

func InstitutionToDTO(inst *models.Institution) dtos.InstitutionDto {
	return dtos.InstitutionDto{
		Id:   inst.ID,
		Name: inst.Name,
		Site: inst.Site,
	}
}

func InstitutionsToDTOs(institutions []*models.Institution) []dtos.InstitutionDto {
	dtosSlice := make([]dtos.InstitutionDto, 0, len(institutions))
	for _, i := range institutions {
		dtosSlice = append(dtosSlice, InstitutionToDTO(i))
	}
	return dtosSlice
}

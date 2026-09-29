package dtos

type LessonsCreateDTO struct {
	InstitutionID int64             `json:"institutionID,required"`
	Lessons       []LessonCreateDTO `json:"lessons,required"`
}

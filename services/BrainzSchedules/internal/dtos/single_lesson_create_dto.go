package dtos

import "time"

type LessonCreateDTO struct {
	Name        string    `json:"name,required"          vd:"len($)>=1 && len($)<=100, msg:'Название урока должно быть от 1 до 100 символов'"`
	CabNum      string    `json:"cab_num,omitempty"      vd:"len($)<=20, msg:'Номер кабинета не должен превышать 20 символов'"` // optional в БД
	TeacherName string    `json:"teacher_name,required"  vd:"len($)>=2 && len($)<=100, msg:'Имя преподавателя от 2 до 100 символов'"`
	StartTime   time.Time `json:"start_time,required"    vd:""`
	EndTime     time.Time `json:"end_time,required"      vd:"gt($field.StartTime), msg:'Время окончания должно быть позже начала'"`
	Num         uint8     `json:"num,required"           vd:"$>=1 && $<=15, msg:'Номер урока должен быть от 1 до 15'"`
	GroupID     uint      `json:"group_id,required"      vd:"$>0, msg:'ID группы обязателен и должен быть положительным'"`
}

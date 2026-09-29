package dtos

import "time"

type LessonsTimings struct {
	Lesson1 LessonTiming `json:"lesson_1"`
	Lesson2 LessonTiming `json:"lesson_2"`
	Lesson3 LessonTiming `json:"lesson_3"`
	Lesson4 LessonTiming `json:"lesson_4"`
	Lesson5 LessonTiming `json:"lesson_5"`
	Lesson6 LessonTiming `json:"lesson_6"`
}

type LessonTiming struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

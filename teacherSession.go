package main

import "time"

type TeacherSession struct {
	Login    string
	Password string

	TeacherId uint

	IsCustomGroup bool
	GroupName     string
	GroupPageUrl  string

	DisciplineId   uint
	DisciplineName string
	Semester       uint

	LessonId   uint
	LessonDate time.Time
}

func NewTeacherSession(teacherWithActiveLesson *TeacherWithActiveLesson) *TeacherSession {
	return &TeacherSession{
		IsCustomGroup:  teacherWithActiveLesson.IsCustomGroup,
		Login:          teacherWithActiveLesson.Login,
		Password:       teacherWithActiveLesson.Password,
		TeacherId:      teacherWithActiveLesson.TeacherId,
		GroupName:      teacherWithActiveLesson.GroupName,
		DisciplineId:   teacherWithActiveLesson.DisciplineId,
		DisciplineName: teacherWithActiveLesson.DisciplineName,
		Semester:       teacherWithActiveLesson.Semester,
		LessonId:       teacherWithActiveLesson.LessonId,
		LessonDate:     teacherWithActiveLesson.LessonDate,
	}
}

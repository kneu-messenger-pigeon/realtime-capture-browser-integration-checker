package main

import "time"

type TeacherSession struct {
	Login    string
	Password string

	TeacherId uint

	GroupName    string
	GroupPageUrl string

	DisciplineId   uint
	DisciplineName string
	Semester       uint

	LessonId   uint
	LessonDate time.Time

	LogoutUrl string
}

func NewTeacherSession(teacherWithActiveLesson *TeacherWithActiveLesson) *TeacherSession {
	teacherSession.Login = teacherWithActiveLesson.Login
	teacherSession.Password = teacherWithActiveLesson.Password
	teacherSession.GroupName = teacherWithActiveLesson.GroupName
	teacherSession.DisciplineId = teacherWithActiveLesson.DisciplineId
	teacherSession.DisciplineName = teacherWithActiveLesson.DisciplineName
	teacherSession.Semester = teacherWithActiveLesson.Semester

	teacherSession.LessonId = teacherWithActiveLesson.LessonId
	teacherSession.LessonDate = teacherWithActiveLesson.LessonDate

	return &TeacherSession{
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

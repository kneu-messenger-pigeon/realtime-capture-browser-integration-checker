package main

import "time"

type TeacherSession struct {
	Login    string
	Password string

	GroupName    string
	GroupPageUrl string

	DisciplineId uint
	Semester     uint8

	LessonId   uint
	LessonDate time.Time

	LogoutUrl string
}

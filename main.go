package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

var config Config

var stubMatch = func(pat, str string) (bool, error) { return true, nil }

var chromeCtx context.Context

var dekanatRepository *DekanatRepository

var teacherSession = &TeacherSession{}

var dekanatReverseProxy *DekanatReverseProxy

func main() {
	var err error
	var cancel context.CancelFunc

	envFilename := ""
	if _, err = os.Stat(".env"); err == nil {
		envFilename = ".env"
	}

	config, err = loadConfig(envFilename)

	dekanatReverseProxy = NewReverseProxy(config.dekenatWebHost)

	// create context
	chromeCtx, cancel = createChromeContext(config.chromeWsUrl)
	defer cancel()

	dekanatRepository, err = NewDekanatRepository(config.dekanatDbDSN, config.dekanatSecret)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer dekanatRepository.Close()

	teacherWithActiveLesson := dekanatRepository.GetTeacherWithActiveLesson()
	if teacherWithActiveLesson == nil {
		log.Fatal("Teacher with active lesson not found")
	}

	fmt.Printf("Teacher with active lesson: %+v", teacherWithActiveLesson)

	teacherSession.Login = teacherWithActiveLesson.Login
	teacherSession.Password = teacherWithActiveLesson.Password
	teacherSession.GroupName = teacherWithActiveLesson.GroupName
	teacherSession.DisciplineId = teacherWithActiveLesson.DisciplineId
	teacherSession.Semester = teacherWithActiveLesson.Semester

	teacherSession.LessonId = teacherWithActiveLesson.LessonId
	teacherSession.LessonDate = teacherWithActiveLesson.LessonDate

	test := testing.InternalTest{
		Name: "integration testing",
		F: func(t *testing.T) {
			err = chromedp.Run(chromeCtx)
			assert.NoError(t, err)

			//	teacherSession.GroupPageUrl = `http://dekanat.kneu.edu.ua/cgi-bin/teachers.cgi?sesID=1A6F268B-BCAE-4111-AEA6-1A7337736B26&n=1&grp=%B2%C0-401&teacher=6653`
			reachGroupPage(t, teacherSession)

			fmt.Println("Start testing..")
			setupTests(t)
			fmt.Println("Test done")

			fmt.Print("Press enter to exit")
			_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
		},
	}

	testing.Main(stubMatch, []testing.InternalTest{test}, []testing.InternalBenchmark{}, []testing.InternalExample{})

}

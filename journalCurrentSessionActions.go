package main

import (
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
	"testing"
)

type LogoutFunc func()

func LoginAndFetchGroupPageUrl(t *testing.T, teacherSession *TeacherSession) (logoutFunc LogoutFunc) {
	var err error

	err = doLogin(t, teacherSession.Login, teacherSession.Password)
	makeScreenshot("login_finished")

	assert.NoError(t, err)

	logoutUrl := getLogoutUrl()
	if !assert.NotEmpty(t, logoutUrl, "Logout url is empty") {
		return
	}

	logoutFunc = func() {
		doLogout(logoutUrl)
	}

	teacherSession.GroupPageUrl = chooseGroup(teacherSession.GroupName, teacherSession.IsCustomGroup)
	fmt.Printf("Group page url: %s\n", teacherSession.GroupPageUrl)
	if !assert.NotEmpty(t, teacherSession.GroupPageUrl, "Group page url is empty") {
		return
	}

	makeScreenshot("group_page")

	return
}

func chooseDiscipline() (err error) {
	var currentLocation string

	_ = chromedp.Run(chromeCtx, chromedp.Location(&currentLocation))

	if currentLocation != teacherSession.GroupPageUrl {
		err = chromedp.Run(chromeCtx, chromedp.Navigate(teacherSession.GroupPageUrl))
		makeScreenshot("return_to_group_page")
		if err != nil {
			return err
		}
	}

	if teacherSession.IsCustomGroup {
		return chooseDisciplineInCustomGroup()
	} else {
		return chooseDisciplineInRegularGroup(teacherSession.DisciplineId, teacherSession.Semester)
	}
}

func verifyLessonOrScoreForm(t *testing.T) {
	if teacherSession.IsCustomGroup {
		verifyLessonOrScoreFormCustomGroup(t, teacherSession.GroupName)
	} else {
		verifyLessonOrScoreFormRegularGroup(t, teacherSession.GroupName, teacherSession.DisciplineName)
	}
}

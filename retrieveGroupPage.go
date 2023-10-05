package main

import (
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
	"testing"
)

func reachGroupPage(t *testing.T, teacherSession *TeacherSession) {
	var err error

	if teacherSession.GroupPageUrl == "" {
		err = doLogin(t, teacherSession.Login, teacherSession.Password)
		makeScreenshot("login_finished")

		assert.NoError(t, err)

		teacherSession.LogoutUrl = getLogoutUrl()
		assert.NotEmpty(t, teacherSession.LogoutUrl, "Logout url is empty")

		teacherSession.GroupPageUrl = chooseGroup(teacherSession.GroupName)
		fmt.Printf("Group page url: %s\n", teacherSession.GroupPageUrl)
		if !assert.NotEmpty(t, teacherSession.GroupPageUrl, "Group page url is empty") {
			return
		}
		makeScreenshot("group_page")

	} else {
		err = chromedp.Run(chromeCtx, chromedp.Navigate(teacherSession.GroupPageUrl))
		if !assert.NoError(t, err, "Failed to navigate to group page") {
			return
		}
		makeScreenshot("return_to_group_page")
	}
	/* */

	err = chooseDiscipline(teacherSession.DisciplineId, teacherSession.Semester)
	makeScreenshot("discipline_page")
	if !assert.NoError(t, err, "Failed to choose discipline") {
		return
	}

}

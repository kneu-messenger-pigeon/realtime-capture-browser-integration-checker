package main

import (
	"fmt"
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

	teacherSession.GroupPageUrl = chooseGroup(teacherSession.GroupName)
	fmt.Printf("Group page url: %s\n", teacherSession.GroupPageUrl)
	if !assert.NotEmpty(t, teacherSession.GroupPageUrl, "Group page url is empty") {
		return
	}

	makeScreenshot("group_page")

	return
}

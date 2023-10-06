package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func LoginAndFetchGroupPageUrl(t *testing.T, teacherSession *TeacherSession) {
	var err error

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
}

package main

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

var loginError = fmt.Errorf("failed to Login user")

const UserAlreadyLoggedInErrorText = "Користувач з таким ім'ям вже працює!"

func doLogin(t *testing.T, login string, password string) error {
	loginCtx, loginContextCancel := context.WithTimeout(chromeCtx, time.Second*60)
	defer loginContextCancel()

	var errorText string
	var userName string

	loginInput := `input[name="user_name"]`
	passwordInput := `input[name="user_pwd"]`

	formErrorElement := `form .bg-danger`

	err := chromedp.Run(chromeCtx, chromedp.Navigate(dekanatReverseProxy.ProxyOrigin+`/cgi-bin/kaf.cgi?n=999&t=98`))
	assert.NoError(t, err)
	if err != nil {
		return err
	}

	submitFormAndCheckError := func() {
		errorText = ""
		userName = ""

		err = chromedp.Run(loginCtx, chromedp.Tasks{
			chromedp.WaitVisible(loginInput),
			chromedp.SendKeys(loginInput, login),
			chromedp.SendKeys(passwordInput, password),
			chromedp.Submit(passwordInput),
			chromedp.WaitReady(`//body`),
		})

		assert.NoError(t, err)

		if err != nil {
			return
		}

		submitResultCtx, submitResultCtxCancel := context.WithCancel(loginCtx)
		go func() {
			_ = chromedp.Run(submitResultCtx, chromedp.Text(formErrorElement, &errorText))
			submitResultCtxCancel()
		}()

		_ = chromedp.Run(submitResultCtx, chromedp.Text(`.navbar-header a`, &userName))
		submitResultCtxCancel()
	}
	submitFormAndCheckError()

	if strings.Contains(errorText, UserAlreadyLoggedInErrorText) {
		submitFormAndCheckError()
	}

	isUserLoggedIn := strings.TrimSpace(userName) != "" && strings.TrimSpace(errorText) == ""

	log.Printf("errorText: `%s`", strings.TrimSpace(errorText))
	log.Printf("user name: `%s`", strings.TrimSpace(userName))

	if isUserLoggedIn {
		return nil
	}

	if loginCtx.Err() != nil {
		return loginCtx.Err()
	}

	return fmt.Errorf("%w: `%s`; error: %s", loginError, login, errorText)
}

func getLogoutUrl() string {
	var currentLocation string
	_ = chromedp.Run(chromeCtx, chromedp.Location(&currentLocation))

	currentPageUrl, _ := url.Parse(currentLocation)

	if currentLocation == "" {
		return ""
	}

	var rawRelativeLogoutUrl string
	var ok bool

	ctx, cancel := context.WithTimeout(chromeCtx, time.Second)
	_ = chromedp.Run(ctx, chromedp.AttributeValue(`//a[contains(., "Вихід")]`, "href", &rawRelativeLogoutUrl, &ok))
	cancel()

	if rawRelativeLogoutUrl == "" {
		rawRelativeLogoutUrl = "/cgi-bin/kaf.cgi?n=100&sesID=" + currentPageUrl.Query().Get("sesID")
	}

	logoutUrl, _ := url.Parse(rawRelativeLogoutUrl)
	return currentPageUrl.ResolveReference(logoutUrl).String()
}

func doLogout(logoutUrl string) {
	go func() {
		_, err := http.Get(logoutUrl)
		if err != nil {
			log.Printf("failed to make logout http-request: %s", err)
		}
	}()

	ctx, cancel := context.WithTimeout(chromeCtx, time.Second*20)
	defer cancel()
	err := chromedp.Run(ctx, chromedp.Navigate(logoutUrl))
	cancel()
	if err != nil {
		log.Printf("failed to navigate to logout: %s", err)
	}

	makeScreenshot("logout_finished")
}

package main

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestReverseProxy(t *testing.T) {
	if config.skipReverseProxyTest {
		t.Skip("Skipping reverse proxy test")
		return
	}

	fmt.Println("TestReverseProxy")
	defer printTestResult(t, "TestReverseProxy")

	doPOST := func(url string, formPost string) int {
		res, _ := http.Post(
			dekanatReverseProxy.ProxyOrigin+url, "application/x-www-form-urlencoded",
			bytes.NewBuffer([]byte(formPost)),
		)
		return res.StatusCode
	}

	doGET := func(url string, query string) int {
		res, _ := http.Get(dekanatReverseProxy.ProxyOrigin + url + "?" + query)
		return res.StatusCode
	}

	// allow login page
	assert.Equal(t, http.StatusOK, doGET("/cgi-bin/kaf.cgi", "n=1&ts=8937"))

	dekanatReverseProxy.SwitchOffline()
	assert.Equal(t, http.StatusForbidden, doGET("/cgi-bin/kaf.cgi", "n=1&ts=8937"))

	dekanatReverseProxy.SwitchOnline()

	// check login post
	assert.Equal(t, http.StatusOK, doPOST(
		"/cgi-bin/kaf.cgi?n=1&ts=10663",
		"user_name=123&user_pwd=123&n=1&rout=&t=1",
	))

	// check disciplines result post
	assert.Equal(t, http.StatusOK, doPOST(
		"/cgi-bin/teachers.cgi?sesID=00BA2572-0000-4C00-9E4B-4FFDF794CF76",
		"grp=%B2%C0-201&n=7&sesID=FABA2572-9B15-4C7D-9E4B-4FFDF794CF76&teacher=158&irc=0&tid=0&CYKLE=-1&prt=000000&hlf=0&d1=&d2=&m=-1",
	))

	// check delete request
	assert.Equal(t, http.StatusForbidden, doGET(
		"/cgi-bin/teachers.cgi",
		"sesID=00BA2572-0000-4C00-9E4B-4FFDF794CF76&n=11&action=delete&tid=0&CYKLE=-1&prt=202090&hlf=0&d1=&d2=&m=-1",
	))

}

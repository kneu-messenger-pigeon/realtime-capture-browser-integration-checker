package main

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

func chooseGroup(groupName string) (groupPageUrl string) {
	ctx, cancel := context.WithTimeout(chromeCtx, time.Second*60)
	defer cancel()

	groupName = strings.ReplaceAll(groupName, `""`, `\"`)

	err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Click(`//div[contains(@class, "jumbotron")]//a[contains(., "Академічні групи")]`),
		chromedp.WaitVisible(`//h2`),
	})

	if err != nil {
		return ""
	}

	fetchLinkCtx, cancelFetchLinkCtx := context.WithTimeout(ctx, time.Second*2)
	err = chromedp.Run(fetchLinkCtx, chromedp.Click(fmt.Sprintf(`//div[contains(@class, "jumbotron")]//a[text() = "%s"]`, groupName)))
	cancelFetchLinkCtx()

	if err != nil {
		fmt.Printf("Group %s not found\n", groupName)
		makeScreenshot("group_not_found")
		return ""
	}

	_ = chromedp.Run(ctx, chromedp.Tasks{
		chromedp.WaitVisible(`//h2`),
		chromedp.Location(&groupPageUrl),
	})

	return
}

func chooseDiscipline(disciplineId uint, semester uint) (err error) {
	var currentLocation string

	_ = chromedp.Run(chromeCtx, chromedp.Location(&currentLocation))

	if currentLocation != teacherSession.GroupPageUrl {
		err = chromedp.Run(chromeCtx, chromedp.Navigate(teacherSession.GroupPageUrl))
		makeScreenshot("return_to_group_page")
		if err != nil {
			return err
		}
	}

	var currentDisciplineId string

	var semesterLabel string
	if semester == 1 {
		semesterLabel = "перше"
	} else {
		semesterLabel = "друге"
	}

	form := findVisibleForm(`.jumbotron form[method="post"]`)
	fromForm := chromedp.FromNode(form)
	formXPath := form.FullXPathByID()

	ctx, cancel := context.WithTimeout(chromeCtx, time.Second*3)
	defer cancel()

	semesterRadioSelector := fmt.Sprintf(`//label[text() = "%s"]//input`, semesterLabel)
	err = chromedp.Run(ctx, chromedp.Tasks{
		// get current selected discipline. Its value is stored in hidden input for single discipline or in select for multiple disciplines
		chromedp.Value(formXPath+`//*[@name="prt"]`, &currentDisciplineId),
		chromedp.SetAttributeValue(formXPath+`//option[text() = "За весь період"]`, "selected", "selected"),
		chromedp.Click(formXPath + semesterRadioSelector),
		//	chromedp.SetAttributeValue(semesterRadioSelector, "checked", "checked", fromForm),
	})

	if err != nil {
		return err
	}

	fmt.Printf("Current discipline id: %s; target discipline id %d; \n", currentDisciplineId, disciplineId)

	if currentDisciplineId != fmt.Sprintf("%d", disciplineId) {
		disciplineOption := fmt.Sprintf(`//option[@value = "%d"]`, disciplineId)
		err = chromedp.Run(ctx, chromedp.SetAttributeValue(formXPath+disciplineOption, "selected", "selected"))

		if err != nil {
			return err
		}
	}

	ctx, cancel = context.WithTimeout(chromeCtx, time.Second*25)
	defer cancel()
	err = chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Submit(`[name="grade"]`, fromForm),
		chromedp.WaitReady(`//body`),
	})

	makeScreenshot("discipline_page")

	return err
}

func findVisibleForm(selector string) *cdp.Node {
	var formNodes []*cdp.Node

	err := chromedp.Run(chromeCtx, chromedp.Nodes(selector, &formNodes))

	if err != nil {
		return nil
	}

	executor := chromedp.FromContext(chromeCtx).Target
	isVisible := func(node *cdp.Node) bool {
		boxModel, visibleErr := dom.GetBoxModel().WithNodeID(node.NodeID).Do(cdp.WithExecutor(chromeCtx, executor))
		return visibleErr == nil && boxModel != nil
	}

	for _, formNode := range formNodes {
		if isVisible(formNode) {
			return formNode
		}
	}

	return nil
}

func expectBlockedPage(t *testing.T) {
	ctx, cancel := context.WithTimeout(chromeCtx, time.Second*1)
	defer cancel()

	err := chromedp.Run(ctx, chromedp.WaitVisible(`#__blocked_page`, chromedp.ByQuery))
	if !assert.NoError(t, err, "Unexpected page, must be blocked page") {
		makeScreenshot("must_be_blocked_page")
		t.FailNow()
	}
}

// //table[@id ="mMarks"]//th[contains(., "03.10.2023")]

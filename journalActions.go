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

func chooseGroup(groupName string, isCustomGroup bool) (groupPageUrl string) {
	ctx, cancel := context.WithTimeout(chromeCtx, time.Second*60)
	defer cancel()

	groupName = strings.ReplaceAll(groupName, `""`, `\"`)

	var groupListLabelText string
	if isCustomGroup {
		groupListLabelText = "Збірні групи"
	} else {
		groupListLabelText = "Академічні групи"
	}
	groupListSelector := fmt.Sprintf(
		`//div[contains(@class, "jumbotron")]//a[contains(text(), "%s")]`,
		groupListLabelText,
	)

	err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Click(groupListSelector),
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

func chooseDisciplineInCustomGroup() (err error) {
	formXPath := findVisibleForm(`.jumbotron form[method="post"]`).FullXPathByID()

	ctx, cancel := context.WithTimeout(chromeCtx, time.Second*25)
	defer cancel()
	err = chromedp.Run(ctx, chromedp.Tasks{
		chromedp.SetAttributeValue(
			formXPath+`//option[text() = "За весь період"]`,
			"selected", "selected",
		),
		chromedp.Submit(formXPath + `//*[@name="grade"]`),
		chromedp.WaitReady(`//body`),
	})

	makeScreenshot("discipline_page")

	return err
}

func chooseDisciplineInRegularGroup(disciplineId uint, semester uint) (err error) {
	var currentDisciplineId string

	var semesterLabel string
	if semester == 1 {
		semesterLabel = "перше"
	} else {
		semesterLabel = "друге"
	}

	form := findVisibleForm(`.jumbotron form[method="post"]`)
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
		chromedp.Submit(formXPath + `//*[@name="grade"]`),
		chromedp.WaitReady(`//body`),
	})

	makeScreenshot("discipline_page")

	return err
}

func openLessonPopup(lessonDate time.Time) (err error) {
	lessonSelector := fmt.Sprintf(
		`//table[@id="mMarks"]//th[contains(., "%s")][last()]//a[contains(text(), "%s")]`,
		lessonDate.Format("02.01.2006"), lessonDate.Format("02.01.2006"),
	)

	ctx, cancel := context.WithTimeout(chromeCtx, time.Second*5)
	defer cancel()

	err = chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Click(lessonSelector),
		chromedp.WaitVisible(`//*[contains(@class, "modal-title")][contains(text(), "Дії для заняття")]`),
		chromedp.Sleep(time.Millisecond * 400),
	})

	return err
}

func verifyLessonOrScoreFormRegularGroup(t *testing.T, expectedGroupName string, expectedDisciplineName string) {
	var currentGroup string
	var currentDiscipline string

	ctx, cancel := context.WithTimeout(chromeCtx, time.Millisecond*300)
	defer cancel()

	err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Text(`//*[contains(text(), "Академічна група")]`, &currentGroup),
		chromedp.Text(`//*[contains(text(), "Дисципліна")]`, &currentDiscipline),
	})

	if !assert.NoError(t, err, "Failed to get current group and discipline") {
		return
	}

	assert.Contains(t, currentGroup, expectedGroupName, "Wrong group name")
	assert.Contains(t, currentDiscipline, expectedDisciplineName, "Wrong discipline name")
}

var replacers = [5]*strings.Replacer{
	strings.NewReplacer(`+`, ``),
	strings.NewReplacer(`"`, ``),
	strings.NewReplacer(`<`, ``),
	strings.NewReplacer(`>`, ``),
	strings.NewReplacer(`&`, ``),
}

func verifyLessonOrScoreFormCustomGroup(t *testing.T, expectedGroupName string) {
	var currentGroup string

	ctx, cancel := context.WithTimeout(chromeCtx, time.Millisecond*300)
	defer cancel()

	err := chromedp.Run(ctx, chromedp.Text(`//*[contains(text(), "Збірна група")]`, &currentGroup))

	if !assert.NoError(t, err, "Failed to get current group and discipline") {
		return
	}

	clearExpectedGroupName := expectedGroupName
	for _, replacer := range replacers {
		clearExpectedGroupName = replacer.Replace(clearExpectedGroupName)
	}

	assert.Contains(t, currentGroup, clearExpectedGroupName, "Wrong group name")
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

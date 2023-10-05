package main

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"strings"
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

func chooseDiscipline(disciplineId uint, semester uint8) error {
	var currentDisciplineId string

	var semesterLabel string
	if semester == 1 {
		semesterLabel = "перше"
	} else {
		semesterLabel = "друге"
	}

	form := findVisibleForm(`.jumbotron form[method="post"]`)
	fromForm := chromedp.FromNode(form)

	ctx, cancel := context.WithTimeout(chromeCtx, time.Second*3000)
	defer cancel()

	semesterRadioSelector := fmt.Sprintf(`//label[text() = "%s"]//input`, semesterLabel)

	err := chromedp.Run(ctx, chromedp.Tasks{
		// get current selected discipline. It's value is stored in hidden input for single discipline or in select for multiple disciplines
		chromedp.Value(`[name="prt"]`, &currentDisciplineId, fromForm),
		chromedp.SetAttributeValue(`//option[text() = "За весь період"]`, "selected", "selected", fromForm),
		chromedp.SetAttributeValue(semesterRadioSelector, "checked", "checked", fromForm),
	})

	if err != nil {
		return err
	}

	fmt.Printf("Current discipline id: %s; target discipline id %d; \n", currentDisciplineId, disciplineId)

	if currentDisciplineId != fmt.Sprintf("%d", disciplineId) {
		disciplineOption := fmt.Sprintf(`option[value="%d"]`, disciplineId)
		err = chromedp.Run(ctx, chromedp.SetAttributeValue(disciplineOption, "selected", "selected", fromForm))

		if err != nil {
			return err
		}
	} else {
		fmt.Println("Discipline is already selected")
	}

	ctx, cancel = context.WithTimeout(chromeCtx, time.Second*20)
	defer cancel()
	err = chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Submit(`[name="grade"]`, fromForm),
		chromedp.WaitReady(`//body`),
	})

	return err
}

func findVisibleForm(selector string) *cdp.Node {
	var visibleForm *cdp.Node
	var formNodes []*cdp.Node

	visibleCxt, cancelVisible := context.WithTimeout(chromeCtx, time.Second*3)
	defer cancelVisible()
	err := chromedp.Run(chromeCtx, chromedp.Nodes(selector, &formNodes))

	if err != nil {
		return nil
	}
	for _, formNode := range formNodes {
		go func(node *cdp.Node) {
			formVisibleErr := chromedp.Run(visibleCxt, chromedp.WaitVisible(`input,button`, chromedp.FromNode(node)))
			if formVisibleErr != nil {
				cancelVisible()
				visibleForm = node
			}
		}(formNode)
	}

	return visibleForm
}

// //table[@id ="mMarks"]//th[contains(., "03.10.2023")]

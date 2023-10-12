package main

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	dekanatEvents "github.com/kneu-messenger-pigeon/dekanat-events"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test4DeleteLesson(t *testing.T) {
	fmt.Println("Test4DeleteLesson")
	defer printTestResult(t, "Test4DeleteLesson")

	err := chooseDiscipline(teacherSession.DisciplineId, teacherSession.Semester)
	if !assert.NoError(t, err, "Failed to choose discipline") {
		return
	}

	err = openLessonPopup(teacherSession.LessonDate)
	makeScreenshot("lesson_popup")
	if !assert.NoError(t, err, "Failed to wait for lesson popup") {
		return
	}

	dekanatReverseProxy.ClearBlockedRequests()
	dekanatReverseProxy.SwitchOffline()
	defer dekanatReverseProxy.SwitchOnline()

	deleteLessonSelector := `//*[contains(@class, "modal-content")]//a[contains(text(), "Видалити")][contains(text(), "заняття")]`

	ctx, cancel := context.WithTimeout(chromeCtx, time.Second*10)
	defer cancel()

	listenCtx, listenCancel := context.WithCancel(ctx)
	defer listenCancel()

	chromedp.ListenTarget(listenCtx, func(ev interface{}) {
		if _, ok := ev.(*page.EventJavascriptDialogOpening); ok {
			go func() {
				listenCancel()
				errDialog := chromedp.Run(ctx, page.HandleJavaScriptDialog(true))
				assert.NoError(t, errDialog, "Failed to click on 'Видалити заняття' confirm modal")
			}()
		}
	})

	err = chromedp.Run(ctx, chromedp.Click(deleteLessonSelector))
	assert.NoError(t, err, "Failed to click on 'Видалити заняття' button")

	// assert
	expectBlockedPage(t)
	assert.Equal(t, 1, len(dekanatReverseProxy.BlockedRequests), "Wrong number of blocked requests")

	event := realtimeQueue.Fetch(time.Second * 15)

	assert.NotNil(t, event, "Event not found")
	assert.IsType(t, dekanatEvents.LessonDeletedEvent{}, event, "Wrong event type")

	lessonDeletedEvent, ok := event.(dekanatEvents.LessonDeletedEvent)
	if !ok {
		return
	}

	assert.Equal(t, teacherSession.DisciplineId, lessonDeletedEvent.GetDisciplineId(), "Wrong group id")
	assert.Equal(t, teacherSession.Semester, lessonDeletedEvent.GetSemester(), "Wrong semester")
	assert.Equal(t, teacherSession.LessonId, lessonDeletedEvent.GetLessonId(), "Wrong semester")

}

package main

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	dekanatEvents "github.com/kneu-messenger-pigeon/dekanat-events"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test2EditLesson(t *testing.T) {
	fmt.Println("Test2EditLesson")
	defer printTestResult(t, "Test2EditLesson")

	err := chooseDiscipline(teacherSession.DisciplineId, teacherSession.Semester)
	if !assert.NoError(t, err, "Failed to choose discipline") {
		return
	}

	err = openLessonPopup(teacherSession.LessonDate)
	makeScreenshot("lesson_popup")
	if !assert.NoError(t, err, "Failed to wait for lesson popup") {
		return
	}

	editLessonSelector := `//*[contains(@class, "modal-content")]//a[contains(text(), "Змінити загальні дані")]`

	ctx, cancel := context.WithTimeout(chromeCtx, time.Second*10)
	defer cancel()

	err = chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Click(editLessonSelector),
		chromedp.WaitVisible(`//body`),
	})

	verifyLessonForm(t, teacherSession.GroupName, teacherSession.DisciplineName)
	makeScreenshot("edit_lesson_form")

	if t.Failed() {
		return
	}

	dekanatReverseProxy.ClearBlockedRequests()
	dekanatReverseProxy.SwitchOffline()
	defer dekanatReverseProxy.SwitchOnline()

	err = chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Click(`//button[contains(text(), "Зберегти")][1]`),
		chromedp.WaitVisible(`//body`),
	})
	assert.NoError(t, err, "Failed to click on 'Зберегти' button")

	// assert
	expectBlockedPage(t)
	assert.Equal(t, 1, len(dekanatReverseProxy.BlockedRequests), "Wrong number of blocked requests")

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	event := realtimeQueue.Fetch(ctx)
	cancel()

	assert.NotNil(t, event, "Event not found")
	assert.IsType(t, dekanatEvents.LessonEditEvent{}, event, "Wrong event type")

	lessonEditEvent, ok := event.(dekanatEvents.LessonEditEvent)
	if !ok {
		return
	}

	assert.Equal(t, teacherSession.DisciplineId, lessonEditEvent.GetDisciplineId(), "Wrong group id")
	assert.Equal(t, teacherSession.Semester, lessonEditEvent.GetSemester(), "Wrong semester")
	assert.Equal(t, teacherSession.LessonId, lessonEditEvent.GetLessonId(), "Wrong semester")

	expectedLessonDate := teacherSession.LessonDate.Format("02.01.2006")
	assert.Equal(t, expectedLessonDate, lessonEditEvent.Date, "Wrong date")
}

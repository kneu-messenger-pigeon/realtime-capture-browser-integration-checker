package main

import "testing"

func setupTests(t *testing.T) {
	t.Run("Test1CreateLesson", Test1CreateLesson)
	t.Run("Test2EditLesson", Test2EditLesson)
}

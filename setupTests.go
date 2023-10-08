package main

import "testing"

func setupTests(t *testing.T) {
	t.Run("Test1CreateLesson", Test1CreateLesson)
	t.Run("Test2EditLesson", Test2EditLesson)
	t.Run("TestTest3EditScore", TestTest3EditScore)
	t.Run("Test4DeleteLesson", Test4DeleteLesson)
}

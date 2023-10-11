package main

import (
	"bytes"
	"testing"
)

type CaptureScriptUrlReplacer struct {
	prodScriptUrl     []byte
	testScriptUrl     []byte
	replacedPageCount int
}

func NewCaptureScriptUrlReplacer(prodScriptUrl, testScriptUrl string) *CaptureScriptUrlReplacer {
	return &CaptureScriptUrlReplacer{
		prodScriptUrl: []byte(prodScriptUrl),
		testScriptUrl: []byte(testScriptUrl),
	}
}

func (r *CaptureScriptUrlReplacer) Replace(body []byte) []byte {
	if !bytes.Contains(body, r.prodScriptUrl) {
		return body
	}

	r.replacedPageCount++

	return bytes.ReplaceAll(body, r.prodScriptUrl, r.testScriptUrl)
}

func (r *CaptureScriptUrlReplacer) GetReplacedPageCount() int {
	return r.replacedPageCount
}

func (r *CaptureScriptUrlReplacer) AssertReplaced(t *testing.T) bool {
	if r.replacedPageCount == 0 {
		t.Error("Capture script url not replaced")
		return false
	}
	return true
}

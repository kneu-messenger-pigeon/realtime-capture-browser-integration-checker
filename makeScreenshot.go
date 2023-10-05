package main

import (
	"github.com/chromedp/chromedp"
	"log"
	"os"
)

const screenshotDir = "screenshots"

func makeScreenshot(name string) {
	if _, err := os.Stat(screenshotDir); err != nil {
		err = os.MkdirAll(screenshotDir, 0o755)
		if err != nil {
			log.Fatal(err)
		}
	}

	var imgBuf []byte
	// capture entire browser viewport, returning png with quality=90
	err := chromedp.Run(chromeCtx, chromedp.FullScreenshot(&imgBuf, 90))
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(screenshotDir+"/"+name+".png", imgBuf, 0o644)
	if err != nil {
		log.Fatal(err)
	}
}

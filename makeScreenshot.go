package main

import (
	"github.com/chromedp/chromedp"
	"log"
	"os"
	"time"
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

	t := time.Now().Format("2006-01-02-15-04-05_")

	err = os.WriteFile(screenshotDir+"/"+t+name+".png", imgBuf, 0o644)
	if err != nil {
		log.Fatal(err)
	}
}

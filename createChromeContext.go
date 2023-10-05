package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/chromedp/chromedp"
	"log"
	"os"
)

func createChromeContext(chromeWsUrl string) (context.Context, context.CancelFunc) {
	var allocCtx context.Context
	var allocCtxCancel context.CancelFunc

	if chromeWsUrl == "EXEC" || chromeWsUrl == "DESKTOP" {
		allocCtx, allocCtxCancel = createDesktopChromeAllocator(chromeWsUrl != "DESKTOP")
	} else {
		allocCtx, allocCtxCancel = createRemoteChromeAllocator(chromeWsUrl)

	}

	logFile, err := os.Create("chrome.log")
	if err != nil {
		log.Fatal(err)
	}
	logPrint := func(format string, v ...any) {
		_, _ = fmt.Fprintf(logFile, format+"\n", v...)
	}

	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(logPrint), chromedp.WithDebugf(logPrint))

	return taskCtx, func() {
		cancel()
		allocCtxCancel()

		_ = logFile.Close()
	}
}

func createRemoteChromeAllocator(chromeWsUrl string) (context.Context, context.CancelFunc) {
	devtoolsWsURL := flag.String("devtools-ws-url", chromeWsUrl, "DevTools WebSocket URL")
	flag.Parse()

	return chromedp.NewRemoteAllocator(context.Background(), *devtoolsWsURL)
}

func createDesktopChromeAllocator(headless bool) (context.Context, context.CancelFunc) {
	return chromedp.NewExecAllocator(
		context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("headless", headless))...,
	)
}

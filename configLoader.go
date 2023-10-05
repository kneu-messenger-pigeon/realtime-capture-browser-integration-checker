package main

import (
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	chromeWsUrl    string
	dekanatDbDSN   string
	dekanatSecret  string
	sqsQueueUrl    string
	dekenatWebHost string
}

func loadConfig(envFilename string) (Config, error) {
	if envFilename != "" {
		err := godotenv.Load(envFilename)
		if err != nil {
			return Config{}, errors.New(fmt.Sprintf("Error loading %s file: %s", envFilename, err))
		}
	}
	loadedConfig := Config{
		chromeWsUrl:    os.Getenv("DEVTOOLS_WS_URL"),
		dekanatDbDSN:   os.Getenv("DEKANAT_DB_DSN"),
		dekanatSecret:  os.Getenv("DEKANAT_SECRET"),
		sqsQueueUrl:    os.Getenv("AWS_SQS_QUEUE_URL"),
		dekenatWebHost: os.Getenv("DEKANAT_WEB_HOST"),
	}

	if loadedConfig.chromeWsUrl == "" {
		return Config{}, errors.New("empty DEVTOOLS_WS_URL")
	}

	if loadedConfig.dekanatDbDSN == "" {
		return Config{}, errors.New("empty DEKANAT_DB_DSN")
	}
	if loadedConfig.dekanatSecret == "" {
		return Config{}, errors.New("empty DEKANAT_SECRET")
	}

	if loadedConfig.sqsQueueUrl == "" {
		return Config{}, errors.New("empty AWS_SQS_QUEUE_URL")
	}

	if loadedConfig.dekenatWebHost == "" {
		loadedConfig.dekenatWebHost = "http://dekanat.kneu.edu.ua"
	}

	return loadedConfig, nil
}

package main

import (
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"time"
)

type Config struct {
	kafkaHost                          string
	kafkaTimeout                       time.Duration
	kafkaAttempts                      int
	primaryDekanatDbDSN                string
	secondaryDekanatDbDSN              string
	secondaryDbCheckInterval           time.Duration
	sqsQueueUrl                        string
	telegramToken                      string
	repeatScoreChangesTimeframeSeconds time.Duration
	appStartDelay                      time.Duration
	kneuBaseUrl                        string
	kneuClientId                       int
	kneuClientSecret                   string
	publicUrl                          string
	skipWait                           bool
	debugUpdates                       bool
}

func loadConfig(envFilename string) (Config, error) {
	if envFilename != "" {
		err := godotenv.Load(envFilename)
		if err != nil {
			return Config{}, errors.New(fmt.Sprintf("Error loading %s file: %s", envFilename, err))
		}
	}

	kafkaTimeout, err := strconv.Atoi(os.Getenv("KAFKA_TIMEOUT"))
	if kafkaTimeout == 0 || err != nil {
		kafkaTimeout = 10
	}

	kafkaAttempts, err := strconv.Atoi(os.Getenv("KAFKA_ATTEMPTS"))
	if kafkaAttempts == 0 || err != nil {
		kafkaAttempts = 0
	}

	kneuClientId, err := strconv.Atoi(os.Getenv("KNEU_CLIENT_ID"))
	if err != nil || kneuClientId < 1 {
		return Config{}, errors.New(fmt.Sprintf("Wrong KNEU client (%d) ID %s", kneuClientId, err))
	}

	secondaryDbCheckInterval, err := strconv.ParseInt(os.Getenv("SECONDARY_DB_CHECK_INTERVAL"), 10, 0)
	if secondaryDbCheckInterval == 0 || err != nil {
		secondaryDbCheckInterval = 10
	}

	appStartDelay, err := strconv.Atoi(os.Getenv("APP_START_DELAY"))
	if appStartDelay == 0 || err != nil {
		appStartDelay = 10
	}

	repeatScoreChangesTimeframeSeconds, err := strconv.Atoi(os.Getenv("TIMEFRAME_TO_COMBINE_REPEAT_SCORE_CHANGES"))
	if repeatScoreChangesTimeframeSeconds == 0 || err != nil {
		repeatScoreChangesTimeframeSeconds = 5
	}

	loadedConfig := Config{
		kafkaHost:                          os.Getenv("KAFKA_HOST"),
		kafkaTimeout:                       time.Second * time.Duration(kafkaTimeout),
		kafkaAttempts:                      kafkaAttempts,
		primaryDekanatDbDSN:                os.Getenv("PRIMARY_DEKANAT_DB_DSN"),
		secondaryDekanatDbDSN:              os.Getenv("SECONDARY_DEKANAT_DB_DSN"),
		secondaryDbCheckInterval:           time.Second * time.Duration(secondaryDbCheckInterval),
		sqsQueueUrl:                        os.Getenv("AWS_SQS_QUEUE_URL"),
		telegramToken:                      os.Getenv("TELEGRAM_TOKEN"),
		repeatScoreChangesTimeframeSeconds: time.Second * time.Duration(repeatScoreChangesTimeframeSeconds),
		appStartDelay:                      time.Second * time.Duration(appStartDelay),
		kneuBaseUrl:                        os.Getenv("KNEU_BASE_URI"),
		kneuClientId:                       kneuClientId,
		kneuClientSecret:                   os.Getenv("KNEU_CLIENT_SECRET"),
		publicUrl:                          os.Getenv("PUBLIC_URL"),
		skipWait:                           os.Getenv("SKIP_WAIT") == "true",
		debugUpdates:                       os.Getenv("DEBUG_UPDATES") == "true",
	}

	if loadedConfig.kafkaHost == "" {
		return Config{}, errors.New("empty KAFKA_HOST")
	}

	if loadedConfig.telegramToken == "" {
		err = errors.New("empty TELEGRAM_TOKEN")
	}

	if loadedConfig.kneuClientSecret == "" {
		return Config{}, errors.New("empty KNEU_CLIENT_SECRET")
	}

	return loadedConfig, nil
}

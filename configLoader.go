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
	kafkaHost             string
	kafkaTimeout          time.Duration
	kafkaAttempts         int
	primaryDekanatDbDSN   string
	secondaryDekanatDbDSN string
	telegramToken         string
	kneuBaseUrl           string
	kneuClientId          int
	kneuClientSecret      string
	publicUrl             string
	skipWait              bool
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

	loadedConfig := Config{
		kafkaHost:             os.Getenv("KAFKA_HOST"),
		kafkaTimeout:          time.Second * time.Duration(kafkaTimeout),
		kafkaAttempts:         kafkaAttempts,
		primaryDekanatDbDSN:   os.Getenv("PRIMARY_DEKANAT_DB_DSN"),
		secondaryDekanatDbDSN: os.Getenv("SECONDARY_DEKANAT_DB_DSN"),
		telegramToken:         os.Getenv("TELEGRAM_TOKEN"),
		kneuBaseUrl:           os.Getenv("KNEU_BASE_URI"),
		kneuClientId:          kneuClientId,
		kneuClientSecret:      os.Getenv("KNEU_CLIENT_SECRET"),
		publicUrl:             os.Getenv("PUBLIC_URL"),
		skipWait:              os.Getenv("SKIP_WAIT") == "true",
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

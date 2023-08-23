package main

import (
	"context"
	"fmt"
	"github.com/kneu-messenger-pigeon/events"
	"github.com/segmentio/kafka-go"
	"os"
	"time"
)

func WaitSecondaryDbScoreProcessedEvent() {
	var err error
	var m kafka.Message

	reader := kafka.NewReader(
		kafka.ReaderConfig{
			Brokers:     []string{config.kafkaHost},
			GroupID:     "integration-testing",
			Topic:       events.MetaEventsTopic,
			MinBytes:    10,
			MaxBytes:    10e3,
			MaxWait:     time.Second,
			MaxAttempts: config.kafkaAttempts,
			Dialer: &kafka.Dialer{
				Timeout:   config.kafkaTimeout,
				DualStack: kafka.DefaultDialer.DualStack,
			},
		},
	)

	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancelFunc()

	fmt.Println()
	for err == nil {
		m, err = reader.FetchMessage(ctx)

		if err == nil && string(m.Key) == events.SecondaryDbScoreProcessedEventName {
			fmt.Printf("Receive %s \n", string(m.Key))

			// wait while all message in topic will be processed
			time.Sleep(5 * time.Second)
			return
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}

	panic("SecondaryDbScoreProcessedEventName is not happen")
}

func WaitTelegramAppStarted() {
	timeLimit := time.Now().Add(5 * time.Minute)

	fmt.Print("waiting for call telegramAPI /getMe ")
	for mocks.TelegramMockServer.scopedGetMe.IsPending() && time.Now().Before(timeLimit) {
		fmt.Print(".")
		time.Sleep(3 * time.Second)
	}
	fmt.Println("")

	if mocks.TelegramMockServer.scopedGetMe.IsPending() {
		panic("getMeMock is still pending")
	}
}

package main

import (
	"context"
	"fmt"
	"github.com/kneu-messenger-pigeon/events"
	"github.com/segmentio/kafka-go"
	"os"
	"strings"
	"time"
)

func WaitKafkaEvent(topicName string, eventName string) {
	var err error
	var m kafka.Message

	reader := kafka.NewReader(
		kafka.ReaderConfig{
			Brokers:     []string{config.kafkaHost},
			Topic:       topicName,
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
	defer reader.Close()

	ctx, cancelFunc := context.WithTimeout(context.Background(), 1*time.Minute)
	eventReceivedCount := 0

	fmt.Println("Waiting for " + eventName + " event...")
	for ctx.Err() == nil {
		m, err = reader.FetchMessage(ctx)

		if err != nil && err != context.DeadlineExceeded {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)

		} else if strings.HasPrefix(string(m.Key), eventName) {
			eventReceivedCount++

			// wait for repeat event during 5 seconds
			cancelFunc()
			ctx, cancelFunc = context.WithTimeout(context.Background(), 5*time.Second)
		}

	}
	cancelFunc()

	if eventReceivedCount == 0 {
		fmt.Println(eventName + " is not happen")
	} else {
		fmt.Printf("Receive %s - count %d \n", eventName, eventReceivedCount)
	}
}

func WaitSecondaryDbScoreProcessedEvent() {
	WaitKafkaEvent(events.MetaEventsTopic, events.SecondaryDbScoreProcessedEventName)
}

func WaitScoreChangedEvent() {
	WaitKafkaEvent(events.ScoresChangesFeedTopic, events.ScoreChangedEventName)
}

func WaitTelegramAppStarted() {
	fmt.Print("waiting for call telegramAPI /getMe ")
	waitUntilCalled(mocks.TelegramMockServer.scopedGetMe, time.Minute)

	if mocks.TelegramMockServer.scopedGetMe.IsPending() {
		fmt.Println("getMeMock is still pending")
	}
}

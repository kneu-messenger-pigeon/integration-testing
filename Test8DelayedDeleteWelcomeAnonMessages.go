package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test8DelayedDeleteWelcomeAnonMessages(t *testing.T) {
	fmt.Println("Test8DelayedDeleteWelcomeAnonMessages")
	defer printTestResult(t, "Test8DelayedDeleteWelcomeAnonMessages")

	fmt.Printf("mocks.delayedDeleteMessageHandler.messages length: %d", len(mocks.delayedDeleteMessageHandler.messages))
	fmt.Printf("mocks.delayedDeleteMessageHandler.scopedMocks length: %d", len(mocks.delayedDeleteMessageHandler.scopedMocks))

	waitTime := time.Now().Add(config.authStateLifetime)
	for len(mocks.delayedDeleteMessageHandler.messages) != 0 && len(mocks.delayedDeleteMessageHandler.scopedMocks) == 0 && time.Now().Before(waitTime) {
		time.Sleep(time.Second * 5)
		fmt.Printf("wait for delete message scope mock to be added (count: %d)\n", len(mocks.delayedDeleteMessageHandler.messages))
	}

	assert.NotEmpty(t, mocks.delayedDeleteMessageHandler.messages)
	assert.NotEmpty(t, mocks.delayedDeleteMessageHandler.scopedMocks)

	calledCount := 0
	for _, scopedMock := range mocks.delayedDeleteMessageHandler.scopedMocks {
		if scopedMock.Called() {
			calledCount++
		}

		scopedMock.Clean()
	}

	fmt.Printf("calledCount: %d (total expactation : %d)\n", calledCount, len(mocks.delayedDeleteMessageHandler.scopedMocks))

	assert.NotEmpty(t, calledCount)
	assert.GreaterOrEqual(t, calledCount, len(mocks.delayedDeleteMessageHandler.scopedMocks)-4)
}

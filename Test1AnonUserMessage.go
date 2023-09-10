package main

import (
	"fmt"
	"testing"
)

func Test1AnonUserMessage(t *testing.T) {
	fmt.Println("Test1AnonUserMessage")

	userId := test1LoginUserId

	sendMessageMockScope := mocks.TelegramMockServer.mocha.AddMocks(expectAuthorizationMessage(userId))
	defer sendMessageMockScope.Clean()

	captureNotMatchedScope := captureNotMatchedSendMessage(userId)
	defer captureNotMatchedScope.Clean()

	<-mocks.TelegramMockServer.SendUpdate(TelegramUpdate{
		ID: 12344,
		Message: &Message{
			ID: 12344,
			Sender: &User{
				ID:       int64(userId),
				Username: "testUser1",
			},
			Text: "/list",
		},
	})

	sendMessageMockScope.AssertCalled(t)
	captureNotMatchedScope.AssertNotCalled(t)

	if !t.Failed() {
		fmt.Println("Test1AnonUserMessage passed")
	} else {
		fmt.Println("Test1AnonUserMessage failed")
	}
}

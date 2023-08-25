package main

import (
	"github.com/stretchr/testify/assert"
	"mvdan.cc/xurls/v2"
	"testing"
)

func Test2EnsureAuthFlow(t *testing.T) {
	userId := test2AuthFlowUserId

	catchText := &CatchMessageTextPostAction{}

	sendAuthorizationMessageMock := expectAuthorizationMessage(userId).PostAction(catchText)

	sendMessageMockScope := mocks.TelegramMockServer.mocha.AddMocks(sendAuthorizationMessageMock)
	defer sendMessageMockScope.Clean()

	<-mocks.TelegramMockServer.SendUpdate(TelegramUpdate{
		ID: 12344,
		Message: &Message{
			ID: 12344,
			Sender: &User{
				ID:       int64(userId),
				Username: "testUser2",
			},
			Text: "/start",
		},
	})

	sendMessageMockScope.AssertCalled(t)
	sendMessageMockScope.Clean()

	sendWelcomeMessageMock := expectWelcomeMessage(userId).PostAction(catchText)
	sendWelcomeMessageScope := mocks.TelegramMockServer.mocha.AddMocks(sendWelcomeMessageMock)

	captureNotMatchedScope := captureNotMatchedSendMessage(userId)
	defer captureNotMatchedScope.Clean()

	authUrl := xurls.Relaxed().FindString(catchText.Text)

	mocks.KneuAuthMockServer.EmulateAuthFlow(t, authUrl)

	sendWelcomeMessageScope.AssertCalled(t)
	captureNotMatchedScope.AssertNotCalled(t)

	assert.Equal(t, "Пане Петр, відтепер Ви будете отримувати сповіщення про нові оцінки!", catchText.Text)
}

package main

import (
	"github.com/vitorsalgado/mocha/v3"
	"github.com/vitorsalgado/mocha/v3/expect"
	"github.com/vitorsalgado/mocha/v3/reply"
	"mvdan.cc/xurls/v2"
	"testing"
)

func Test2EnsureAuthFlow(t *testing.T) {
	userId := test2AuthFlowUserId

	catchText := &CatchMessageTextPostAction{}

	sendMessageMock := mocha.Post(expect.URLPath("/sendMessage")).
		Repeat(1).
		Body(
			expectMarkdownV2, expectChatId(userId),
			expect.JSONPath("text", expect.ToContain("авториз")),
			expect.JSONPath("text", expect.ToContain(`redirect_uri`)),
		).
		Reply(reply.OK().BodyJSON(getSendMessageSuccessResponse())).
		PostAction(catchText)

	sendMessageMockScope := mocks.TelegramMockServer.mocha.AddMocks(sendMessageMock)
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

	captureNotMatchedScope := captureNotMatchedSendMessage(userId)
	defer captureNotMatchedScope.Clean()

	authUrl := xurls.Relaxed().FindString(catchText.Text)

	mocks.KneuAuthMockServer.EmulateAuthFlow(t, authUrl)

	captureNotMatchedScope.AssertNotCalled(t)
}

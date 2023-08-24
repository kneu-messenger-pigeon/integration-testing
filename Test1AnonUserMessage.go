package main

import (
	"fmt"
	"github.com/vitorsalgado/mocha/v3"
	"github.com/vitorsalgado/mocha/v3/expect"
	"github.com/vitorsalgado/mocha/v3/reply"
	"net/url"
	"regexp"
	"testing"
)

func Test1AnonUserMessage(t *testing.T) {
	userId := test1LoginUserId

	authUrl := escapeTelegramString(fmt.Sprintf(
		`%s/oauth?response_type=code&client_id=%d&redirect_uri=%s&state=`,
		config.kneuBaseUrl,
		config.kneuClientId,
		url.QueryEscape(config.publicUrl+"/authorizer/complete"),
	))

	jwtRegexp := `(?:[\w-\\]*(\.)){2}[\w-\\]*`

	authUrlRegexp, err := regexp.Compile(
		`\[.+?\]` +
			`\(` + regexp.QuoteMeta(authUrl) + jwtRegexp + `\)`,
	)
	if err != nil {
		panic("regexp.Compile failed: " + err.Error())
	}

	sendMessageMock := mocha.Post(expect.URLPath("/sendMessage")).
		Repeat(1).
		Body(
			expectMarkdownV2, expectChatId(userId),
			expect.JSONPath("text", expect.ToContain(authUrl)), expect.JSONPath("text", expect.ToMatchExpr(authUrlRegexp)),
		).
		Reply(
			reply.OK().BodyJSON(getSendMessageSuccessResponse()),
		)

	sendMessageMockScope := mocks.TelegramMockServer.mocha.AddMocks(sendMessageMock)
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
}

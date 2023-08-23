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

const testLoginUserId = 103

func TestLogin(t *testing.T) {
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
		Body(
			expectChatId(testLoginUserId),
			expect.JSONPath("parse_mode", expect.ToEqual("MarkdownV2")),
			expect.JSONPath("text", expect.ToContain(authUrl)),
			expect.JSONPath("text", expect.ToMatchExpr(authUrlRegexp)),
		).
		Reply(reply.OK())

	sendMessageMockScope := mocks.TelegramMockServer.mocha.AddMocks(sendMessageMock)
	defer sendMessageMockScope.Clean()

	captureFailedMock := mocha.Post(expect.URLPath("/sendMessage")).
		Body(expectChatId(testLoginUserId)).
		ReplyFunction(DumpRequest)

	captureFailedMockScope := mocks.TelegramMockServer.mocha.AddMocks(captureFailedMock)
	defer captureFailedMockScope.Clean()

	<-mocks.TelegramMockServer.SendUpdate(TelegramUpdate{
		ID: 12344,
		Message: &Message{
			ID: 12344,
			Sender: &User{
				ID:       testLoginUserId,
				Username: "testUser1",
			},
			Text: "/start",
		},
	})

	sendMessageMockScope.AssertCalled(t)
	captureFailedMockScope.AssertNotCalled(t)
}

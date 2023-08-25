package main

import (
	"encoding/json"
	"fmt"
	"github.com/vitorsalgado/mocha/v3"
	"github.com/vitorsalgado/mocha/v3/expect"
	"github.com/vitorsalgado/mocha/v3/reply"
	"net/url"
	"regexp"
	"strconv"
)

var sendMessageLastId = 100

var expectMarkdownV2 = expect.JSONPath("parse_mode", expect.ToEqual("MarkdownV2"))

func initTelegramHelpers(server *TelegramMockServer) {
	var err error
	server.authUrl = escapeTelegramString(fmt.Sprintf(
		`%s/oauth?response_type=code&client_id=%d&redirect_uri=%s&state=`,
		config.kneuBaseUrl,
		config.kneuClientId,
		url.QueryEscape(config.publicUrl+"/authorizer/complete"),
	))

	jwtRegexp := `(?:[\w-\\]*(\.)){2}[\w-\\]*`

	server.authUrlRegexp, err = regexp.Compile(
		`\[.+?\]` +
			`\(` + regexp.QuoteMeta(server.authUrl) + jwtRegexp + `\)`,
	)
	if err != nil {
		panic("regexp.Compile failed: " + err.Error())
	}
}

func getSendMessageSuccessResponse() map[string]interface{} {
	sendMessageLastId++
	return map[string]interface{}{
		"ok": true,
		"result": map[string]interface{}{
			"message_id": sendMessageLastId,
		},
	}
}

func expectChatId(chatId int) expect.Matcher {
	return expect.JSONPath(
		"chat_id",
		expect.ToEqual(strconv.Itoa(chatId)),
	)
}

func captureNotMatchedSendMessage(chatId int) *mocha.Scoped {
	return mocks.TelegramMockServer.mocha.AddMocks(
		mocha.Post(expect.URLPath("/sendMessage")).
			Body(expectChatId(chatId)).
			Repeat(0).
			ReplyFunction(DumpRequest),
	)
}

func expectAuthorizationMessage(chatId int) *mocha.MockBuilder {
	return mocha.Post(expect.URLPath("/sendMessage")).
		Repeat(1).
		Body(
			expectMarkdownV2, expectChatId(chatId),
			expect.JSONPath("text", expect.ToContain("авториз")),
			expect.JSONPath("text", expect.ToContain(mocks.TelegramMockServer.authUrl)),
			expect.JSONPath("text", expect.ToMatchExpr(mocks.TelegramMockServer.authUrlRegexp)),
		).
		Reply(reply.OK().BodyJSON(getSendMessageSuccessResponse()))
}

func expectWelcomeMessage(chatId int) *mocha.MockBuilder {
	return mocha.Post(expect.URLPath("/sendMessage")).
		Repeat(1).
		Body(
			expectMarkdownV2, expectChatId(chatId),
			expect.JSONPath("text", expect.ToContain("будете отримувати сповіщення")),
			expect.JSONPath(
				"reply_markup",
				expectJsonPayload(
					expect.JSONPath("keyboard.[0][0].text", expect.ToHavePrefix("/list")),
				),
			),
		).
		Reply(reply.OK().BodyJSON(getSendMessageSuccessResponse()))
}

func expectJsonPayload(matcher expect.Matcher) expect.Matcher {
	m := expect.Matcher{}
	m.Name = "JsonPayload"
	m.Matches = func(value any, a expect.Args) (bool, error) {
		jsonString, valid := value.(string)
		if !valid {
			return false, fmt.Errorf("expected string, got %T", value)
		}

		jsonPayload := map[string]interface{}{}
		err := json.Unmarshal([]byte(jsonString), &jsonPayload)
		if err != nil {
			return false, err
		}

		valid, err = matcher.Matches(jsonPayload, a)
		return valid, err
	}

	m.DescribeMismatch = func(p string, v any) string {
		return fmt.Sprintf("value is not valid JSON. value: %v", v)
	}

	return m
}

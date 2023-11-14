package main

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/vitorsalgado/mocha/v3"
	"github.com/vitorsalgado/mocha/v3/expect"
	"github.com/vitorsalgado/mocha/v3/reply"
	"integration-testing/expectjson"
	"mvdan.cc/xurls/v2"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

var sendMessageLastId = 100

var expectMarkdownV2 = expectjson.JSONPathOptional("parse_mode", expect.ToEqual("MarkdownV2"))

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

func getSendMessageSuccessResponse() *reply.StdReply {
	sendMessageLastId++
	return getEditMessageSuccessResponse()
}

func getEditMessageSuccessResponse() *reply.StdReply {
	return reply.OK().BodyJSON(map[string]interface{}{
		"ok": true,
		"result": map[string]interface{}{
			"message_id": sendMessageLastId,
		},
	})
}

func getDeleteMessageSuccessResponse() *reply.StdReply {
	return reply.OK().BodyJSON(map[string]interface{}{
		"ok": true,
	})
}

func logoutUser(chatId int) {
	catchMessage := &CatchMessagePostAction{}

	sendMessageScope := mocks.TelegramMockServer.mocha.AddMocks(
		mocha.Post(expect.URLPath("/sendMessage")).
			Body(expectMarkdownV2, expectChatId(chatId)).
			Repeat(1).
			Reply(getSendMessageSuccessResponse()).
			PostAction(catchMessage),
	)
	defer sendMessageScope.Clean()

	<-mocks.TelegramMockServer.SendUpdate(TelegramUpdate{
		ID: 12344,
		Message: &Message{
			ID: 12344,
			Sender: &User{
				ID:       int64(chatId),
				Username: "test",
			},
			Text: "/reset",
		},
	})

	waitUntilCalled(sendMessageScope, 5*time.Second)

	if len(catchMessage.Text) > 100 {
		catchMessage.Text = strings.Trim(catchMessage.Text[0:80], " ") + "..."
	}

	fmt.Println("logoutUser ", strconv.Itoa(chatId), " result: ", catchMessage.Text)
}

func loginUser(t *testing.T, chatId int, fakeUser *FakeUser, sender *User) {
	logoutUser(chatId)

	catchMessage := &CatchMessagePostAction{}

	expectAuthorizationMessageScope := mocks.TelegramMockServer.mocha.AddMocks(
		expectAuthorizationMessage(chatId).PostAction(catchMessage),
	)
	defer expectAuthorizationMessageScope.Clean()

	// 1. Send message to the bot
	<-mocks.TelegramMockServer.SendUpdate(TelegramUpdate{
		ID: 12344,
		Message: &Message{
			ID:     12344,
			Sender: sender,
			Text:   "/start",
		},
	})

	// 2. expect Welcome anon message with Authorization link
	expectAuthorizationMessageScope.AssertCalled(t)
	expectAuthorizationMessageScope.Clean()

	authUrl := xurls.Relaxed().FindString(catchMessage.Text)

	catchMessage.Reset()
	expectWelcomeMessageScope := mocks.TelegramMockServer.mocha.AddMocks(
		expectWelcomeMessage(chatId).PostAction(catchMessage),
	)

	// 3. go to auth url and finish auth
	mocks.KneuAuthMockServer.EmulateAuthFlow(t, authUrl, fakeUser)

	waitUntilCalled(expectWelcomeMessageScope, 10*time.Second)

	// 4. expect Welcome message with success
	expectWelcomeMessageScope.AssertCalled(t)
	expectWelcomeMessageScope.Clean()

	ok := assert.Equal(t, "Пані "+fakeUser.FirstName+", відтепер Ви будете отримувати сповіщення про нові оцінки!", catchMessage.Text)

	if !ok {
		t.FailNow()
	}
}

func expectChatId(chatId int) expect.Matcher {
	return expectjson.JSONPathOptional(
		"chat_id",
		expect.ToEqual(strconv.Itoa(chatId)),
	)
}

func expectMessageId(messageId int) expect.Matcher {
	return expectjson.JSONPathOptional(
		"message_id",
		expect.ToEqual(strconv.Itoa(messageId)),
	)
}

func captureNotMatchedSendMessage(chatId int) *mocha.Scoped {
	return mocks.TelegramMockServer.mocha.AddMocks(
		mocha.Post(expect.URLPath("/sendMessage")).
			Body(expectChatId(chatId)).
			Repeat(0).
			Reply(getSendMessageSuccessResponse()).
			PostAction(DumpRequest),
	)
}

func expectAuthorizationMessage(chatId int) *mocha.MockBuilder {
	return mocha.Post(expect.URLPath("/sendMessage")).
		Repeat(1).
		Body(
			expectMarkdownV2, expectChatId(chatId),
			expectjson.JSONPathOptional("text", expect.ToContain("авториз")),
			expectjson.JSONPathOptional("text", expect.ToContain(mocks.TelegramMockServer.authUrl)),
			expectjson.JSONPathOptional("text", expect.ToMatchExpr(mocks.TelegramMockServer.authUrlRegexp)),
		).
		Reply(getSendMessageSuccessResponse())
}

func expectWelcomeMessage(chatId int) *mocha.MockBuilder {
	return mocha.Post(expect.URLPath("/sendMessage")).
		Repeat(1).
		Body(
			expectMarkdownV2, expectChatId(chatId),
			expectjson.JSONPathOptional("text", expect.ToContain("будете отримувати сповіщення")),
			expectjson.JSONPathOptional(
				"reply_markup",
				expectJsonPayload(
					expectjson.JSONPathOptional("keyboard.[0][0].text", expect.ToHavePrefix("/list")),
				),
			),
		).
		Reply(getSendMessageSuccessResponse())
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

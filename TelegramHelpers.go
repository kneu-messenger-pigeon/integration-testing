package main

import (
	"github.com/vitorsalgado/mocha/v3"
	"github.com/vitorsalgado/mocha/v3/expect"
	"strconv"
)

var sendMessageLastId = 100

var expectMarkdownV2 = expect.JSONPath("parse_mode", expect.ToEqual("MarkdownV2"))

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

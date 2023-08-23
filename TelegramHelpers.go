package main

import (
	"github.com/vitorsalgado/mocha/v3/expect"
	"strconv"
)

func expectChatId(chatId int) expect.Matcher {
	return expect.JSONPath(
		"chat_id",
		expect.ToEqual(strconv.Itoa(chatId)),
	)
}

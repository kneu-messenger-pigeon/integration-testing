package main

import (
	"github.com/vitorsalgado/mocha/v3"
	"github.com/vitorsalgado/mocha/v3/expect"
	"time"
)

type StoredMessage struct {
	ChatId    int
	MessageId int
}

type DelayedDeleteMessageHandler struct {
	messages    []*StoredMessage
	scopedMocks []*mocha.Scoped
}

func NewDelayedDeleteMessageHandler() *DelayedDeleteMessageHandler {
	return &DelayedDeleteMessageHandler{
		messages:    make([]*StoredMessage, 0, 10),
		scopedMocks: make([]*mocha.Scoped, 0, 10),
	}
}

func (h *DelayedDeleteMessageHandler) AddMessage(chatId int, messageId int, timeout time.Duration) {
	h.messages = append(h.messages, &StoredMessage{
		ChatId:    chatId,
		MessageId: messageId,
	})

	go func() {
		time.Sleep(timeout - time.Second*10)

		deleteMessageMock := mocha.Post(expect.URLPath("/deleteMessage")).
			Body(expectChatId(chatId), expectMessageId(messageId)).
			Reply(getDeleteMessageSuccessResponse()).Repeat(1)

		expectDeleteMessageScope := mocks.TelegramMockServer.mocha.AddMocks(deleteMessageMock)

		time.Sleep(time.Second * 30)
		h.scopedMocks = append(h.scopedMocks, expectDeleteMessageScope)
	}()
}

package main

import (
	"encoding/json"
	"fmt"
	"github.com/vitorsalgado/mocha/v3"
	"github.com/vitorsalgado/mocha/v3/expect"
	"github.com/vitorsalgado/mocha/v3/params"
	"github.com/vitorsalgado/mocha/v3/reply"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type TelegramMockServer struct {
	mocha         *mocha.Mocha
	updates       chan TelegramUpdate
	scopedGetMe   *mocha.Scoped
	lastUpdateId  uint32
	authUrl       string
	authUrlRegexp *regexp.Regexp
}

func CreateTelegramMockServer(t *testing.T, token string) *TelegramMockServer {
	telegramMockServer := &TelegramMockServer{
		updates: make(chan TelegramUpdate),
	}

	configure := mocha.Configure()
	configure.Addr(telegramServerAddr)

	configure.Middlewares(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, "/bot"+token)
			r.RequestURI = r.URL.Path
			handler.ServeHTTP(w, r)
		})
	})

	telegramMockServer.mocha = mocha.New(t, configure.Build())

	getUpdatesMock := mocha.Post(expect.URLPath("/getUpdates")).
		ReplyFunction(telegramMockServer.telegramGetUpdatesHandler)

	telegramMockServer.mocha.AddMocks(getUpdatesMock)

	botId, _ := strconv.Atoi(strings.Split(config.telegramToken, ":")[0])
	getMeMock := mocha.Post(expect.URLPath("/getMe")).Reply(
		reply.OK().BodyJSON(
			map[string]interface{}{
				"ok": true,
				"result": map[string]interface{}{
					"id":         botId,
					"is_bot":     true,
					"first_name": "test",
					"username":   "test",
				},
			},
		),
	)

	telegramMockServer.scopedGetMe = telegramMockServer.mocha.AddMocks(getMeMock)

	telegramMockServer.mocha.Start()

	initTelegramHelpers(telegramMockServer)

	return telegramMockServer
}

func (mockServer *TelegramMockServer) telegramGetUpdatesHandler(r *http.Request, m reply.M, p params.P) (*reply.Response, error) {
	var updates = make([]TelegramUpdate, 0, 1)
	select {
	case update := <-mockServer.updates:
		update.ID = int(atomic.AddUint32(&mockServer.lastUpdateId, 1))
		updates = append(updates, update)

		go func() {
			time.Sleep(200 * time.Millisecond)
			update.SendDoneChan <- true
		}()

	case <-time.After(10 * time.Second):
	}

	response, err := reply.OK().
		BodyJSON(map[string]interface{}{
			"ok":     true,
			"result": updates,
		}).
		Build(r, m, p)

	updatesJSON, _ := json.Marshal(updates)

	if config.debugUpdates {
		fmt.Println("send updates: " + string(updatesJSON))
	}

	return response, err
}

func (mockServer *TelegramMockServer) telegramGetMeHandler(r *http.Request, m reply.M, p params.P) (*reply.Response, error) {
	response, err := reply.OK().
		BodyJSON(map[string]interface{}{
			"ok": true,
			"result": map[string]interface{}{
				"id":         123,
				"is_bot":     true,
				"first_name": "test",
				"username":   "test",
			},
		}).
		Build(r, m, p)

	return response, err
}

func (mockServer *TelegramMockServer) Close() {
	_ = mockServer.mocha.Close()
}

func (mockServer *TelegramMockServer) SendUpdate(update TelegramUpdate) <-chan bool {
	if update.Message != nil {
		mockServer.prepareMessage(update.Message)
	}

	if update.Callback != nil {
		if update.Callback.Message == nil {
			update.Callback.Message = &Message{
				ID:       12900,
				Sender:   update.Callback.Sender,
				Unixtime: time.Now().Unix(),
			}
		}

		mockServer.prepareMessage(update.Callback.Message)

		if update.Callback.MessageID == "" {
			update.Callback.MessageID = strconv.Itoa(update.Callback.Message.ID)
		}
	}

	update.SendDoneChan = make(chan bool, 1)
	mockServer.updates <- update
	return update.SendDoneChan
}

func (mockServer *TelegramMockServer) prepareMessage(message *Message) {
	if message.Chat == nil {
		message.Chat = &Chat{}
	}

	if message.Chat.ID == 0 && message.Sender != nil {
		message.Chat.ID = message.Sender.ID
	}

	if message.Chat.Type == "" {
		message.Chat.Type = ChatPrivate
	}
}

func (mockServer *TelegramMockServer) SendMessage(message *Message) {
	mockServer.SendUpdate(TelegramUpdate{
		Message: message,
	})
}

func (mockServer *TelegramMockServer) SendCallback(callback *Callback) {
	mockServer.SendUpdate(TelegramUpdate{
		Callback: callback,
	})
}

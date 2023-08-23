package main

import (
	"encoding/json"
	"fmt"
	"github.com/vitorsalgado/mocha/v3"
	"github.com/vitorsalgado/mocha/v3/expect"
	"github.com/vitorsalgado/mocha/v3/params"
	"github.com/vitorsalgado/mocha/v3/reply"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type TelegramMockServer struct {
	mocha        *mocha.Mocha
	updates      chan TelegramUpdate
	scopedGetMe  *mocha.Scoped
	lastUpdateId uint32
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
	fmt.Println("send updates: " + string(updatesJSON))

	return response, err
}

func (mockServer *TelegramMockServer) telegramGetMeHandler(r *http.Request, m reply.M, p params.P) (*reply.Response, error) {
	fmt.Println("sadadsads 123 adsadsads")

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
		if update.Message.Chat == nil {
			update.Message.Chat = &Chat{}
		}

		if update.Message.Chat.ID == 0 && update.Message.Sender != nil {
			update.Message.Chat.ID = update.Message.Sender.ID
		}
		if update.Message.Chat.Type == "" {
			update.Message.Chat.Type = ChatPrivate
		}
	}

	update.SendDoneChan = make(chan bool, 1)
	mockServer.updates <- update
	return update.SendDoneChan
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

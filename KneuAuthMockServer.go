package main

import (
	"github.com/vitorsalgado/mocha/v3"
	"github.com/vitorsalgado/mocha/v3/expect"
	"github.com/vitorsalgado/mocha/v3/reply"
	"testing"
)

type KneuAuthMockServer struct {
	mocha        *mocha.Mocha
	updates      chan TelegramUpdate
	lastUpdateId uint32
}

func CreateKneuAuthMockServer(t *testing.T, clientId int, clientSecret string) *KneuAuthMockServer {
	kneuAuthMockServer := &KneuAuthMockServer{
		updates: make(chan TelegramUpdate),
	}

	configure := mocha.Configure()
	configure.Addr(kneuAuthServerAddr)

	kneuAuthMockServer.mocha = mocha.New(t, configure.Build())

	kneuAuthMockServer.mocha.AddMocks(
		mocha.Get(expect.URLPath("/auth")).
			Reply(
				reply.OK().BodyString("<h1>2123</h1>"),
			),
	)

	kneuAuthMockServer.mocha.Start()

	return kneuAuthMockServer
}

func (mockServer *KneuAuthMockServer) Close() {
	_ = mockServer.mocha.Close()
}

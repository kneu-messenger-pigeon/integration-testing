package main

import "testing"

type Mocks struct {
	TelegramMockServer *TelegramMockServer
	KneuAuthMockServer *KneuAuthMockServer
}

func createMocks(t *testing.T, config Config) *Mocks {
	return &Mocks{
		TelegramMockServer: CreateTelegramMockServer(t, config.telegramToken),
		KneuAuthMockServer: CreateKneuAuthMockServer(t, config.kneuClientId, config.kneuClientSecret),
	}
}

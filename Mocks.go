package main

import (
	"database/sql"
	"testing"
)

type Mocks struct {
	PrimaryDB          *sql.DB
	SecondaryDB        *sql.DB
	TelegramMockServer *TelegramMockServer
	KneuAuthMockServer *KneuAuthMockServer
	RealtimeQueue      *RealtimeQueue
}

func createMocks(t *testing.T, config Config) *Mocks {
	return &Mocks{
		PrimaryDB:          OpenDbConnection(t, config.primaryDekanatDbDSN),
		SecondaryDB:        OpenDbConnection(t, config.secondaryDekanatDbDSN),
		TelegramMockServer: CreateTelegramMockServer(t, config.telegramToken),
		KneuAuthMockServer: CreateKneuAuthMockServer(t, config.kneuClientId, config.kneuClientSecret),
		RealtimeQueue:      CreateRealtimeQueue(t),
	}
}

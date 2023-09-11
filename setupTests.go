package main

import "testing"

func setupTests(t *testing.T) {
	t.Run("Test1AnonUserMessage", Test1AnonUserMessage)
	t.Run("Test2EnsureAuthFlow", Test2EnsureAuthFlow)
	t.Run("Test3SecondaryDatabaseUpdates", Test3SecondaryDatabaseUpdates)
	t.Run("Test4PhantomEdit", Test4PhantomEdit)
	t.Run("Test5BotDeactivatedByUser", Test5BotDeactivatedByUser)
	t.Run("Test6RealtimeImport", Test6RealtimeImport)
}

package main

import "testing"

func setupTests(t *testing.T) {
	t.Run("Test1AnonUserMessage", Test1AnonUserMessage)
	t.Run("Test2EnsureAuthFlow", Test2EnsureAuthFlow)
	t.Run("Test3SecondaryDatabaseUpdates", Test3SecondaryDatabaseUpdates)
}

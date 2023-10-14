package main

import (
	"fmt"
	"testing"
)

func Test7RealtimeImportCustomGroup(t *testing.T) {
	fmt.Println("Test7RealtimeImportCustomGroup")
	defer printTestResult(t, "Test7RealtimeImportCustomGroup")

	TestRealtimeImport(t,
		test7RealtimeImportCustomGroupUserId,
		198569, "Моделювання інформаційних систем і компʼютерних мереж",
		520650,
	)
}

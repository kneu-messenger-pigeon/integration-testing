package main

import (
	"fmt"
	"testing"
)

func Test6RealtimeImportRegularGroup(t *testing.T) {
	fmt.Println("Test6RealtimeImportRegularGroup")
	defer printTestResult(t, "Test6RealtimeImportRegularGroup")

	TestRealtimeImport(t,
		test6RealtimeImportRegularGroupUserId,
		198569, "Моделювання інформаційних систем і компʼютерних мереж",
		0,
	)
}

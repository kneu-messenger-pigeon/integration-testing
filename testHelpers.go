package main

import "testing"

func printTestResult(t *testing.T, testName string) {
	if !t.Failed() {
		println("✅ " + testName + " passed")
	} else {
		println("❌ " + testName + " failed")
	}
}

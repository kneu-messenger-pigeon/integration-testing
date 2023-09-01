package main

import (
	"github.com/vitorsalgado/mocha/v3"
	"time"
)

func waitUntilCalled(scope *mocha.Scoped, timeout time.Duration) {
	untilTime := time.Now().Add(timeout)
	for scope.IsPending() && time.Now().Before(untilTime) {
		time.Sleep(time.Millisecond * 200)
	}
}

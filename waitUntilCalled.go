package main

import (
	"github.com/vitorsalgado/mocha/v3"
	"time"
)

func waitUntilCalled(scope *mocha.Scoped, timeout time.Duration) {
	waitUntilCalledTimes(scope, timeout, 1)
}

func waitUntilCalledTimes(scope *mocha.Scoped, timeout time.Duration, times int) {
	untilTime := time.Now().Add(timeout)
	for scope.Hits() < times && time.Now().Before(untilTime) {
		time.Sleep(time.Millisecond * 200)
	}
}

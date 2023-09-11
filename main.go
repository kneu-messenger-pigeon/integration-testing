package main

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"
)

var mocks *Mocks

var config Config

var stubMatch = func(pat, str string) (bool, error) { return true, nil }

func main() {
	var err error
	// Empty main function
	envFilename := ""
	if _, err = os.Stat(".env"); err == nil {
		envFilename = ".env"
	}

	config, err = loadConfig(envFilename)
	if err != nil {
		log.Fatalln(err)
	}

	test := testing.InternalTest{
		Name: "integration testing",
		F: func(t *testing.T) {
			mocks = createMocks(t, config)

			if t.Failed() {
				return
			}

			if !config.skipWait {
				WaitTelegramAppStarted()
				WaitSecondaryDbScoreProcessedEvent()
				WaitScoreChangedEvent()

				fmt.Printf("App is started. Wait %d seconds for app to be ready..\n", int(config.appStartDelay.Seconds()))
				time.Sleep(config.appStartDelay)
			}

			fmt.Println("Start testing..")
			setupTests(t)
			fmt.Println("Test done")
		},
	}

	testing.Main(stubMatch, []testing.InternalTest{test}, []testing.InternalBenchmark{}, []testing.InternalExample{})
}

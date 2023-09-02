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

var fakeMatchString = func(pat, str string) (bool, error) { return true, nil }

func main() {
	var err error
	// Empty main function
	envFilename := ""
	if _, err := os.Stat(".env"); err == nil {
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

			if !config.skipWait {
				WaitTelegramAppStarted()
				WaitSecondaryDbScoreProcessedEvent()
				WaitScoreChangedEvent()
				time.Sleep(10 * time.Second)
			}

			fmt.Println("App is started. Start testing..")
			setupTests(t)
		},
	}

	testing.Main(
		func(pat, str string) (bool, error) { return true, nil },
		[]testing.InternalTest{test},
		[]testing.InternalBenchmark{}, []testing.InternalExample{},
	)

}

package main

import (
	"fmt"
	"log"
	"testing"
)

var mocks *Mocks

var config Config

func main() {
	var err error
	// Empty main function
	config, err = loadConfig(".env")
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
			}

			fmt.Println("App is started. Start testing..")
			t.Run("TestLogin", TestLogin)
		},
	}

	testing.Main(
		func(pat, str string) (bool, error) { return true, nil },
		[]testing.InternalTest{test},
		[]testing.InternalBenchmark{}, []testing.InternalExample{},
	)

}

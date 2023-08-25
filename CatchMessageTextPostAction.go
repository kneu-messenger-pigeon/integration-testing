package main

import (
	"encoding/json"
	"errors"
	"github.com/vitorsalgado/mocha/v3"
	"io"
)

type CatchMessageTextPostAction struct {
	Text string
}

func (c *CatchMessageTextPostAction) Run(args mocha.PostActionArgs) error {
	bodyBuffer, err := io.ReadAll(args.Request.Body)

	if err == nil {
		var response map[string]interface{}
		err = json.Unmarshal(bodyBuffer, &response)
		text, ok := response["text"].(string)

		if !ok {
			err = errors.New("Request has not text field")
		} else {
			c.Text = unescapeTelegramString(text)
		}
	}

	return err
}

type CatchMessageTextPostActionInterface interface {
	Run(args mocha.PostActionArgs)
}

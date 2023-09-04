package main

import (
	"encoding/json"
	"fmt"
	"github.com/vitorsalgado/mocha/v3"
	"io"
	"sync"
)

type CatchMessagePostAction struct {
	Text            string `json:"text"`
	MessageId       string `json:"message_id"`
	ReplyMarkup     *ReplyMarkup
	ReplyMarkupJson string `json:"reply_markup"`
	mutex           sync.Mutex
}

func (c *CatchMessagePostAction) Reset() {
	c.Text = ""
	c.ReplyMarkup = nil
	c.ReplyMarkupJson = ""
}

func (c *CatchMessagePostAction) Run(args mocha.PostActionArgs) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.Reset()
	bodyBuffer, err := io.ReadAll(args.Request.Body)
	if err != nil {
		fmt.Println("body read error: ", err)
		return err
	}

	err = json.Unmarshal(bodyBuffer, &c)
	if err != nil {
		fmt.Println("body decode error: ", err)
		return err
	}

	c.Text = unescapeTelegramString(c.Text)

	if c.ReplyMarkupJson != "" {
		err = json.Unmarshal([]byte(c.ReplyMarkupJson), &c.ReplyMarkup)
		c.ReplyMarkupJson = ""
		if err != nil {
			fmt.Println("reply_markup decode error: ", err)
			return err
		}
	}

	return nil
}

func (c *CatchMessagePostAction) GetInlineButton(index int) *InlineButton {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.ReplyMarkup == nil || len(c.ReplyMarkup.InlineKeyboard) == 0 || len(c.ReplyMarkup.InlineKeyboard[0]) <= index {
		return nil
	}

	return &c.ReplyMarkup.InlineKeyboard[0][index]
}

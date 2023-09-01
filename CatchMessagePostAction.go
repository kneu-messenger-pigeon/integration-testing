package main

import (
	"encoding/json"
	"github.com/vitorsalgado/mocha/v3"
	"io"
)

type CatchMessagePostAction struct {
	Text            string `json:"text"`
	ReplyMarkup     *ReplyMarkup
	ReplyMarkupJson string `json:"reply_markup"`
}

func (c *CatchMessagePostAction) Reset() {
	c.Text = ""
	c.ReplyMarkup = nil
	c.ReplyMarkupJson = ""
}

func (c *CatchMessagePostAction) Run(args mocha.PostActionArgs) error {
	c.Reset()
	bodyBuffer, err := io.ReadAll(args.Request.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bodyBuffer, &c)
	if err != nil {
		return err
	}

	c.Text = unescapeTelegramString(c.Text)

	if c.ReplyMarkupJson != "" {
		err = json.Unmarshal([]byte(c.ReplyMarkupJson), &c.ReplyMarkup)
		c.ReplyMarkupJson = ""
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *CatchMessagePostAction) GetInlineButton(index int) *InlineButton {
	if c.ReplyMarkup == nil || len(c.ReplyMarkup.InlineKeyboard) == 0 || len(c.ReplyMarkup.InlineKeyboard[0]) == 0 {
		return nil
	}

	if len(c.ReplyMarkup.InlineKeyboard[0]) <= index {
		return nil
	}

	return &c.ReplyMarkup.InlineKeyboard[0][index]
}

package main

import "strings"

var escapeChar = []string{"[", "]", "(", ")", ">", "#", "+", "-", "=", "{", "}", ".", "!"}

func escapeTelegramString(markdownStr string) string {
	// do not escape special chars to keep format: * bold; _ italic; ~strikethrough; | - spoiler; ` - code
	for _, char := range escapeChar {
		if strings.Contains(markdownStr, char) {
			markdownStr = strings.ReplaceAll(markdownStr, char, "\\"+char)
		}
	}

	return markdownStr
}

func unescapeTelegramString(markdownStr string) string {
	// do not escape special chars to keep format: * bold; _ italic; ~strikethrough; | - spoiler; ` - code
	for _, char := range escapeChar {
		if strings.Contains(markdownStr, "\\"+char) {
			markdownStr = strings.ReplaceAll(markdownStr, "\\"+char, char)
		}
	}

	return markdownStr
}

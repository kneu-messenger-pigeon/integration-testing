package main

import "strings"

func escapeTelegramString(markdownStr string) string {
	// do not escape special chars to keep format: * bold; _ italic; ~strikethrough; | - spoiler; ` - code
	escapeChar := []string{"[", "]", "(", ")", ">", "#", "+", "-", "=", "{", "}", ".", "!"}
	for _, char := range escapeChar {
		if strings.Contains(markdownStr, char) {
			markdownStr = strings.ReplaceAll(markdownStr, char, "\\"+char)
		}
	}

	// drop escaping from inline links
	//	markdownStr = unEscapeMarkDownLinks.ReplaceAllString(markdownStr, unEscapeMarkDownLinksSubstitution)

	return markdownStr
}

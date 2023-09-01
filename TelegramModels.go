package main

type TelegramUpdate struct {
	ID int `json:"update_id"`

	Message  *Message  `json:"message,omitempty"`
	Callback *Callback `json:"callback_query,omitempty"`

	SendDoneChan chan bool `json:"-"`
}

type Callback struct {
	ID string `json:"id"`

	// For message sent to channels, Sender may be empty
	Sender *User `json:"from"`

	// Message will be set if the button that originated the query
	// was attached to a message sent by a bot.
	Message *Message `json:"message"`

	// MessageID will be set if the button was attached to a message
	// sent via the bot in inline mode.
	MessageID string `json:"inline_message_id"`

	// Data associated with the callback button. Be aware that
	// a bad client can send arbitrary data in this field.
	Data string `json:"data"`
}

// Message object represents a message.
type Message struct {
	ID int `json:"message_id"`

	// For message sent to channels, Sender will be nil
	Sender *User `json:"from"`

	// Unixtime, use Message.Time() to get time.Time
	Unixtime int64 `json:"date"`

	// Conversation the message belongs to.
	Chat *Chat `json:"chat"`

	// Note that the Message object in this field will not
	// contain further ReplyTo fields even if it
	// itself is a reply.
	ReplyTo *Message `json:"reply_to_message"`

	// For a text message, the actual UTF-8 text of the message.
	Text string `json:"text"`

	// Inline keyboard attached to the message.
	ReplyMarkup *ReplyMarkup `json:"reply_markup,omitempty"`
}

// User object represents a Telegram user, bot.
type User struct {
	ID int64 `json:"id"`

	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
	IsBot        bool   `json:"is_bot"`
	IsPremium    bool   `json:"is_premium"`
	AddedToMenu  bool   `json:"added_to_attachment_menu"`

	// Returns only in getMe
	CanJoinGroups   bool `json:"can_join_groups"`
	CanReadMessages bool `json:"can_read_all_group_messages"`
	SupportsInline  bool `json:"supports_inline_queries"`
}

// ChatType represents one of the possible chat types.
type ChatType string

const (
	ChatPrivate        ChatType = "private"
	ChatGroup          ChatType = "group"
	ChatSuperGroup     ChatType = "supergroup"
	ChatChannel        ChatType = "channel"
	ChatChannelPrivate ChatType = "privatechannel"
)

// Chat object represents a Telegram user, bot, group or a channel.
type Chat struct {
	ID int64 `json:"id"`

	// See ChatType and consts.
	Type ChatType `json:"type"`

	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

// ReplyMarkup controls two convenient options for bot-user communications
// such as reply keyboard and inline "keyboard" (a grid of buttons as a part
// of the message).
type ReplyMarkup struct {
	// InlineKeyboard is a grid of InlineButtons displayed in the message.
	//
	// Note: DO NOT confuse with ReplyKeyboard and other keyboard properties!
	InlineKeyboard [][]InlineButton `json:"inline_keyboard,omitempty"`

	// ReplyKeyboard is a grid, consisting of keyboard buttons.
	//
	// Note: you don't need to set HideCustomKeyboard field to show custom keyboard.
	ReplyKeyboard [][]ReplyButton `json:"keyboard,omitempty"`

	// ForceReply forces Telegram clients to display
	// a reply interface to the user (act as if the user
	// has selected the bot‘s message and tapped "Reply").
	ForceReply bool `json:"force_reply,omitempty"`

	// Requests clients to resize the keyboard vertically for optimal fit
	// (e.g. make the keyboard smaller if there are just two rows of buttons).
	//
	// Defaults to false, in which case the custom keyboard is always of the
	// same height as the app's standard keyboard.
	ResizeKeyboard bool `json:"resize_keyboard,omitempty"`

	// Requests clients to hide the reply keyboard as soon as it's been used.
	//
	// Defaults to false.
	OneTimeKeyboard bool `json:"one_time_keyboard,omitempty"`

	// Requests clients to remove the reply keyboard.
	//
	// Defaults to false.
	RemoveKeyboard bool `json:"remove_keyboard,omitempty"`
}

// InlineButton represents a button displayed in the message.
type InlineButton struct {
	// Unique slagish name for this kind of button,
	// try to be as specific as possible.
	//
	// It will be used as a callback endpoint.
	Unique string `json:"unique,omitempty"`
	Text   string `json:"text"`
	URL    string `json:"url,omitempty"`
	Data   string `json:"callback_data,omitempty"`
}

type ReplyButton struct {
	Text string `json:"text"`

	Contact  bool `json:"request_contact,omitempty"`
	Location bool `json:"request_location,omitempty"`
}

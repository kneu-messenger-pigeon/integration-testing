package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/vitorsalgado/mocha/v3"
	"github.com/vitorsalgado/mocha/v3/expect"
	"strings"
	"testing"
	"time"
)

func Test2EnsureAuthFlow(t *testing.T) {
	fmt.Println("Test2EnsureAuthFlow")

	userId := test2AuthFlowUserId
	fakeUser := &FakeUser{
		Id:         220,
		StudentId:  113508,
		GroupId:    16880,
		LastName:   "Степанченко",
		FirstName:  "Ірина",
		MiddleName: "Володимирівна",
		Gender:     "female",
	}

	sender := &User{
		ID:       int64(userId),
		Username: "testUser2",
	}

	loginUser(t, userId, fakeUser, sender)

	catchMessage := &CatchMessagePostAction{}
	sendDisciplinesListMessageMock := mocha.Post(expect.URLPath("/sendMessage")).
		Repeat(1).
		Body(
			expectMarkdownV2, expectChatId(userId),
			expect.JSONPath("text", expect.ToContain("Ваша загальна успішність у навчанні")),
		).
		Reply(getSendMessageSuccessResponse()).
		PostAction(catchMessage)

	sendDisciplinesListMessageScope := mocks.TelegramMockServer.mocha.AddMocks(sendDisciplinesListMessageMock)
	defer sendDisciplinesListMessageScope.Clean()

	// 5. list command
	<-mocks.TelegramMockServer.SendUpdate(TelegramUpdate{
		ID: 12344,
		Message: &Message{
			ID:     12344,
			Sender: sender,
			Text:   "/list",
		},
	})

	// 6. expect a discipline list message
	sendDisciplinesListMessageScope.AssertCalled(t)
	sendDisciplinesListMessageScope.Clean()

	catchMessage.Text = strings.Trim(catchMessage.Text, " \n")
	lines := strings.Split(catchMessage.Text, "\n")

	if !assert.GreaterOrEqual(t, len(lines), 5) {
		return
	}

	assert.Equal(t, "Пані "+fakeUser.FirstName+", Ваша загальна успішність у навчанні:", lines[0])

	assert.Equal(t, lines[3], "1. Системи управління знаннями")
	assert.Equal(t, lines[4], "     *результат 24*, _рейтинг #3/3_")

	assert.Equal(t, "Вимкнути бот - /reset", lines[len(lines)-5])
	assert.Equal(t, "❗Увага❗", lines[len(lines)-3])
	assert.Equal(t, "Перевіряйте оцінки в [офіційному журналі успішності КНЕУ](https://cutt.ly/Dekanat)", lines[len(lines)-2])
	assert.Equal(t, "Цей Бот не є офіційним джерелом даних про успішність.", lines[len(lines)-1])

	// 7. press discipline button
	firstButton := catchMessage.GetInlineButton(0)
	if !assert.NotNil(t, firstButton) {
		return
	}
	assert.Equal(t, "Системи управління знаннями", firstButton.Text)

	// 8. expect discipline score
	catchMessage.Reset()
	sendDisciplinesScoresMessageMock := mocha.Post(expect.URLPath("/sendMessage")).
		Repeat(1).
		Body(
			expectMarkdownV2, expectChatId(userId),
			expect.JSONPath("text", expect.ToHavePrefix("*Системи управління знаннями*")),
		).
		Reply(getSendMessageSuccessResponse()).
		PostAction(catchMessage)

	sendDisciplinesScoresMessageScope := mocks.TelegramMockServer.mocha.AddMocks(sendDisciplinesScoresMessageMock)
	defer sendDisciplinesScoresMessageScope.Clean()

	<-mocks.TelegramMockServer.SendUpdate(TelegramUpdate{
		ID: 12344,
		Callback: &Callback{
			ID:     "12356",
			Sender: sender,
			Data:   firstButton.Data,
		},
	})

	sendDisciplinesScoresMessageScope.AssertCalled(t)

	// 8. expect discipline score
	lines = strings.Split(catchMessage.Text, "\n")

	if !assert.GreaterOrEqual(t, len(lines), 6) {
		return
	}
	assert.Equal(t, "*Системи управління знаннями*: 24", lines[0])
	assert.Equal(t, "рейтинг #3/3", lines[1])
	assert.Equal(t, "", lines[2])
	assert.Equal(t, "Загалом по групі: max 32, min 24", lines[3])
	assert.Equal(t, "", lines[4])
	assert.Equal(t, "03.07.2023 *24* _Лабораторна роб._", lines[5])

	if !assert.NotNil(t, catchMessage.GetInlineButton(0)) {
		return
	}
	assert.Equal(t, "Назад", catchMessage.GetInlineButton(0).Text)

	catchMessage.Reset()

	sendMessageNotificationStoppedScope := mocks.TelegramMockServer.mocha.AddMocks(
		mocha.Post(expect.URLPath("/sendMessage")).
			Repeat(1).
			Body(
				expectMarkdownV2, expectChatId(userId),
				expect.JSONPath("text", expect.ToHavePrefix("Відтепер надсилання сповіщень зупинено")),
			).
			Reply(getSendMessageSuccessResponse()).
			PostAction(catchMessage),
	)
	defer sendMessageNotificationStoppedScope.Clean()

	// 9. reset command
	<-mocks.TelegramMockServer.SendUpdate(TelegramUpdate{
		ID: 12344,
		Message: &Message{
			ID:     12344,
			Sender: sender,
			Text:   "/reset",
		},
	})

	waitUntilCalled(sendMessageNotificationStoppedScope, 5*time.Second)

	sendMessageNotificationStoppedScope.AssertCalled(t)

	expectAuthorizationMessageAfterResetScope := mocks.TelegramMockServer.mocha.AddMocks(expectAuthorizationMessage(userId))
	defer expectAuthorizationMessageAfterResetScope.Clean()

	// 11. press discipline button
	<-mocks.TelegramMockServer.SendUpdate(TelegramUpdate{
		ID: 12344,
		Callback: &Callback{
			ID:     "12356",
			Sender: sender,
			Data:   firstButton.Data,
		},
	})

	// 12. expect Welcome anon message
	expectAuthorizationMessageAfterResetScope.AssertCalled(t)
}

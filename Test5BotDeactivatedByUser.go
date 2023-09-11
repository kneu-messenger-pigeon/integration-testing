package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/vitorsalgado/mocha/v3"
	"github.com/vitorsalgado/mocha/v3/expect"
	"github.com/vitorsalgado/mocha/v3/reply"
	"testing"
	"time"
)

func createScoresForTest5(
	t *testing.T, fakeUser *FakeUser,
	disciplineId int, lessonDate time.Time, score1Value int,
) (score1Id int) {
	lesson := &Lesson{
		GroupId:      fakeUser.GroupId,
		DisciplineId: disciplineId,
		Semester:     2,
		LessonTypeId: 3,
		LessonDate:   lessonDate,
		TeachId:      6479,
		TeachUserId:  2715,
		RegDate:      lessonDate,
	}

	lesson.LessonId = AddLesson(t, mocks.SecondaryDB, *lesson)

	score1 := &Score{
		Lesson:     lesson,
		StudentId:  fakeUser.StudentId,
		LessonPart: 1,
		Score:      score1Value,
	}

	score1Id = AddScore(t, mocks.SecondaryDB, score1)

	fmt.Printf("Create lesson %d with score: %d\n", lesson.LessonId, score1Id)

	UpdateDbDatetimeAndWait(t, mocks.SecondaryDB, lessonDate.Add(time.Hour*1))

	return
}

func Test5BotDeactivatedByUser(t *testing.T) {
	fmt.Println("Test5BotDeactivatedByUser")
	defer printTestResult(t, "Test5BotDeactivatedByUser")

	startRegDate := time.Date(2023, 7, 6, 12, 0, 0, 0, time.UTC)
	UpdateDbDatetime(t, mocks.SecondaryDB, startRegDate)

	userId := Test5BotDeactivatedUserId
	fakeUser := &FakeUser{
		Id:         220,
		StudentId:  113507,
		GroupId:    16880,
		LastName:   "Стеценко",
		FirstName:  "Анна",
		MiddleName: "Олегівна",
		Gender:     "female",
	}

	sender := &User{
		ID:       int64(userId),
		Username: "testUser2",
	}

	loginUser(t, userId, fakeUser, sender)

	catchMessage := &CatchMessagePostAction{}
	expectNewScoreMessageScope := mocks.TelegramMockServer.mocha.AddMocks(
		mocha.Post(expect.URLPath("/sendMessage")).
			Body(expectMarkdownV2, expectChatId(userId)).
			Repeat(2).
			Reply(reply.Seq().
				Add(reply.BadRequest().BodyJSON(map[string]interface{}{
					"ok":          false,
					"error_code":  403,
					"description": "Forbidden: bot was blocked by the user",
				})).
				Add(reply.BadRequest().BodyJSON(map[string]interface{}{"ok": true})),
			).
			PostAction(catchMessage),
	)
	defer expectNewScoreMessageScope.Clean()

	expectDisciplineName := "Нейрокомпʼютерні системи"
	expectDisciplineId := 198568

	scoreValue := 3
	createScoresForTest5(t, fakeUser, expectDisciplineId, startRegDate.Add(time.Minute*30), scoreValue)

	expectedText := fmt.Sprintf(
		"Новий запис: %s, заняття %s _Лабораторна роб._: %d",
		expectDisciplineName, startRegDate.Format("02.01.2006"), 3,
	)

	waitUntilCalled(expectNewScoreMessageScope, 10*time.Second)
	expectNewScoreMessageScope.AssertCalled(t)
	assert.Equal(t, 1, expectNewScoreMessageScope.Hits())
	assert.Equal(t, expectedText, catchMessage.Text)

	// wait to unregister user after error
	catchMessage.Reset()
	waitUntilCalledTimes(expectNewScoreMessageScope, 10*time.Second, 2)

	assert.Equal(t, 2, expectNewScoreMessageScope.Hits())
	assert.Equal(t, "Відтепер надсилання сповіщень зупинено.", catchMessage.Text)
	fmt.Println("catchMessage.Text", catchMessage.Text)

	expectNewScoreMessageScope.Clean()

	// try to send message - expect welcome auth message
	sendMessageMockScope := mocks.TelegramMockServer.mocha.AddMocks(expectAuthorizationMessage(userId))
	defer sendMessageMockScope.Clean()

	<-mocks.TelegramMockServer.SendUpdate(TelegramUpdate{
		ID: 12544,
		Message: &Message{
			ID:     12544,
			Sender: sender,
			Text:   "/list",
		},
	})
	sendMessageMockScope.AssertCalled(t)
}

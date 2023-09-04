package main

import (
	"database/sql"
	"fmt"
	_ "github.com/nakagami/firebirdsql"
	"github.com/stretchr/testify/assert"
	"github.com/vitorsalgado/mocha/v3"
	"github.com/vitorsalgado/mocha/v3/expect"
	"strconv"
	"testing"
	"time"
)

func createScoresForTest3(
	t *testing.T, secondaryDekanatDb *sql.DB, fakeUser *FakeUser,
	disciplineId int, lessonDate time.Time,
	score1Value int, score2Value int,
) {
	lesson := &Lesson{
		GroupId:      fakeUser.GroupId,
		DisciplineId: disciplineId,
		Semester:     2,
		LessonTypeId: 15,
		LessonDate:   lessonDate,
		TeachId:      6479,
		TeachUserId:  2715,
	}

	lesson.LessonId = AddLesson(t, secondaryDekanatDb, *lesson)

	score1 := Score{
		Lesson:     lesson,
		StudentId:  fakeUser.StudentId,
		LessonPart: 1,
		Score:      score1Value,
	}

	score1Id := AddScore(t, secondaryDekanatDb, score1)

	score2 := Score{
		Lesson:     lesson,
		StudentId:  fakeUser.StudentId,
		LessonPart: 2,
		Score:      score2Value,
	}

	score2Id := AddScore(t, secondaryDekanatDb, score2)

	fmt.Printf("Create lesson %d with two scores: %d and %d\n", lesson.LessonId, score1Id, score2Id)

	UpdateDbDatetimeAndWait(t, secondaryDekanatDb, lessonDate)
}

func Test3SecondaryDatabaseUpdates(t *testing.T) {
	fmt.Println("Test3SecondaryDatabaseUpdates")

	secondaryDekanatDb, err := sql.Open("firebirdsql", config.secondaryDekanatDbDSN)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	defer secondaryDekanatDb.Close()

	err = secondaryDekanatDb.Ping()
	assert.NoError(t, err, "failed to ping secondary db")

	userId := test3SecondaryDatabaseUserId
	fakeUser := &FakeUser{
		Id:         320,
		StudentId:  113509,
		GroupId:    16880,
		LastName:   "Ткаченко",
		FirstName:  "Любов",
		MiddleName: "Сергійвна",
		Gender:     "female",
	}

	sender := &User{
		ID:       int64(userId),
		Username: "testUser3",
	}

	// 1. Login as user
	loginUser(t, userId, fakeUser, sender)
	defer logoutUser(userId)

	sendMessageResponse := getSendMessageSuccessResponse()
	scoreMessageId := sendMessageLastId
	expectNewScoreMessageScope := mocks.TelegramMockServer.mocha.AddMocks(
		mocha.Post(expect.URLPath("/sendMessage")).
			Body(
				expectMarkdownV2, expectChatId(userId),
				expect.JSONPath("text", expect.ToContain(" запис: ")),
				expect.JSONPath("text", expect.ToContain(" заняття ")),
			).
			Reply(sendMessageResponse).
			Repeat(1),
	)
	defer expectNewScoreMessageScope.Clean()

	fmt.Println("scoreMessageId", scoreMessageId)

	catchMessage := &CatchMessagePostAction{}
	expectEditScoreMessageScope := mocks.TelegramMockServer.mocha.AddMocks(
		mocha.Post(expect.URLPath("/editMessageText")).
			Body(expectMarkdownV2, expectChatId(userId)).
			Reply(sendMessageResponse).
			Repeat(1).
			PostAction(catchMessage),
	)
	defer expectEditScoreMessageScope.Clean()

	// 2. push new records into the secondary
	expectDisciplineName := "Нейрокомпʼютерні системи"
	expectDisciplineId := 198568
	lessonDate := time.Date(2023, 7, 6, 0, 0, 0, 0, time.UTC)
	score1Value := 3
	score2Value := 4
	// 3. Update the database timestamp and wait X seconds
	createScoresForTest3(t, secondaryDekanatDb, fakeUser, expectDisciplineId, lessonDate, score1Value, score2Value)

	startTime := time.Now()
	waitUntilCalled(expectNewScoreMessageScope, 20*time.Second)
	actualWaitingTime := time.Since(startTime)

	if !expectNewScoreMessageScope.AssertCalled(t) {
		return
	}
	fmt.Println("Receive new score message in ", actualWaitingTime)

	startTime = time.Now()
	waitUntilCalled(expectEditScoreMessageScope, 15*time.Second)
	actualWaitingTime = time.Since(startTime)
	if !expectEditScoreMessageScope.AssertCalled(t) {
		return
	}
	fmt.Println("Receive edited score message in ", actualWaitingTime)

	fmt.Println("catchMessage", catchMessage)

	expectedText := fmt.Sprintf(
		"Новий запис: %s, заняття %s _Зан.в дистанц.реж._: %d та %d",
		expectDisciplineName, lessonDate.Format("02.01.2006"), score1Value, score2Value,
	)
	assert.Equal(t, expectedText, catchMessage.Text)
	assert.Equal(t, strconv.Itoa(scoreMessageId), catchMessage.MessageId)

	disciplineButton := catchMessage.GetInlineButton(0)
	if !assert.NotNil(t, disciplineButton) {
		return
	}

	assert.Contains(t, disciplineButton.Data, strconv.Itoa(expectDisciplineId))
	assert.Equal(t, expectDisciplineName, disciplineButton.Text)
}

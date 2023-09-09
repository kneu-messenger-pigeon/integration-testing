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
) (score1Id int, score2Id int) {
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

	score1Id = AddScore(t, secondaryDekanatDb, score1)

	score2 := Score{
		Lesson:     lesson,
		StudentId:  fakeUser.StudentId,
		LessonPart: 2,
		Score:      score2Value,
	}

	score2Id = AddScore(t, secondaryDekanatDb, score2)

	fmt.Printf("Create lesson %d with two scores: %d and %d\n", lesson.LessonId, score1Id, score2Id)

	UpdateDbDatetimeAndWait(t, secondaryDekanatDb, lessonDate)

	return
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

	UpdateDbDatetime(t, secondaryDekanatDb, time.Date(2023, 7, 6, 0, 0, 0, 0, time.UTC))

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

	catchMessage := &CatchMessagePostAction{}
	expectNewScoreMessageScope := mocks.TelegramMockServer.mocha.AddMocks(
		mocha.Post(expect.URLPath("/sendMessage")).
			Body(
				expectMarkdownV2, expectChatId(userId),
				expect.JSONPath("text", expect.ToContain(" запис: ")),
				expect.JSONPath("text", expect.ToContain(" заняття ")),
			).
			Reply(getSendMessageSuccessResponse()).
			Repeat(1).
			PostAction(catchMessage),
	)
	scoreMessageId := sendMessageLastId
	defer expectNewScoreMessageScope.Clean()

	fmt.Println("scoreMessageId", scoreMessageId)

	// it's optional call. Correct expected notification message could be received with `sendMessage` or `editMessageText`
	expectEditScoreMessageScope := mocks.TelegramMockServer.mocha.AddMocks(
		mocha.Post(expect.URLPath("/editMessageText")).
			Body(expectMarkdownV2, expectChatId(userId)).
			Reply(getEditMessageSuccessResponse()).
			PostAction(catchMessage),
	)
	defer expectEditScoreMessageScope.Clean()

	// 2. push new records into the secondary
	expectDisciplineName := "Нейрокомпʼютерні системи"
	expectDisciplineId := 198568
	lessonDate := time.Date(2023, 7, 8, 0, 0, 0, 0, time.UTC)
	score1Value := 3
	score2Value := 4
	// 3. Update the database timestamp and wait X seconds
	score1Id, score2Id := createScoresForTest3(
		t, secondaryDekanatDb, fakeUser, expectDisciplineId,
		lessonDate, score1Value, score2Value,
	)

	expectedText := fmt.Sprintf(
		"Новий запис: %s, заняття %s _Зан.в дистанц.реж._: %d та %d",
		expectDisciplineName, lessonDate.Format("02.01.2006"), score1Value, score2Value,
	)

	startTime := time.Now()
	waitUntilCalled(expectNewScoreMessageScope, 15*time.Second)
	actualWaitingTime := time.Since(startTime)

	assert.Equal(t, 1, expectNewScoreMessageScope.Hits())
	if !expectNewScoreMessageScope.AssertCalled(t) {
		return
	}
	expectNewScoreMessageScope.Clean()
	fmt.Println("Receive new score message in ", actualWaitingTime)

	if catchMessage.Text == expectedText && expectEditScoreMessageScope.IsPending() {
		fmt.Println("Receive send message with expected text")
	} else {
		// if in first iteration we not receive expected message, we should wait for edit message
		fmt.Println("wait for edit message: ", catchMessage.Text)
		catchMessage.Reset()

		startTime = time.Now()
		waitUntilCalled(expectEditScoreMessageScope, 5*time.Second)
		actualWaitingTime = time.Since(startTime)
		if !assert.Equal(t, 1, expectEditScoreMessageScope.Hits()) {
			return
		}
		fmt.Println("Receive edited score message in ", actualWaitingTime)
	}

	assert.Equal(t, expectedText, catchMessage.Text)
	assert.Equal(t, strconv.Itoa(scoreMessageId), catchMessage.MessageId)

	disciplineButton := catchMessage.GetInlineButton(0)
	if !assert.NotNil(t, disciplineButton) {
		return
	}

	assert.Contains(t, disciplineButton.Data, strconv.Itoa(expectDisciplineId))
	assert.Equal(t, expectDisciplineName, disciplineButton.Text)

	fmt.Println("")
	fmt.Println("Change score and expect change message with changes score")
	// 5. Change scores in the secondary DB
	catchMessage.Reset()
	newRegTime := lessonDate.Add(time.Minute * 30)
	editedScore1Value := 5
	UpdateScore(t, secondaryDekanatDb, score1Id, editedScore1Value, false, newRegTime)
	// 6. Update the database timestamp and wait X seconds
	UpdateDbDatetimeAndWait(t, secondaryDekanatDb, newRegTime)

	fmt.Println("score1Id, score2Id", score1Id, score2Id)

	// 7. Expect an edit message.
	startTime = time.Now()
	waitUntilCalledTimes(expectEditScoreMessageScope, 10*time.Second, 2)
	actualWaitingTime = time.Since(startTime)

	expectEditScoreMessageScope.AssertCalled(t)
	if !assert.Equal(t, 2, expectEditScoreMessageScope.Hits()) {
		t.FailNow()
	}
	if !expectEditScoreMessageScope.AssertCalled(t) {
		t.FailNow()
	}
	expectEditScoreMessageScope.Clean()

	fmt.Println("Receive edited score message in ", actualWaitingTime)

	assert.NotNil(t, catchMessage.GetInlineButton(0))
	expectedText = fmt.Sprintf(
		"Новий запис: %s, заняття %s _Зан.в дистанц.реж._: %d та %d",
		expectDisciplineName, lessonDate.Format("02.01.2006"), editedScore1Value, score2Value,
	)
	assert.Equal(t, expectedText, catchMessage.Text)

	// 8. Delete records
	expectDeleteMessageScope := mocks.TelegramMockServer.mocha.AddMocks(
		mocha.Post(expect.URLPath("/deleteMessage")).
			Body(expectChatId(userId), expectMessageId(scoreMessageId)).
			Reply(getDeleteMessageSuccessResponse()),
	)
	defer expectDeleteMessageScope.Clean()

	newRegTime = newRegTime.Add(time.Minute * 30)
	DeleteScore(t, secondaryDekanatDb, score1Id, newRegTime)
	DeleteScore(t, secondaryDekanatDb, score2Id, newRegTime)
	// 9. Update the database timestamp and wait X seconds
	UpdateDbDatetimeAndWait(t, secondaryDekanatDb, newRegTime)

	// 10. Expect to delete the message.
	waitUntilCalled(expectDeleteMessageScope, 15*time.Second)
	expectDeleteMessageScope.AssertCalled(t)
	expectDeleteMessageScope.Clean()
}
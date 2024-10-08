package main

import (
	"fmt"
	_ "github.com/nakagami/firebirdsql"
	"github.com/stretchr/testify/assert"
	"github.com/vitorsalgado/mocha/v3"
	"github.com/vitorsalgado/mocha/v3/expect"
	"github.com/vitorsalgado/mocha/v3/reply"
	"strconv"
	"testing"
	"time"
)

func createScoresForTest3(
	t *testing.T, fakeUser *FakeUser,
	disciplineId int, lessonDate time.Time,
	score1Value int, score2Value int,
	dbUpdateTime time.Time,
) (score1Id int, score2Id int) {
	lesson := &Lesson{
		GroupId:      fakeUser.GroupId,
		DisciplineId: disciplineId,
		Semester:     2,
		LessonTypeId: 15,
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

	score2 := &Score{
		Lesson:     lesson,
		StudentId:  fakeUser.StudentId,
		LessonPart: 2,
		Score:      score2Value,
	}

	score2Id = AddScore(t, mocks.SecondaryDB, score2)

	fmt.Printf("Create lesson %d with two scores: %d and %d\n", lesson.LessonId, score1Id, score2Id)

	UpdateDbDatetimeAndWait(t, mocks.SecondaryDB, dbUpdateTime)

	return
}

func Test3SecondaryDatabaseUpdates(t *testing.T) {
	fmt.Println("➡️Test3SecondaryDatabaseUpdates")
	defer printTestResult(t, "Test3SecondaryDatabaseUpdates")

	if config.repeatScoreChangesTimeframeSeconds < time.Second*30 {
		fmt.Println("Warning! Too small repeatScoreChangesTimeframeSeconds < 30 seconds")
	}

	firstDbUpdateTime := time.Date(2023, 7, 6, 0, 0, 0, 0, time.UTC)
	secondDbUpdateTime := time.Date(2023, 7, 6, 6, 0, 0, 0, time.UTC)
	thirdDbUpdateTime := time.Date(2023, 7, 6, 12, 0, 0, 0, time.UTC)
	fourthDbUpdateTime := time.Date(2023, 7, 6, 18, 0, 0, 0, time.UTC)

	UpdateDbDatetime(t, mocks.SecondaryDB, firstDbUpdateTime)

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
			Body(expectMarkdownV2, expectChatId(userId)).
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
			Reply(
				reply.Seq().
					Add(getEditMessageSuccessResponse(), getEditMessageSuccessResponse()),
			).
			PostAction(catchMessage),
	)
	defer expectEditScoreMessageScope.Clean()

	// 2. push new records into the secondary
	expectDisciplineName := "Нейрокомпʼютерні системи"
	expectDisciplineId := 198568
	lessonDate := secondDbUpdateTime
	lessonDate.Add(-10 * time.Minute)
	score1Value := 3
	score2Value := 4
	// 3. Update the database timestamp and wait X seconds
	score1Id, score2Id := createScoresForTest3(t, fakeUser, expectDisciplineId, lessonDate, score1Value, score2Value, secondDbUpdateTime)

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

	if catchMessage.Text == expectedText {
		fmt.Println("Receive send message with expected text")
	} else {
		// if in first iteration we not receive expected message, we should wait for edit message
		fmt.Println("wait for edit message: ", catchMessage.Text)
		catchMessage.Reset()

		startTime = time.Now()
		waitUntilCalled(expectEditScoreMessageScope, 10*time.Second)
		actualWaitingTime = time.Since(startTime)
		if !assert.Equal(t, 1, expectEditScoreMessageScope.Hits()) {
			return
		}

		// assert that edit same message as send
		assert.Equal(t, strconv.Itoa(scoreMessageId), catchMessage.MessageId)

		fmt.Println("Receive message ", catchMessage.Text)
		fmt.Println("Receive edited score message in ", actualWaitingTime)
	}

	assert.Equal(t, expectedText, catchMessage.Text)

	disciplineButton := catchMessage.GetInlineButton(0)
	if !assert.NotNil(t, disciplineButton) {
		return
	}

	assert.Contains(t, disciplineButton.Data, strconv.Itoa(expectDisciplineId))
	assert.Equal(t, expectDisciplineName, disciplineButton.Text)

	fmt.Println("")
	fmt.Println("Change score and expect change message with changes score")
	// reset called amount
	for expectEditScoreMessageScope.ListAll()[0].Hits() > 0 {
		expectEditScoreMessageScope.ListAll()[0].Dec()
	}

	// 5. Change scores in the secondary DB
	catchMessage.Reset()
	newRegTime := thirdDbUpdateTime
	newRegTime.Add(-10 * time.Minute)
	editedScore1Value := 5
	UpdateScore(t, mocks.SecondaryDB, score1Id, editedScore1Value, false, newRegTime)
	// 6. Update the database timestamp and wait X seconds
	UpdateDbDatetimeAndWait(t, mocks.SecondaryDB, thirdDbUpdateTime)

	fmt.Println("score1Id, score2Id", score1Id, score2Id)

	// 7. Expect an edit message.
	startTime = time.Now()
	waitUntilCalledTimes(expectEditScoreMessageScope, 10*time.Second, 2)
	actualWaitingTime = time.Since(startTime)

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

	newRegTime = fourthDbUpdateTime
	newRegTime.Add(-10 * time.Minute)
	DeleteScore(t, mocks.SecondaryDB, score1Id, newRegTime)
	DeleteScore(t, mocks.SecondaryDB, score2Id, newRegTime)

	// 9. Update the database timestamp and wait X seconds
	UpdateDbDatetimeAndWait(t, mocks.SecondaryDB, fourthDbUpdateTime)

	// 10. Expect to delete the message.
	waitUntilCalled(expectDeleteMessageScope, 15*time.Second)
	expectDeleteMessageScope.AssertCalled(t)
	expectDeleteMessageScope.Clean()
}

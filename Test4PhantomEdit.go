package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/vitorsalgado/mocha/v3"
	"github.com/vitorsalgado/mocha/v3/expect"
	"testing"
	"time"
)

/*
student id  113508
lesson id 2706428
score 2 id 32788942
score 1 id 32788943
*/

func Test4PhantomEdit(t *testing.T) {
	fmt.Println("➡️Test4PhantomEdit")
	defer printTestResult(t, "Test4PhantomEdit")

	startRegDate := time.Date(2023, 7, 6, 6, 0, 0, 0, time.UTC)
	UpdateDbDatetime(t, mocks.SecondaryDB, startRegDate)

	userId := test4PhantomEditUserId
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
	defer logoutUser(userId)

	score1Id := 32441822
	score2Id := 32441823
	lessonId := 2672002

	initialScore1Value := GetScoreValue(t, mocks.SecondaryDB, score1Id)
	initialScore2Value := GetScoreValue(t, mocks.SecondaryDB, score2Id)

	fmt.Printf("Lesson id %d - initial score values: %d and %d \n", lessonId, initialScore1Value, initialScore2Value)

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
			Body(expectChatId(userId)). // no expectMarkdownV2, because could be error
			Reply(getEditMessageSuccessResponse()).
			PostAction(catchMessage),
	)
	defer expectEditScoreMessageScope.Clean()

	regDate := startRegDate.Add(time.Minute * 30)
	newScore1Value := initialScore1Value + 1
	newScore2Value := initialScore2Value + 2

	UpdateScore(t, mocks.SecondaryDB, score1Id, newScore1Value, false, regDate)
	UpdateScore(t, mocks.SecondaryDB, score2Id, newScore2Value, false, regDate)
	UpdateDbDatetimeAndWait(t, mocks.SecondaryDB, startRegDate.Add(time.Hour))

	expectedText := fmt.Sprintf(
		"Змінено запис: Фахова іноземна мова, заняття 28.04.2023 _Зан.в дистанц.реж._: %d та %d (було ~%d та %d~)",
		newScore1Value, newScore2Value, initialScore1Value, initialScore2Value,
	)

	waitUntilCalled(expectNewScoreMessageScope, 10*time.Second)
	expectNewScoreMessageScope.AssertCalled(t)
	expectNewScoreMessageScope.Clean()
	if expectedText == catchMessage.Text {
		fmt.Println("Receive expected new score message")
	} else {
		fmt.Println("Receive intermediate new score message. Wait for expected edit score message..")
		waitUntilCalled(expectEditScoreMessageScope, 10*time.Second)
		expectEditScoreMessageScope.AssertCalled(t)
	}

	fmt.Println("catchMessage.Text", catchMessage.Text)
	if !assert.Equal(t, expectedText, catchMessage.Text) {
		return
	}

	// revert back changes and expect for delete message
	expectDeleteMessageScope := mocks.TelegramMockServer.mocha.AddMocks(
		mocha.Post(expect.URLPath("/deleteMessage")).
			Body(expectChatId(userId), expectMessageId(scoreMessageId)).
			Reply(getDeleteMessageSuccessResponse()),
	)
	defer expectDeleteMessageScope.Clean()

	regDate = startRegDate.Add(time.Minute * 90)
	UpdateScore(t, mocks.SecondaryDB, score1Id, initialScore1Value, false, regDate)
	UpdateScore(t, mocks.SecondaryDB, score2Id, initialScore2Value, false, regDate)
	UpdateDbDatetimeAndWait(t, mocks.SecondaryDB, startRegDate.Add(time.Hour*2))

	waitUntilCalled(expectDeleteMessageScope, 10*time.Second)
	expectDeleteMessageScope.AssertCalled(t)
}

/*

INSERT INTO T_EV_9(ID, ID_OBJ, REGDATE, XI_1, XI_2, XI_3, XI_4, XI_5, XR_1, XR_2, XR_3, XR_4, XR_5, XS10_1, XS10_2,
                   XS10_3, XS10_4, XS10_5, XS100_1, XS100_2, XS100_3, XS100_4, XS100_5, XRS_1, XRS_2, XRS_3, XD_1, XD_2,
                   ID_GRP, ID_T_PD_CMS, ID_USER, ID_CG, ID_ZANCG)
VALUES (32441822, 113508, '2023-04-29 12:28:58.0', null, 2672002, null, 1, 2, 4.0, null, null, null, null, null, null,
        '6.075¤4370     ', null, 'Так', null, null, null, null, null, '1779', '15', '1441', '2023-04-28 00:00:00.0',
        null, 16880, 198572, 1212, null, null);

INSERT INTO T_EV_9(ID, ID_OBJ, REGDATE, XI_1, XI_2, XI_3, XI_4, XI_5, XR_1, XR_2, XR_3, XR_4, XR_5, XS10_1, XS10_2,
                   XS10_3, XS10_4, XS10_5, XS100_1, XS100_2, XS100_3, XS100_4, XS100_5, XRS_1, XRS_2, XRS_3, XD_1, XD_2,
                   ID_GRP, ID_T_PD_CMS, ID_USER, ID_CG, ID_ZANCG)
VALUES (32441823, 113508, '2023-04-29 12:28:58.0', null, 2672002, null, 2, 2, 4.0, null, null, null, null, null, null,
        '6.075¤4370     ', null, 'Так', null, null, null, null, null, '1779', '15', '1441', '2023-04-28 00:00:00.0',
        null, 16880, 198572, 1212, null, null);
*/

package main

import (
	"context"
	"encoding/json"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
	"time"
)

const DekanatFormDateFormat = "02.01.2006"

type Form map[string]string

type RealtimeQueue struct {
	client      *sqs.Client
	sqsQueueUrl *string
	t           *testing.T
}

type EventMessage struct {
	ReceiptHandle *string
	Timestamp     int64  `json:"timestamp"`
	Ip            string `json:"ip"`
	Referer       string `json:"referer"`
	Form          *Form  `json:"form"`
}

const EventMessageJSON = `{
	"timestamp": %d,
	"ip": "127.0.0.1",
	"referer": "http://example.com",
	"form": %s
}`

func CreateRealtimeQueue(t *testing.T) *RealtimeQueue {
	keyPairMapping := [2][2]string{
		{"AWS_ACCESS_KEY_ID", "PRODUCER_AWS_ACCESS_KEY_ID"},
		{"AWS_SECRET_ACCESS_KEY", "PRODUCER_AWS_SECRET_ACCESS_KEY"},
	}
	backupsValues := [len(keyPairMapping)]string{}
	for index, keyPair := range keyPairMapping {
		backupsValues[index] = os.Getenv(keyPair[0])
		_ = os.Setenv(keyPair[0], os.Getenv(keyPair[1]))
	}

	// load config with overridden env vars
	awsCfg, err := awsConfig.LoadDefaultConfig(context.Background())
	for index, keyPair := range keyPairMapping {
		_ = os.Setenv(keyPair[0], backupsValues[index])
	}

	assert.NoError(t, err, "awsConfig.LoadDefaultConfig(context.Background()) failed")

	client := sqs.NewFromConfig(awsCfg)

	return &RealtimeQueue{
		client:      client,
		sqsQueueUrl: &config.sqsQueueUrl,
		t:           t,
	}
}

func (queue *RealtimeQueue) sendForm(form *Form) {
	message := EventMessage{
		Timestamp: time.Now().Unix(),
		Ip:        "127.0.0.1",
		Referer:   "http://example.com",
		Form:      form,
	}

	messageBody, err := json.Marshal(message)
	messageBodyString := string(messageBody)

	sendResult, err := queue.client.SendMessage(context.Background(), &sqs.SendMessageInput{
		QueueUrl:    queue.sqsQueueUrl,
		MessageBody: &messageBodyString,
	})

	assert.NoError(queue.t, err, "queue.client.SendMessage(context.Background(), &sqs.SendMessageInput{...}) failed")
	queue.t.Log("send to realtime queue message with id: " + *sendResult.MessageId)
}

func (queue *RealtimeQueue) SendLessonCreateEvent(lesson *Lesson) {
	queue.sendForm(&Form{
		"hlf":     strconv.Itoa(lesson.Semester),
		"prt":     strconv.Itoa(lesson.DisciplineId),
		"prti":    "0",
		"teacher": strconv.Itoa(lesson.TeachId),
		"action":  "insert",
		"n":       "10",
		"sesID":   "00AB0000-0000-0000-0000-000CD0000AA0",
		"m":       "-1",
		"date_z":  lesson.LessonDate.Format(DekanatFormDateFormat),
		"tzn":     strconv.Itoa(lesson.LessonTypeId),
		"result":  "3",
		"grade":   "",
	})
}

func (queue *RealtimeQueue) SendLessonEditEvent(lesson *Lesson) {
	queue.sendForm(&Form{
		"hlf":     strconv.Itoa(lesson.Semester),
		"prt":     strconv.Itoa(lesson.DisciplineId),
		"prti":    strconv.Itoa(lesson.LessonId),
		"teacher": strconv.Itoa(lesson.TeachId),
		"action":  "edit",
		"n":       "10",
		"sesID":   "00AB0000-0000-0000-0000-000CD0000AA0",
		"m":       "-1",
		"date_z":  lesson.LessonDate.Format(DekanatFormDateFormat),
		"tzn":     strconv.Itoa(lesson.LessonTypeId),
		"result":  "",
		"grade":   "2",
	})
}

func (queue *RealtimeQueue) SendLessonDeletedEvent(lesson *Lesson) {
	queue.sendForm(&Form{
		"sesID":  "00AB0000-0000-0000-0000-000CD0000AA0",
		"n":      "11",
		"action": "delete",
		"prti":   strconv.Itoa(lesson.LessonId),
		"prt":    strconv.Itoa(lesson.DisciplineId),
		"d1":     "",
		"d2":     "",
		"m":      "-1",
		"hlf":    strconv.Itoa(lesson.Semester),
		"course": "undefined",
	})
}

func (queue *RealtimeQueue) SendScoreEditEvent(lesson *Lesson, scores []*Score) {
	form := Form{
		"sesID":    "00AB0000-0000-0000-0000-000CD0000AA0",
		"n":        "4",
		"action":   "",
		"prti":     strconv.Itoa(lesson.LessonId),
		"prt":      strconv.Itoa(lesson.DisciplineId),
		"d1":       lesson.LessonDate.Format(DekanatFormDateFormat),
		"d2":       lesson.LessonDate.Format(DekanatFormDateFormat),
		"m":        "-1",
		"hlf":      strconv.Itoa(lesson.Semester),
		"course":   "3",
		"AddEstim": "0",
	}

	var scoreValue string
	for _, score := range scores {
		if score.IsAbsent {
			scoreValue = "нб/нп"
		} else {
			scoreValue = strconv.Itoa(score.Score)
		}
		form["st"+strconv.Itoa(score.StudentId)+"_"+strconv.Itoa(score.LessonPart)+"-999999"] = scoreValue
	}

	queue.sendForm(&form)
}

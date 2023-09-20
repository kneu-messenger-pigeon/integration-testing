package main

import (
	"context"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	dekanatEvents "github.com/kneu-messenger-pigeon/dekanat-events"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
	"time"
)

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

func (queue *RealtimeQueue) sendMessage(message *dekanatEvents.Message) {
	message.Timestamp = time.Now().Unix()
	message.Ip = "127.0.0.1"
	message.Referer = "http://example.com"

	sendResult, err := queue.client.SendMessage(context.Background(), &sqs.SendMessageInput{
		QueueUrl:    queue.sqsQueueUrl,
		MessageBody: message.ToJson(),
	})

	assert.NoError(queue.t, err, "queue.client.SendMessage(context.Background(), &sqs.SendMessageInput{...}) failed")
	queue.t.Log("send to realtime queue message with id: " + *sendResult.MessageId)
}

func (queue *RealtimeQueue) SendLessonCreateEvent(lesson *Lesson) {
	event := dekanatEvents.LessonCreateEvent{
		CommonEventData: dekanatEvents.CommonEventData{
			LessonId:     "0",
			DisciplineId: strconv.Itoa(lesson.DisciplineId),
			Semester:     strconv.Itoa(lesson.Semester),
		},
		TypeId:    strconv.Itoa(lesson.LessonTypeId),
		Date:      lesson.LessonDate.Format(dekanatEvents.DekanatFormDateFormat),
		TeacherId: strconv.Itoa(lesson.TeachId),
	}
	queue.sendMessage(event.ToMessage())
}

func (queue *RealtimeQueue) SendLessonEditEvent(lesson *Lesson) {
	event := dekanatEvents.LessonEditEvent{
		CommonEventData: dekanatEvents.CommonEventData{
			LessonId:     strconv.Itoa(lesson.LessonId),
			DisciplineId: strconv.Itoa(lesson.DisciplineId),
			Semester:     strconv.Itoa(lesson.Semester),
		},
		TypeId:    strconv.Itoa(lesson.LessonTypeId),
		TeacherId: strconv.Itoa(lesson.TeachId),
		Date:      lesson.LessonDate.Format(dekanatEvents.DekanatFormDateFormat),
	}
	queue.sendMessage(event.ToMessage())
}

func (queue *RealtimeQueue) SendLessonDeletedEvent(lesson *Lesson) {
	event := dekanatEvents.LessonDeletedEvent{
		CommonEventData: dekanatEvents.CommonEventData{
			LessonId:     strconv.Itoa(lesson.LessonId),
			DisciplineId: strconv.Itoa(lesson.DisciplineId),
			Semester:     strconv.Itoa(lesson.Semester),
		},
	}
	queue.sendMessage(event.ToMessage())
}

func (queue *RealtimeQueue) SendScoreEditEvent(lesson *Lesson, scores []*Score) {
	event := dekanatEvents.ScoreEditEvent{
		CommonEventData: dekanatEvents.CommonEventData{
			LessonId:     strconv.Itoa(lesson.LessonId),
			DisciplineId: strconv.Itoa(lesson.DisciplineId),
			Semester:     strconv.Itoa(lesson.Semester),
		},
		Date:   lesson.LessonDate.Format(dekanatEvents.DekanatFormDateFormat),
		Scores: map[int]map[uint8]string{},
	}

	var scoreValue string
	var hasKey bool
	for _, score := range scores {
		if score.IsAbsent {
			scoreValue = "нб/нп"
		} else {
			scoreValue = strconv.Itoa(score.Score)
		}

		if _, hasKey = event.Scores[score.StudentId]; !hasKey {
			event.Scores[score.StudentId] = make(map[uint8]string)
		}
		event.Scores[score.StudentId][score.LessonPart] = scoreValue
	}
	queue.sendMessage(event.ToMessage())
}

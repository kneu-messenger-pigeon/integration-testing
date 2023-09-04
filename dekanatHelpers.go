package main

import (
	"database/sql"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

const dateFormat = "2006-01-02 15:04:05"

type Lesson struct {
	LessonId     int
	DisciplineId int
	GroupId      int
	Semester     int
	LessonTypeId int
	LessonDate   time.Time
	TeachId      int
	TeachUserId  int
}

type Score struct {
	Lesson     *Lesson
	StudentId  int
	LessonPart int
	Score      int
	IsAbsent   bool
}

const insertLessonQuery = `INSERT INTO T_PRJURN(
		 GRP_ID, NUM_PREDM, NUM_VARZAN, FSTATUS,
		 REGDATE, DATEZAN, REMARK, 
		 NUMWEEK, NUMDAYOFWEEK, NUMPARY,
		 ID_TEACH, ID_USER, HALF, BLOCKED,
		 ID_CG, ID_ZANCG, NUM_MOD, REMARK_IS_PUBL, PERIOD, COURSE
	 ) VALUES (
	   ?, ?, ?, 1,
	   ?, ?, null,
	   null, null, null,
	   ?, ?, ?, 1, 
	   null, null, 1, 0, 2, 3
	) RETURNING ID;`

/*
*

		   ID_OBJ AS STUDENT_ID,
		   XI_2 AS LESSON_ID,
		   XI_4 as LESSON_PART,
		   ID_T_PD_CMS AS DISCIPLINE_ID,
		   XI_5 as SEMESTER,
		   COALESCE(XR_1, 0) AS SCORE,
	       IIF(XS10_4 IS NULL, 0, 1) AS IS_ABSENT,
	       REGDATE,
		   IS_DELETED = XS10_5 != 'Так' OR (XR_1 IS NULL AND XS10_4 IS NULL)

	      XRS_1 = NUM_PREDM
		  XRS_3 AS TEACHER_ID,
	      XD_1 AS REGDATE
*/
const insertScoreQuery = `INSERT INTO T_EV_9(
	   ID_OBJ, REGDATE, XD_1,
	   XI_2, XI_4, ID_GRP, ID_T_PD_CMS,
	   XI_5, XR_1, XS10_4, XS10_5,
       XRS_1, XRS_2, XRS_3, 
	   XI_1, XI_3,  XR_2, XR_3, XR_4, XR_5, XS10_1, XS10_2,
	   XS10_3,  XS100_1, XS100_2, XS100_3, XS100_4, XS100_5, 
	   XD_2, ID_USER, ID_CG, ID_ZANCG
   ) VALUES (
	   ?, ?, ?,
	   ?, ?, ?, ?,
	   ?, ?, ?, 'Так', 
	  '9999', ?, ?, 
	   null,  null, null, null, null, null, null, null,
	  '6.075¤4370     ', null, null, null, null, null,
	  null, ?, null, null
   ) RETURNING ID;`

func AddLesson(t *testing.T, db *sql.DB, lesson Lesson) int {
	var idRow *sql.Row
	var err error

	lessonDate := lesson.LessonDate.Format(dateFormat)

	// create new lesson
	idRow = db.QueryRow(
		insertLessonQuery,
		lesson.GroupId, lesson.DisciplineId, lesson.LessonTypeId,
		lessonDate, lessonDate,
		lesson.TeachId, lesson.TeachUserId, lesson.Semester,
	)
	if !assert.NoError(t, idRow.Err()) {
		t.FailNow()
	}

	err = idRow.Scan(&lesson.LessonId)
	assert.NoError(t, err)
	fmt.Println("insert lesson with id " + strconv.Itoa(lesson.LessonId))

	return lesson.LessonId
}

func AddScore(t *testing.T, db *sql.DB, score Score) int {
	var idRow *sql.Row
	var err error

	lesson := score.Lesson
	lessonDate := lesson.LessonDate.Format(dateFormat)

	// create new score
	var isAbsent *bool
	var scoreValue *int
	if score.IsAbsent {
		isAbsent = &score.IsAbsent
	} else {
		scoreValue = &score.Score
	}

	idRow = db.QueryRow(
		insertScoreQuery,
		score.StudentId, lessonDate, lessonDate,
		lesson.LessonId, score.LessonPart, lesson.GroupId, lesson.DisciplineId,
		lesson.Semester,

		scoreValue, isAbsent,
		strconv.Itoa(lesson.LessonTypeId), strconv.Itoa(lesson.TeachId),
		lesson.TeachUserId,
	)
	if !assert.NoError(t, idRow.Err()) {
		t.FailNow()
	}

	var scoreId int
	err = idRow.Scan(&scoreId)
	assert.NoError(t, err)
	fmt.Println("insert score with id " + strconv.Itoa(scoreId))

	return scoreId
}

func UpdateDbDatetimeAndWait(t *testing.T, db *sql.DB, datetime time.Time) {
	randomizedDateTime := datetime.Add(time.Second*10 + time.Duration(rand.Intn(300))*time.Second)

	result, err := db.Exec("UPDATE TSESS_LOG SET CON_DATA = ?", randomizedDateTime.Format(dateFormat))
	assert.NoError(t, err)
	affected, err := result.RowsAffected()
	assert.NoError(t, err)
	assert.NotEmpty(t, affected)

	fmt.Printf(
		"Done inserting new scores. Wait secondary db update interval - %d seconds\n",
		int(config.secondaryDbCheckInterval.Seconds()),
	)
	time.Sleep(config.secondaryDbCheckInterval)
}

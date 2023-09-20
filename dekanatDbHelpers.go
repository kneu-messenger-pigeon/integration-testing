package main

import (
	"database/sql"
	"fmt"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

const dateFormat = "2006-01-02 15:04:05"

const absentStringValue = "нб/нп"

var absentString = absentStringValue

type Lesson struct {
	LessonId     int
	DisciplineId int
	GroupId      int
	Semester     int
	LessonTypeId int
	LessonDate   time.Time
	TeachId      int
	TeachUserId  int
	RegDate      time.Time
}

type Score struct {
	Lesson     *Lesson
	StudentId  int
	LessonPart uint8
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

	regDate := lesson.RegDate.Format(dateFormat)
	lessonDate := lesson.LessonDate.Format(dateFormat)

	// create new lesson
	idRow = db.QueryRow(
		insertLessonQuery,
		lesson.GroupId, lesson.DisciplineId, lesson.LessonTypeId,
		regDate, lessonDate,
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

func AddScore(t *testing.T, db *sql.DB, score *Score) int {
	var idRow *sql.Row
	var err error

	lesson := score.Lesson
	regDate := lesson.RegDate.Format(dateFormat)
	lessonDate := lesson.LessonDate.Format(dateFormat)

	// create new score
	var absentValue *string
	var scoreValue *int
	if score.IsAbsent {
		absentValue = &absentString
	} else {
		scoreValue = &score.Score
	}

	idRow = db.QueryRow(
		insertScoreQuery,
		score.StudentId, regDate, lessonDate,
		lesson.LessonId, score.LessonPart, lesson.GroupId, lesson.DisciplineId,
		lesson.Semester,

		scoreValue, absentValue,
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

func UpdateScore(t *testing.T, db *sql.DB, scoreId int, score int, isAbsent bool, datetime time.Time) {
	var err error
	var result sql.Result

	// create new score
	var argAbsent *string
	var argScoreValue *int
	if isAbsent {
		argAbsent = &absentString
	} else {
		argScoreValue = &score
	}

	result, err = db.Exec(
		"UPDATE T_EV_9 SET XR_1 = ?, XS10_4 = ?, REGDATE = ? WHERE ID = ?",
		argScoreValue, argAbsent,
		datetime.Format(dateFormat),
		scoreId,
	)
	assert.NoError(t, err)

	affected, err := result.RowsAffected()
	assert.NoError(t, err)
	assert.NotEmpty(t, affected)
}

func DeleteLesson(t *testing.T, db *sql.DB, lessonId int, datetime time.Time) {
	var err error
	var result sql.Result

	result, err = db.Exec(
		"UPDATE T_PRJURN SET FSTATUS = 0, REGDATE = ? WHERE ID = ?",
		datetime.Format(dateFormat), lessonId,
	)
	assert.NoError(t, err)

	affected, err := result.RowsAffected()
	assert.NoError(t, err)
	assert.NotEmpty(t, affected)
}

func DeleteScore(t *testing.T, db *sql.DB, scoreId int, datetime time.Time) {
	var err error
	var result sql.Result

	result, err = db.Exec(
		"UPDATE T_EV_9 SET XS10_5 = 'Ні', REGDATE = ? WHERE ID = ?",
		datetime.Format(dateFormat), scoreId,
	)
	assert.NoError(t, err)

	affected, err := result.RowsAffected()
	assert.NoError(t, err)
	assert.NotEmpty(t, affected)
}

func GetScoreValue(t *testing.T, db *sql.DB, scoreId int) int {
	var err error
	var row *sql.Row

	row = db.QueryRow(
		"SELECT XR_1 FROM T_EV_9 WHERE ID = ?",
		scoreId,
	)
	assert.NoError(t, err)

	var scoreValue int
	err = row.Scan(&scoreValue)
	assert.NoError(t, err)

	return scoreValue
}

func UpdateDbDatetime(t *testing.T, db *sql.DB, datetime time.Time) {
	result, err := db.Exec("UPDATE TSESS_LOG SET CON_DATA = ?", datetime.Format(dateFormat))
	assert.NoError(t, err)
	affected, err := result.RowsAffected()
	assert.NoError(t, err)
	assert.NotEmpty(t, affected)

}

func UpdateDbDatetimeAndWait(t *testing.T, db *sql.DB, datetime time.Time) {
	UpdateDbDatetime(t, db, datetime)
	fmt.Printf(
		"Done inserting new scores. Set DB time: %v, Wait secondary db update interval - %d seconds\n",
		datetime, int(config.secondaryDbCheckInterval.Seconds()),
	)
	time.Sleep(config.secondaryDbCheckInterval)
}

func SyncDbAutoIncrements() {
	tableNames := [2]string{
		"T_PRJURN",
		"T_EV_9",
	}

	for _, tableName := range tableNames {
		primaryDbMaxId := getMaxId(mocks.SecondaryDB, tableName)
		secondaryDbMaxId := getMaxId(mocks.SecondaryDB, tableName)
		if secondaryDbMaxId < primaryDbMaxId {
			setAutoIncrement(mocks.SecondaryDB, tableName, primaryDbMaxId+1)
		} else if primaryDbMaxId < secondaryDbMaxId {
			setAutoIncrement(mocks.PrimaryDB, tableName, secondaryDbMaxId+1)
		}
	}
}

func getMaxId(db *sql.DB, tableName string) int {
	row := db.QueryRow("SELECT MAX(ID) FROM " + tableName)
	maxId := 0
	err := row.Scan(&maxId)
	if err != nil {
		panic(err)
	}

	return maxId
}

func setAutoIncrement(db *sql.DB, tableName string, nextId int) {
	_, err := db.Exec("ALTER TABLE " + tableName + " ALTER COLUMN ID RESTART WITH " + strconv.Itoa(nextId))
	if err != nil {
		panic(err)
	}
}

func OpenDbConnection(t *testing.T, dsn string) *sql.DB {
	db, err := sql.Open("firebirdsql", dsn)
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}

	err = db.Ping()
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}

	return db
}

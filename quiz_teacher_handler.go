package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ============ CREATE QUIZ ============

type QuizInput struct {
	SubjectID   uint      `json:"subject_id" binding:"required"`
	Title       string    `json:"title" binding:"required"`
	QuizType    string    `json:"quiz_type" binding:"required"` // "kuis" atau "ujian"
	IsTimed     bool      `json:"is_timed"`
	DurationMin int       `json:"duration_min"`
	StartTime   time.Time `json:"start_time" binding:"required"`
	EndTime     time.Time `json:"end_time" binding:"required"`
}

func CreateQuiz(c *gin.Context) {
	roleID, _ := c.Get("role_id")
	if uint(roleID.(float64)) != 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak! Anda bukan Guru."})
		return
	}
	userID, _ := c.Get("user_id")

	var input QuizInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	quiz := Quiz{
		SubjectID:   input.SubjectID,
		TeacherID:   userID.(string),
		Title:       input.Title,
		QuizType:    input.QuizType,
		IsTimed:     input.IsTimed,
		DurationMin: input.DurationMin,
		StartTime:   input.StartTime,
		EndTime:     input.EndTime,
	}
	DB.Create(&quiz)

	c.JSON(http.StatusOK, gin.H{"message": "Kuis/Ujian berhasil dibuat", "data": quiz})
}

// ============ ADD QUESTION ============

type QuestionInput struct {
	QuestionText  string `json:"question_text" binding:"required"`
	QuestionType  string `json:"question_type" binding:"required"` // "pilihan_ganda" atau "esai"
	OptionA       string `json:"option_a"`
	OptionB       string `json:"option_b"`
	OptionC       string `json:"option_c"`
	OptionD       string `json:"option_d"`
	CorrectOption string `json:"correct_option"`
	Score         int    `json:"score"`
}

func AddQuestion(c *gin.Context) {
	roleID, _ := c.Get("role_id")
	if uint(roleID.(float64)) != 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak! Anda bukan Guru."})
		return
	}

	quizID := c.Param("quiz_id")

	var input QuestionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	score := input.Score
	if score == 0 {
		score = 10
	}

	question := Question{
		QuizID:        parseUint(quizID),
		QuestionText:  input.QuestionText,
		QuestionType:  input.QuestionType,
		OptionA:       input.OptionA,
		OptionB:       input.OptionB,
		OptionC:       input.OptionC,
		OptionD:       input.OptionD,
		CorrectOption: input.CorrectOption,
		Score:         score,
	}
	DB.Create(&question)

	c.JSON(http.StatusOK, gin.H{"message": "Soal berhasil ditambahkan", "data": question})
}

// ============ GET QUIZ LIST + DETAIL (untuk guru lihat kuis miliknya) ============

func GetMyQuizzes(c *gin.Context) {
	roleID, _ := c.Get("role_id")
	if uint(roleID.(float64)) != 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak! Anda bukan Guru."})
		return
	}
	userID, _ := c.Get("user_id")

	var quizzes []Quiz
	DB.Where("teacher_id = ?", userID.(string)).Preload("Subject").Find(&quizzes)

	c.JSON(http.StatusOK, gin.H{"data": quizzes})
}

func GetQuizQuestions(c *gin.Context) {
	roleID, _ := c.Get("role_id")
	if uint(roleID.(float64)) != 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak! Anda bukan Guru."})
		return
	}

	quizID := c.Param("quiz_id")
	var questions []Question
	DB.Where("quiz_id = ?", quizID).Find(&questions)

	c.JSON(http.StatusOK, gin.H{"data": questions})
}

// ============ LIHAT SUBMISSION SISWA ============

func GetQuizSubmissions(c *gin.Context) {
	roleID, _ := c.Get("role_id")
	if uint(roleID.(float64)) != 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak! Anda bukan Guru."})
		return
	}

	quizID := c.Param("quiz_id")
	var submissions []QuizSubmission
	DB.Where("quiz_id = ?", quizID).Preload("Student").Find(&submissions)

	c.JSON(http.StatusOK, gin.H{"data": submissions})
}

func GetSubmissionAnswers(c *gin.Context) {
	roleID, _ := c.Get("role_id")
	if uint(roleID.(float64)) != 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak! Anda bukan Guru."})
		return
	}

	submissionID := c.Param("submission_id")
	var answers []Answer
	DB.Where("quiz_submission_id = ?", submissionID).Preload("Question").Find(&answers)

	c.JSON(http.StatusOK, gin.H{"data": answers})
}

// ============ NILAI ESAI (manual) ============

type GradeEssayInput struct {
	Score          int    `json:"score" binding:"required"`
	TeacherComment string `json:"teacher_comment"`
}

func GradeEssayAnswer(c *gin.Context) {
	roleID, _ := c.Get("role_id")
	if uint(roleID.(float64)) != 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak! Anda bukan Guru."})
		return
	}

	answerID := c.Param("answer_id")
	var answer Answer
	if err := DB.First(&answer, answerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Jawaban tidak ditemukan"})
		return
	}

	var input GradeEssayInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	DB.Model(&answer).Updates(map[string]interface{}{
		"Score":          input.Score,
		"TeacherComment": input.TeacherComment,
	})

	var allAnswers []Answer
	DB.Where("quiz_submission_id = ?", answer.QuizSubmissionID).Find(&allAnswers)
	total := 0
	for _, a := range allAnswers {
		total += a.Score
	}
	DB.Model(&QuizSubmission{}).Where("id = ?", answer.QuizSubmissionID).Updates(map[string]interface{}{
		"TotalScore": total,
		"Status":     "sudah_dinilai",
	})

	c.JSON(http.StatusOK, gin.H{"message": "Nilai esai berhasil disimpan"})
}

// ============ UPDATE & DELETE QUIZ ============

func UpdateQuiz(c *gin.Context) {
	roleID, _ := c.Get("role_id")
	if uint(roleID.(float64)) != 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak! Anda bukan Guru."})
		return
	}
	userID, _ := c.Get("user_id")

	quizID := c.Param("quiz_id")
	var quiz Quiz
	if err := DB.Where("id = ? AND teacher_id = ?", quizID, userID.(string)).First(&quiz).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kuis tidak ditemukan atau bukan milik Anda"})
		return
	}

	var input QuizInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	DB.Model(&quiz).Updates(map[string]interface{}{
		"SubjectID":   input.SubjectID,
		"Title":       input.Title,
		"QuizType":    input.QuizType,
		"IsTimed":     input.IsTimed,
		"DurationMin": input.DurationMin,
		"StartTime":   input.StartTime,
		"EndTime":     input.EndTime,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Kuis berhasil diperbarui"})
}

func DeleteQuiz(c *gin.Context) {
	roleID, _ := c.Get("role_id")
	if uint(roleID.(float64)) != 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak! Anda bukan Guru."})
		return
	}
	userID, _ := c.Get("user_id")

	quizID := c.Param("quiz_id")
	var quiz Quiz
	if err := DB.Where("id = ? AND teacher_id = ?", quizID, userID.(string)).First(&quiz).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kuis tidak ditemukan atau bukan milik Anda"})
		return
	}

	DB.Where("quiz_id = ?", quizID).Delete(&Question{})
	DB.Delete(&quiz)

	c.JSON(http.StatusOK, gin.H{"message": "Kuis dan semua soalnya dihapus"})
}

// ============ UPDATE & DELETE QUESTION ============

func UpdateQuestion(c *gin.Context) {
	roleID, _ := c.Get("role_id")
	if uint(roleID.(float64)) != 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak! Anda bukan Guru."})
		return
	}

	questionID := c.Param("question_id")
	var question Question
	if err := DB.First(&question, questionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Soal tidak ditemukan"})
		return
	}

	var input QuestionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	DB.Model(&question).Updates(map[string]interface{}{
		"QuestionText":  input.QuestionText,
		"QuestionType":  input.QuestionType,
		"OptionA":       input.OptionA,
		"OptionB":       input.OptionB,
		"OptionC":       input.OptionC,
		"OptionD":       input.OptionD,
		"CorrectOption": input.CorrectOption,
		"Score":         input.Score,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Soal berhasil diperbarui"})
}

func DeleteQuestion(c *gin.Context) {
	roleID, _ := c.Get("role_id")
	if uint(roleID.(float64)) != 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak! Anda bukan Guru."})
		return
	}

	questionID := c.Param("question_id")
	DB.Delete(&Question{}, questionID)

	c.JSON(http.StatusOK, gin.H{"message": "Soal dihapus"})
}

// ============ HELPER ============

func parseUint(s string) uint {
	var result uint
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0
		}
		result = result*10 + uint(ch-'0')
	}
	return result
}
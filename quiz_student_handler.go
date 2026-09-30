package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ============ LIHAT KUIS YANG TERSEDIA ============

func GetAvailableQuizzes(c *gin.Context) {
	roleID, _ := c.Get("role_id")
	if uint(roleID.(float64)) != 3 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak! Anda bukan Siswa."})
		return
	}

	now := time.Now()
	var quizzes []Quiz
	DB.Where("start_time <= ? AND end_time >= ?", now, now).Preload("Subject").Find(&quizzes)

	c.JSON(http.StatusOK, gin.H{"data": quizzes})
}

func GetQuizForStudent(c *gin.Context) {
	roleID, _ := c.Get("role_id")
	if uint(roleID.(float64)) != 3 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak! Anda bukan Siswa."})
		return
	}

	quizID := c.Param("quiz_id")
	var questions []Question
	DB.Where("quiz_id = ?", quizID).Find(&questions)

	// Sembunyikan kunci jawaban dari siswa
	type SafeQuestion struct {
		ID           uint   `json:"id"`
		QuestionText string `json:"question_text"`
		QuestionType string `json:"question_type"`
		OptionA      string `json:"option_a"`
		OptionB      string `json:"option_b"`
		OptionC      string `json:"option_c"`
		OptionD      string `json:"option_d"`
		Score        int    `json:"score"`
	}

	safeList := []SafeQuestion{}
	for _, q := range questions {
		safeList = append(safeList, SafeQuestion{
			ID: q.ID, QuestionText: q.QuestionText, QuestionType: q.QuestionType,
			OptionA: q.OptionA, OptionB: q.OptionB, OptionC: q.OptionC, OptionD: q.OptionD,
			Score: q.Score,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": safeList})
}

// ============ MULAI KUIS ============

func StartQuiz(c *gin.Context) {
	roleID, _ := c.Get("role_id")
	if uint(roleID.(float64)) != 3 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak! Anda bukan Siswa."})
		return
	}
	userID, _ := c.Get("user_id")
	quizID := c.Param("quiz_id")

	// Cek apakah siswa sudah pernah mulai kuis ini
	var existing QuizSubmission
	if err := DB.Where("quiz_id = ? AND student_id = ?", quizID, userID.(string)).First(&existing).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "Kuis sudah pernah dimulai", "data": existing})
		return
	}

	submission := QuizSubmission{
		QuizID:    parseUint(quizID),
		StudentID: userID.(string),
		StartedAt: time.Now(),
		Status:    "mengerjakan",
	}
	DB.Create(&submission)

	c.JSON(http.StatusOK, gin.H{"message": "Kuis dimulai", "data": submission})
}

// ============ SUBMIT JAWABAN (dengan auto-grading PG) ============

type SubmitAnswerInput struct {
	QuestionID     uint   `json:"question_id" binding:"required"`
	SelectedOption string `json:"selected_option"`
	EssayAnswer    string `json:"essay_answer"`
}

type SubmitQuizInput struct {
	Answers []SubmitAnswerInput `json:"answers" binding:"required"`
}

func SubmitQuiz(c *gin.Context) {
	roleID, _ := c.Get("role_id")
	if uint(roleID.(float64)) != 3 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak! Anda bukan Siswa."})
		return
	}
	userID, _ := c.Get("user_id")
	submissionID := c.Param("submission_id")

	var submission QuizSubmission
	if err := DB.Where("id = ? AND student_id = ?", submissionID, userID.(string)).First(&submission).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Submission tidak ditemukan"})
		return
	}

	if submission.Status != "mengerjakan" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kuis sudah pernah disubmit"})
		return
	}

	var input SubmitQuizInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	totalScore := 0
	hasEssay := false

	for _, ans := range input.Answers {
		var question Question
		if err := DB.First(&question, ans.QuestionID).Error; err != nil {
			continue
		}

		answer := Answer{
			QuizSubmissionID: submission.ID,
			QuestionID:       ans.QuestionID,
			SelectedOption:   ans.SelectedOption,
			EssayAnswer:      ans.EssayAnswer,
		}

		if question.QuestionType == "pilihan_ganda" {
			// Auto-grading PG
			if ans.SelectedOption == question.CorrectOption {
				answer.Score = question.Score
			} else {
				answer.Score = 0
			}
			totalScore += answer.Score
		} else {
			// Esai nunggu dinilai manual guru
			hasEssay = true
			answer.Score = 0
		}

		DB.Create(&answer)
	}

	now := time.Now()
	status := "selesai"
	if !hasEssay {
		status = "sudah_dinilai" // kalau gak ada esai, PG doang udah final
	}

	DB.Model(&submission).Updates(map[string]interface{}{
		"SubmittedAt": &now,
		"TotalScore":  totalScore,
		"Status":      status,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Kuis berhasil disubmit", "total_score_sementara": totalScore})
}

// ============ LIHAT HASIL KUIS SENDIRI ============

func GetMyQuizResult(c *gin.Context) {
	roleID, _ := c.Get("role_id")
	if uint(roleID.(float64)) != 3 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak! Anda bukan Siswa."})
		return
	}
	userID, _ := c.Get("user_id")
	submissionID := c.Param("submission_id")

	var submission QuizSubmission
	if err := DB.Where("id = ? AND student_id = ?", submissionID, userID.(string)).Preload("Quiz").First(&submission).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": submission})
}
package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GenerateGrades(c *gin.Context) {
	roleID, _ := c.Get("role_id")
	if uint(roleID.(float64)) != 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak! Anda bukan Guru."})
		return
	}

	subjectID := c.Param("subject_id")

	// Ambil semua siswa (role_id 3) — bisa disaring lebih spesifik per kelas kalau perlu nanti
	var students []User
	DB.Where("role_id = 3").Find(&students)

	var results []Grade

	for _, student := range students {
		// Hitung rata-rata nilai tugas (dari Submission -> Assignment yang SubjectID cocok)
		var assignmentSubs []Submission
		DB.Joins("JOIN assignments ON assignments.id = submissions.assignment_id").
			Where("submissions.student_id = ? AND assignments.subject_id = ?", student.ID, subjectID).
			Find(&assignmentSubs)

		assignmentAvg := 0.0
		if len(assignmentSubs) > 0 {
			total := 0
			for _, s := range assignmentSubs {
				total += s.Score
			}
			assignmentAvg = float64(total) / float64(len(assignmentSubs))
		}

		// Hitung rata-rata nilai kuis/ujian (dari QuizSubmission -> Quiz yang SubjectID cocok)
		var quizSubs []QuizSubmission
		DB.Joins("JOIN quizzes ON quizzes.id = quiz_submissions.quiz_id").
			Where("quiz_submissions.student_id = ? AND quizzes.subject_id = ? AND quiz_submissions.status = 'sudah_dinilai'", student.ID, subjectID).
			Find(&quizSubs)

		quizAvg := 0.0
		if len(quizSubs) > 0 {
			total := 0
			for _, q := range quizSubs {
				total += q.TotalScore
			}
			quizAvg = float64(total) / float64(len(quizSubs))
		}

		// Skip siswa yang belum punya nilai tugas maupun kuis sama sekali
		if len(assignmentSubs) == 0 && len(quizSubs) == 0 {
			continue
		}

		finalScore := (assignmentAvg + quizAvg) / 2
		if len(assignmentSubs) == 0 {
			finalScore = quizAvg
		} else if len(quizSubs) == 0 {
			finalScore = assignmentAvg
		}

		var existing Grade
		err := DB.Where("student_id = ? AND subject_id = ?", student.ID, subjectID).First(&existing).Error

		if err == nil {
			// Update kalau sudah ada
			DB.Model(&existing).Updates(map[string]interface{}{
				"AssignmentAvg": assignmentAvg,
				"QuizAvg":       quizAvg,
				"FinalScore":    finalScore,
				"GeneratedAt":   time.Now(),
			})
			results = append(results, existing)
		} else {
			// Buat baru kalau belum ada
			grade := Grade{
				StudentID:     student.ID,
				SubjectID:     parseUint(subjectID),
				AssignmentAvg: assignmentAvg,
				QuizAvg:       quizAvg,
				FinalScore:    finalScore,
				GeneratedAt:   time.Now(),
			}
			DB.Create(&grade)
			results = append(results, grade)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Nilai berhasil digenerate",
		"data":    results,
	})
}

func GetGradesBySubject(c *gin.Context) {
	roleID, _ := c.Get("role_id")
	if uint(roleID.(float64)) != 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak! Anda bukan Guru."})
		return
	}

	subjectID := c.Param("subject_id")
	var grades []Grade
	DB.Where("subject_id = ?", subjectID).Preload("Student").Preload("Subject").Find(&grades)

	c.JSON(http.StatusOK, gin.H{"data": grades})
}

func GetMyGrades(c *gin.Context) {
	roleID, _ := c.Get("role_id")
	if uint(roleID.(float64)) != 3 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak! Anda bukan Siswa."})
		return
	}
	userID, _ := c.Get("user_id")

	var grades []Grade
	DB.Where("student_id = ?", userID.(string)).Preload("Subject").Find(&grades)

	c.JSON(http.StatusOK, gin.H{"data": grades})
}
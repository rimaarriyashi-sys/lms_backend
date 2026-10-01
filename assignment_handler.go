package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type AssignmentInput struct {
	SubjectID uint      `json:"subject_id" binding:"required"`
	Title     string    `json:"title" binding:"required"`
	Deadline  time.Time `json:"deadline" binding:"required"`
	MaxScore  int       `json:"max_score"`
}

func CreateAssignment(c *gin.Context) {
	roleID, _ := c.Get("role_id")
	if uint(roleID.(float64)) != 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak! Anda bukan Guru."})
		return
	}

	var input AssignmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	maxScore := input.MaxScore
	if maxScore == 0 {
		maxScore = 100
	}

	assignment := Assignment{
		SubjectID: input.SubjectID,
		Title:     input.Title,
		Deadline:  input.Deadline,
		MaxScore:  maxScore,
	}
	DB.Create(&assignment)

	c.JSON(http.StatusOK, gin.H{"message": "Tugas berhasil dibuat", "data": assignment})
}

func GetAssignments(c *gin.Context) {
	_, exists := c.Get("role_id")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak!"})
		return
	}

	var assignments []Assignment
	DB.Preload("Subject").Find(&assignments)

	c.JSON(http.StatusOK, gin.H{"data": assignments})
}

func UpdateAssignment(c *gin.Context) {
	roleID, _ := c.Get("role_id")
	if uint(roleID.(float64)) != 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak! Anda bukan Guru."})
		return
	}

	id := c.Param("id")
	var assignment Assignment
	if err := DB.First(&assignment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tugas tidak ditemukan"})
		return
	}

	var input AssignmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	DB.Model(&assignment).Updates(map[string]interface{}{
		"SubjectID": input.SubjectID,
		"Title":     input.Title,
		"Deadline":  input.Deadline,
		"MaxScore":  input.MaxScore,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Tugas berhasil diperbarui"})
}

func DeleteAssignment(c *gin.Context) {
	roleID, _ := c.Get("role_id")
	if uint(roleID.(float64)) != 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak! Anda bukan Guru."})
		return
	}

	id := c.Param("id")
	DB.Delete(&Assignment{}, id)

	c.JSON(http.StatusOK, gin.H{"message": "Tugas dihapus"})
}

type GradeAssignmentInput struct {
	Score    int    `json:"score" binding:"required"`
	Feedback string `json:"feedback"`
}

func GradeSubmission(c *gin.Context) {
	roleID, _ := c.Get("role_id")
	if uint(roleID.(float64)) != 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak! Anda bukan Guru."})
		return
	}

	submissionID := c.Param("submission_id")
	var submission Submission
	if err := DB.First(&submission, submissionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Submission tidak ditemukan"})
		return
	}

	var input GradeAssignmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	DB.Model(&submission).Updates(map[string]interface{}{
		"Score":    input.Score,
		"Feedback": input.Feedback,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Nilai tugas berhasil disimpan"})
}
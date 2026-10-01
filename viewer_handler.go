package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func isKurikulumOrKepsek(c *gin.Context) bool {
	roleID, exists := c.Get("role_id")
	if !exists || (uint(roleID.(float64)) != 4 && uint(roleID.(float64)) != 5) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak!"})
		c.Abort()
		return false
	}
	return true
}

// 1. Lihat nilai per mapel (semua siswa, semua mapel)
func ViewAllGrades(c *gin.Context) {
	if !isKurikulumOrKepsek(c) { return }

	var grades []Grade
	DB.Preload("Student").Preload("Subject").Find(&grades)

	c.JSON(http.StatusOK, gin.H{"data": grades})
}

// 2. Lihat tugas yang diberi guru (semua Assignment + info mapelnya)
func ViewAllAssignments(c *gin.Context) {
	if !isKurikulumOrKepsek(c) { return }

	var assignments []Assignment
	DB.Preload("Subject").Preload("Subject.Teacher").Find(&assignments)

	c.JSON(http.StatusOK, gin.H{"data": assignments})
}

// 3. Lihat guru yang membuat soal ulangan (semua Quiz dengan tipe "ujian", + info guru)
func ViewQuizzesByTeacher(c *gin.Context) {
	if !isKurikulumOrKepsek(c) { return }

	var quizzes []Quiz
	DB.Preload("Teacher").Preload("Subject").Find(&quizzes)

	c.JSON(http.StatusOK, gin.H{"data": quizzes})
}

// 4. Lihat data guru (role_id = 2)
func ViewAllTeachers(c *gin.Context) {
	if !isKurikulumOrKepsek(c) { return }

	var teachers []User
	DB.Where("role_id = ?", 2).Find(&teachers)

	c.JSON(http.StatusOK, gin.H{"data": teachers})
}

// 5. Lihat data siswa (role_id = 3)
func ViewAllStudents(c *gin.Context) {
	if !isKurikulumOrKepsek(c) { return }

	var students []User
	DB.Where("role_id = ?", 3).Preload("Class").Find(&students)

	c.JSON(http.StatusOK, gin.H{"data": students})
}
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type SubjectInput struct {
	SubjectName string `json:"subject_name" binding:"required"`
	TeacherID   string `json:"teacher_id" binding:"required"`
}

func CreateSubject(c *gin.Context) {
	if !isAdmin(c) { return }

	var input SubjectInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validasi: pastikan TeacherID itu beneran guru (role_id 2)
	var teacher User
	if err := DB.Where("id = ? AND role_id = 2", input.TeacherID).First(&teacher).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Guru tidak ditemukan atau bukan role Guru"})
		return
	}

	subject := Subject{
		SubjectName: input.SubjectName,
		TeacherID:   input.TeacherID,
	}
	DB.Create(&subject)

	c.JSON(http.StatusOK, gin.H{"message": "Mapel berhasil dibuat", "data": subject})
}

func GetSubjects(c *gin.Context) {
	// Admin, Guru, dan Siswa semua boleh lihat daftar mapel
	roleID, exists := c.Get("role_id")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak!"})
		return
	}
	r := uint(roleID.(float64))
	if r != 1 && r != 2 && r != 3 && r != 4 && r != 5 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak!"})
		return
	}

	var subjects []Subject
	DB.Preload("Teacher").Find(&subjects)

	c.JSON(http.StatusOK, gin.H{"data": subjects})
}

func UpdateSubject(c *gin.Context) {
	if !isAdmin(c) { return }

	id := c.Param("id")
	var subject Subject
	if err := DB.First(&subject, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mapel tidak ditemukan"})
		return
	}

	var input SubjectInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var teacher User
	if err := DB.Where("id = ? AND role_id = 2", input.TeacherID).First(&teacher).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Guru tidak ditemukan atau bukan role Guru"})
		return
	}

	DB.Model(&subject).Updates(map[string]interface{}{
		"SubjectName": input.SubjectName,
		"TeacherID":   input.TeacherID,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Mapel berhasil diperbarui"})
}

func DeleteSubject(c *gin.Context) {
	if !isAdmin(c) { return }

	id := c.Param("id")
	DB.Delete(&Subject{}, id)

	c.JSON(http.StatusOK, gin.H{"message": "Mapel dihapus"})
}
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type MaterialInput struct {
	SubjectID  uint   `json:"subject_id" binding:"required"`
	Title      string `json:"title" binding:"required"`
	ContentURL string `json:"content_url"`
}

func CreateMaterial(c *gin.Context) {
	roleID, _ := c.Get("role_id")
	if uint(roleID.(float64)) != 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak Anda bukan Guru."})
		return
	}

	userID, _ := c.Get("user_id")

	var input MaterialInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	material := Material{
		SubjectID:  input.SubjectID,
		Title:      input.Title,
		ContentURL: input.ContentURL,
		UploadedBy: userID.(string),
	}
	DB.Create(&material)

	c.JSON(http.StatusOK, gin.H{"message": "Materi baru berhasil diunggah", "data": material})
}

func GetMaterials(c *gin.Context) {
	roleID, _ := c.Get("role_id")
	if uint(roleID.(float64)) != 2 && uint(roleID.(float64)) != 3 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak"})
		return
	}

	var materials []Material
	DB.Find(&materials)

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil data materi",
		"data":    materials,
	})
}

func UpdateMaterial(c *gin.Context) {
	roleID, _ := c.Get("role_id")
	if uint(roleID.(float64)) != 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak Anda bukan Guru."})
		return
	}
	userID, _ := c.Get("user_id")

	id := c.Param("id")
	var material Material
	if err := DB.Where("id = ? AND uploaded_by = ?", id, userID.(string)).First(&material).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Materi tidak ditemukan atau bukan milik Anda"})
		return
	}

	var input MaterialInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	DB.Model(&material).Updates(map[string]interface{}{
		"SubjectID":  input.SubjectID,
		"Title":      input.Title,
		"ContentURL": input.ContentURL,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Materi berhasil diperbarui"})
}

func DeleteMaterial(c *gin.Context) {
	roleID, _ := c.Get("role_id")
	if uint(roleID.(float64)) != 2 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak Anda bukan Guru."})
		return
	}
	userID, _ := c.Get("user_id")

	id := c.Param("id")
	DB.Where("id = ? AND uploaded_by = ?", id, userID.(string)).Delete(&Material{})

	c.JSON(http.StatusOK, gin.H{"message": "Materi dihapus"})
}
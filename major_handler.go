package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type MajorInput struct {
	MajorName string `json:"major_name" binding:"required"`
}

func CreateMajor(c *gin.Context) {
	if !isAdmin(c) { return }
	var input MajorInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	major := Major{MajorName: input.MajorName}
	if err := DB.Create(&major).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Jurusan sudah ada atau gagal disimpan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Jurusan berhasil dibuat!", "data": major})
}

func GetMajors(c *gin.Context) {
	if !isAdmin(c) { return }
	var majors []Major
	DB.Find(&majors)
	c.JSON(http.StatusOK, gin.H{"data": majors})
}

func UpdateMajor(c *gin.Context) {
	if !isAdmin(c) { return }
	id := c.Param("id")
	var major Major
	if err := DB.First(&major, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Jurusan tidak ditemukan"})
		return
	}
	var input MajorInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	DB.Model(&major).Update("MajorName", input.MajorName)
	c.JSON(http.StatusOK, gin.H{"message": "Jurusan berhasil diperbarui"})
}

func DeleteMajor(c *gin.Context) {
	if !isAdmin(c) { return }
	id := c.Param("id")
	DB.Delete(&Major{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "Jurusan dihapus"})
}
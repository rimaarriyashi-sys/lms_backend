package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ScheduleInput struct {
	ClassID   uint   `json:"class_id" binding:"required"`
	SubjectID uint   `json:"subject_id" binding:"required"`
	DayOfWeek string `json:"day_of_week" binding:"required"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
}

func CreateSchedule(c *gin.Context) {
	if !isAdmin(c) { return }

	var input ScheduleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schedule := Schedule{
		ClassID:   input.ClassID,
		SubjectID: input.SubjectID,
		DayOfWeek: input.DayOfWeek,
		StartTime: input.StartTime,
		EndTime:   input.EndTime,
	}
	DB.Create(&schedule)

	c.JSON(http.StatusOK, gin.H{"message": "Jadwal berhasil dibuat", "data": schedule})
}

func GetSchedules(c *gin.Context) {
	// Semua role boleh lihat jadwal
	_, exists := c.Get("role_id")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak!"})
		return
	}

	var schedules []Schedule
	DB.Preload("Class").Preload("Subject").Preload("Subject.Teacher").Find(&schedules)

	c.JSON(http.StatusOK, gin.H{"data": schedules})
}

func GetScheduleByClass(c *gin.Context) {
	_, exists := c.Get("role_id")
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Akses Ditolak!"})
		return
	}

	classID := c.Param("class_id")
	var schedules []Schedule
	DB.Where("class_id = ?", classID).Preload("Subject").Preload("Subject.Teacher").Find(&schedules)

	c.JSON(http.StatusOK, gin.H{"data": schedules})
}

func UpdateSchedule(c *gin.Context) {
	if !isAdmin(c) { return }

	id := c.Param("id")
	var schedule Schedule
	if err := DB.First(&schedule, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Jadwal tidak ditemukan"})
		return
	}

	var input ScheduleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	DB.Model(&schedule).Updates(map[string]interface{}{
		"ClassID":   input.ClassID,
		"SubjectID": input.SubjectID,
		"DayOfWeek": input.DayOfWeek,
		"StartTime": input.StartTime,
		"EndTime":   input.EndTime,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Jadwal berhasil diperbarui"})
}

func DeleteSchedule(c *gin.Context) {
	if !isAdmin(c) { return }

	id := c.Param("id")
	DB.Delete(&Schedule{}, id)

	c.JSON(http.StatusOK, gin.H{"message": "Jadwal dihapus"})
}
package handlers

import (
	"fmt"
	"net/http"
	"noteshare-backend/config"
	"noteshare-backend/database"
	"noteshare-backend/middleware"
	"noteshare-backend/models"
	"noteshare-backend/utils"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetNotes - GET /api/v1/notes
func GetNotes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	courseID := c.Query("course_id")

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	query := database.DB.Model(&models.Note{}).
		Preload("User").
		Preload("Course").
		Where("is_public = ?", true)

	if search != "" {
		query = query.Where("title ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if courseID != "" {
		query = query.Where("course_id = ?", courseID)
	}

	var total int64
	query.Count(&total)

	var notes []models.Note
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&notes).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve notes")
		return
	}

	utils.PaginatedSuccessResponse(c, http.StatusOK, "Notes retrieved", notes, page, limit, total)
}

// GetMyNotes - GET /api/v1/notes/my
func GetMyNotes(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	var total int64
	database.DB.Model(&models.Note{}).Where("user_id = ?", userID).Count(&total)

	var notes []models.Note
	if err := database.DB.Where("user_id = ?", userID).
		Preload("Course").
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&notes).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve notes")
		return
	}

	utils.PaginatedSuccessResponse(c, http.StatusOK, "My notes retrieved", notes, page, limit, total)
}

// GetNote - GET /api/v1/notes/:id
func GetNote(c *gin.Context) {
	id := c.Param("id")

	var note models.Note
	if err := database.DB.Preload("User").Preload("Course").First(&note, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Note not found")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Note retrieved", note)
}

// UploadNote - POST /api/v1/notes
func UploadNote(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req models.CreateNoteRequest
	if err := c.ShouldBind(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Validate course exists
	var course models.Course
	if err := database.DB.First(&course, req.CourseID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Course not found")
		return
	}

	// Handle file upload
	file, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "File is required")
		return
	}

	// Validate file size
	maxSize := config.AppConfig.MaxFileSizeMB * 1024 * 1024
	if file.Size > maxSize {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("File size exceeds %d MB limit", config.AppConfig.MaxFileSizeMB))
		return
	}

	// Validate file type
	allowedTypes := map[string]bool{
		".pdf":  true,
		".doc":  true,
		".docx": true,
		".ppt":  true,
		".pptx": true,
		".txt":  true,
		".png":  true,
		".jpg":  true,
		".jpeg": true,
	}
	ext := filepath.Ext(file.Filename)
	if !allowedTypes[ext] {
		utils.ErrorResponse(c, http.StatusBadRequest, "File type not allowed. Allowed: pdf, doc, docx, ppt, pptx, txt, png, jpg, jpeg")
		return
	}

	// Generate unique filename
	uniqueFilename := fmt.Sprintf("%d_%d%s", userID, time.Now().UnixNano(), ext)
	filePath := filepath.Join(config.AppConfig.UploadDir, uniqueFilename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save file")
		return
	}

	note := models.Note{
		Title:       req.Title,
		Description: req.Description,
		FilePath:    filePath,
		FileName:    file.Filename,
		FileSize:    file.Size,
		FileType:    ext,
		CourseID:    req.CourseID,
		UserID:      userID,
		IsPublic:    req.IsPublic,
	}

	if err := database.DB.Create(&note).Error; err != nil {
		os.Remove(filePath) // cleanup file on DB error
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save note")
		return
	}

	database.DB.Preload("User").Preload("Course").First(&note, note.ID)
	utils.SuccessResponse(c, http.StatusCreated, "Note uploaded successfully", note)
}

// DownloadNote - GET /api/v1/notes/:id/download
func DownloadNote(c *gin.Context) {
	id := c.Param("id")

	var note models.Note
	if err := database.DB.First(&note, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Note not found")
		return
	}

	if !note.IsPublic {
		userID, exists := c.Get("userID")
		if !exists || userID.(uint) != note.UserID {
			utils.ErrorResponse(c, http.StatusForbidden, "This note is private")
			return
		}
	}

	// Increment download count
	database.DB.Model(&note).UpdateColumn("downloads", note.Downloads+1)

	c.FileAttachment(note.FilePath, note.FileName)
}

// UpdateNote - PUT /api/v1/notes/:id
func UpdateNote(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := c.Param("id")

	var note models.Note
	if err := database.DB.First(&note, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Note not found")
		return
	}

	if note.UserID != userID {
		utils.ErrorResponse(c, http.StatusForbidden, "You are not authorized to update this note")
		return
	}

	var req models.UpdateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}

	database.DB.Model(&note).Updates(updates)
	utils.SuccessResponse(c, http.StatusOK, "Note updated successfully", note)
}

// DeleteNote - DELETE /api/v1/notes/:id
func DeleteNote(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := c.Param("id")

	var note models.Note
	if err := database.DB.First(&note, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Note not found")
		return
	}

	if note.UserID != userID {
		utils.ErrorResponse(c, http.StatusForbidden, "You are not authorized to delete this note")
		return
	}

	// Remove file from storage
	os.Remove(note.FilePath)

	database.DB.Delete(&note)
	utils.SuccessResponse(c, http.StatusOK, "Note deleted successfully", nil)
}
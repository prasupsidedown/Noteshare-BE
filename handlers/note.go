package handlers

import (
	"context"
	"fmt"
	"net/http"
	"noteshare-be/config"
	"noteshare-be/database"
	"noteshare-be/middleware"
	"noteshare-be/models"
	"noteshare-be/utils"
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

	// Save file temporarily to upload to Cloudinary
	tempFilename := fmt.Sprintf("%d_%d%s", userID, time.Now().UnixNano(), ext)
	tempFilePath := filepath.Join(os.TempDir(), tempFilename)

	if err := c.SaveUploadedFile(file, tempFilePath); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to process file")
		return
	}
	defer os.Remove(tempFilePath) // cleanup temp file

	// Upload to Cloudinary
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	cloudURL, cloudinaryID, err := utils.UploadToCloudinary(ctx, tempFilePath, "noteshare/notes")
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to upload file to cloud storage")
		return
	}

	note := models.Note{
		Title:         req.Title,
		Description:   req.Description,
		CloudinaryURL: cloudURL,
		CloudinaryID:  cloudinaryID,
		FileName:      file.Filename,
		FileSize:      file.Size,
		FileType:      ext,
		Semester:      req.Semester,
		CourseID:      req.CourseID,
		UserID:        userID,
		IsPublic:      req.IsPublic,
	}

	if err := database.DB.Create(&note).Error; err != nil {
		// Cleanup file from Cloudinary on DB error
		deleteCtx, deleteCancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		utils.DeleteFromCloudinary(deleteCtx, cloudinaryID)
		deleteCancel()
		
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save note")
		return
	}

	if err := database.DB.Preload("User").Preload("Course").First(&note, note.ID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve created note")
		return
	}

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
		userID := middleware.GetUserID(c)
		if userID == 0 || userID != note.UserID {
			utils.ErrorResponse(c, http.StatusForbidden, "This note is private")
			return
		}
	}

	// Increment download count
	if err := database.DB.Model(&note).UpdateColumn("downloads", note.Downloads+1).Error; err != nil {
		// Log error but don't fail the download
		fmt.Printf("Warning: Failed to update download count: %v\n", err)
	}

	// Redirect to Cloudinary URL
	c.Redirect(http.StatusFound, note.CloudinaryURL)
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
	if req.Semester != "" {
		updates["semester"] = req.Semester
	}
	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}

	if err := database.DB.Model(&note).Updates(updates).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update note")
		return
	}

	// Reload note with relationships to get updated data
	if err := database.DB.Preload("User").Preload("Course").First(&note, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve updated note")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Note updated successfully", note)
}

// DeleteNote - DELETE /api/v1/notes/:id
// DeleteNote - DELETE /api/v1/notes/:id
func DeleteNote(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := c.Param("id")
	
	// Get note first to retrieve CloudinaryID
	var note models.Note
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&note).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Note not found")
		return
	}
	
	// Delete from Cloudinary
	deleteCtx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	
	if err := utils.DeleteFromCloudinary(deleteCtx, note.CloudinaryID); err != nil {
		// Log error but continue with DB deletion
		fmt.Printf("Warning: Failed to delete file from Cloudinary: %v\n", err)
	}
	
	// Delete from database
	if err := database.DB.Delete(&note).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete note")
		return
	}
	
	utils.SuccessResponse(c, http.StatusOK, "Note deleted successfully", nil)
}
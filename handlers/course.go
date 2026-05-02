package handlers

import (
	"net/http"
	"noteshare-backend/database"
	"noteshare-backend/middleware"
	"noteshare-backend/models"
	"noteshare-backend/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetCourses - GET /api/v1/courses
func GetCourses(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	query := database.DB.Model(&models.Course{}).Preload("User")
	if search != "" {
		query = query.Where("name ILIKE ? OR code ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	query.Count(&total)

	var courses []models.Course
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&courses).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve courses")
		return
	}

	utils.PaginatedSuccessResponse(c, http.StatusOK, "Courses retrieved", courses, page, limit, total)
}

// GetCourse - GET /api/v1/courses/:id
func GetCourse(c *gin.Context) {
	id := c.Param("id")

	var course models.Course
	if err := database.DB.Preload("User").Preload("Notes").First(&course, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Course not found")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Course retrieved", course)
}

// CreateCourse - POST /api/v1/courses
func CreateCourse(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req models.CreateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	course := models.Course{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Semester:    req.Semester,
		UserID:      userID,
	}

	if err := database.DB.Create(&course).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create course")
		return
	}

	database.DB.Preload("User").First(&course, course.ID)
	utils.SuccessResponse(c, http.StatusCreated, "Course created successfully", course)
}

// UpdateCourse - PUT /api/v1/courses/:id
func UpdateCourse(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := c.Param("id")

	var course models.Course
	if err := database.DB.First(&course, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Course not found")
		return
	}

	if course.UserID != userID {
		utils.ErrorResponse(c, http.StatusForbidden, "You are not authorized to update this course")
		return
	}

	var req models.UpdateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Code != "" {
		updates["code"] = req.Code
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Semester != "" {
		updates["semester"] = req.Semester
	}

	database.DB.Model(&course).Updates(updates)
	utils.SuccessResponse(c, http.StatusOK, "Course updated successfully", course)
}

// DeleteCourse - DELETE /api/v1/courses/:id
func DeleteCourse(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := c.Param("id")

	var course models.Course
	if err := database.DB.First(&course, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Course not found")
		return
	}

	if course.UserID != userID {
		utils.ErrorResponse(c, http.StatusForbidden, "You are not authorized to delete this course")
		return
	}

	database.DB.Delete(&course)
	utils.SuccessResponse(c, http.StatusOK, "Course deleted successfully", nil)
}
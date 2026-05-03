package handlers

import (
	"net/http"
	"noteshare-be/database"
	"noteshare-be/models"
	"noteshare-be/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Register - POST /api/v1/auth/register
func Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Cek email sudah terdaftar
	var existing models.User
	if result := database.DB.Where("email = ?", req.Email).First(&existing); result.Error == nil {
		utils.ErrorResponse(c, http.StatusConflict, "Email sudah terdaftar")
		return
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal memproses password")
		return
	}

	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashed),
	}

	if err := database.DB.Create(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat akun")
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Email)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat token")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Registrasi berhasil", gin.H{
		"user":  user.ToResponse(),
		"token": token,
	})
}

// Login - POST /api/v1/auth/login
func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Email atau password salah")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Email atau password salah")
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Email)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat token")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Login berhasil", gin.H{
		"user":  user.ToResponse(),
		"token": token,
	})
}

// GetProfile - GET /api/v1/auth/me
func GetProfile(c *gin.Context) {
	userID, _ := c.Get("userID")

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User tidak ditemukan")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Profil berhasil diambil", user.ToResponse())
}

// UpdateProfile - PUT /api/v1/auth/me
func UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("userID")

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User tidak ditemukan")
		return
	}

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Bio != "" {
		updates["bio"] = req.Bio
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}

	if err := database.DB.Model(&user).Updates(updates).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengupdate profil")
		return
	}

	// Reload user data
	if err := database.DB.First(&user, userID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data profil")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Profil berhasil diupdate", user.ToResponse())
}
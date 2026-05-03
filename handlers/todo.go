package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"noteshare-be/config"
	"noteshare-be/database"
	"noteshare-be/middleware"
	"noteshare-be/models"
	"noteshare-be/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ─── Anthropic API Types ───────────────────────────────────────────────────

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicResponse struct {
	Content []anthropicContent `json:"content"`
}

type generatedTodo struct {
	Title string           `json:"title"`
	Items []models.TodoItem `json:"items"`
}

// ─── Handlers ─────────────────────────────────────────────────────────────

// GenerateTodoList - POST /api/v1/notes/:id/todos/generate
// Uses Anthropic Claude to auto-generate a to-do list from the note metadata
func GenerateTodoList(c *gin.Context) {
	userID := middleware.GetUserID(c)
	noteID := c.Param("id")

	var note models.Note
	if err := database.DB.Preload("Course").First(&note, noteID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Note not found")
		return
	}

	if config.AppConfig.AnthropicAPIKey == "" {
		utils.ErrorResponse(c, http.StatusServiceUnavailable, "AI service is not configured")
		return
	}

	// Build prompt for Claude
	prompt := fmt.Sprintf(`Kamu adalah asisten akademik. Berdasarkan informasi catatan kuliah berikut, buatkan to-do list belajar yang terstruktur dan praktis.

Informasi Catatan:
- Judul: %s
- Deskripsi: %s
- Mata Kuliah: %s (%s)

Balas HANYA dengan JSON valid (tanpa markdown/backtick) dengan format berikut:
{
  "title": "To-Do List: <judul catatan>",
  "items": [
    {
      "task": "deskripsi tugas yang spesifik",
      "priority": "high|medium|low",
      "completed": false
    }
  ]
}

Buat 5-8 to-do item yang relevan dan actionable berdasarkan topik catatan tersebut.`,
		note.Title,
		note.Description,
		note.Course.Name,
		note.Course.Code,
	)

	// Call Anthropic API
	generated, err := callAnthropicAPI(prompt)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate to-do list: "+err.Error())
		return
	}

	// Save to database
	todoList := models.TodoList{
		NoteID: note.ID,
		UserID: userID,
		Title:  generated.Title,
		Items:  generated.Items,
	}

	if err := database.DB.Create(&todoList).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save to-do list")
		return
	}

	// Reload with relationships
	if err := database.DB.Preload("Note").Preload("User").First(&todoList, todoList.ID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve created to-do list")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "To-do list generated successfully", todoList)
}

// GetTodoLists - GET /api/v1/notes/:id/todos
func GetTodoLists(c *gin.Context) {
	userID := middleware.GetUserID(c)
	noteID := c.Param("id")

	var todos []models.TodoList
	if err := database.DB.
		Where("note_id = ? AND user_id = ?", noteID, userID).
		Preload("Note").
		Preload("User").
		Order("created_at DESC").
		Find(&todos).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve to-do lists")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "To-do lists retrieved", todos)
}

// GetTodoList - GET /api/v1/todos/:id
func GetTodoList(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := c.Param("id")

	var todo models.TodoList
	if err := database.DB.Preload("Note").First(&todo, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "To-do list not found")
		return
	}

	if todo.UserID != userID {
		utils.ErrorResponse(c, http.StatusForbidden, "You are not authorized to view this to-do list")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "To-do list retrieved", todo)
}

// UpdateTodoItem - PATCH /api/v1/todos/:id/items
// Toggle complete status of a specific to-do item by index
func UpdateTodoItem(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := c.Param("id")

	var todo models.TodoList
	if err := database.DB.First(&todo, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "To-do list not found")
		return
	}

	if todo.UserID != userID {
		utils.ErrorResponse(c, http.StatusForbidden, "You are not authorized to update this to-do list")
		return
	}

	var req models.UpdateTodoItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if req.Index < 0 || req.Index >= len(todo.Items) {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Index out of range. Valid range: 0-%d", len(todo.Items)-1))
		return
	}

	todo.Items[req.Index].Completed = req.Completed

	if err := database.DB.Save(&todo).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update to-do item")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "To-do item updated", todo)
}

// DeleteTodoList - DELETE /api/v1/todos/:id
func DeleteTodoList(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := c.Param("id")

	var todo models.TodoList
	if err := database.DB.First(&todo, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "To-do list not found")
		return
	}

	// Check authorization FIRST
	if todo.UserID != userID {
		utils.ErrorResponse(c, http.StatusForbidden, "You are not authorized to delete this to-do list")
		return
	}

	// Then delete from database
	if err := database.DB.Delete(&todo).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete to-do list")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "To-do list deleted", nil)
}

// GetMyTodoLists - GET /api/v1/todos/my
func GetMyTodoLists(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	var todos []models.TodoList
	var total int64

	// Get total count
	if err := database.DB.Model(&models.TodoList{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to count to-do lists")
		return
	}

	// Get paginated results
	if err := database.DB.Where("user_id = ?", userID).
		Preload("Note").
		Preload("User").
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&todos).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve to-do lists")
		return
	}

	utils.PaginatedSuccessResponse(c, http.StatusOK, "My to-do lists retrieved", todos, page, limit, total)
}

// AddTodoItem - POST /api/v1/todos/:id/items
// Manually add a to-do item to a to-do list
func AddTodoItem(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id := c.Param("id")

	var todo models.TodoList
	if err := database.DB.First(&todo, id).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "To-do list not found")
		return
	}

	if todo.UserID != userID {
		utils.ErrorResponse(c, http.StatusForbidden, "You are not authorized to add items to this to-do list")
		return
	}

	var req models.ManualTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Add new item to the list
	newItem := models.TodoItem{
		Task:      req.Task,
		Priority:  req.Priority,
		Completed: false,
	}

	todo.Items = append(todo.Items, newItem)

	if err := database.DB.Save(&todo).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to add to-do item")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "To-do item added successfully", todo)
}

// ─── Anthropic API Helper ──────────────────────────────────────────────────

func callAnthropicAPI(prompt string) (*generatedTodo, error) {
	reqBody := anthropicRequest{
		Model:     "claude-sonnet-4-20250514",
		MaxTokens: 1024,
		Messages: []anthropicMessage{
			{Role: "user", Content: prompt},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", config.AppConfig.AnthropicAPIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("anthropic API error (status %d): %s", resp.StatusCode, string(body))
	}

	var anthropicResp anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return nil, err
	}

	if len(anthropicResp.Content) == 0 {
		return nil, fmt.Errorf("empty response from Anthropic API")
	}

	var result generatedTodo
	if err := json.Unmarshal([]byte(anthropicResp.Content[0].Text), &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %v", err)
	}

	return &result, nil
}
package routes

import (
	"noteshare-be/handlers"
	"noteshare-be/middleware"

	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine) {
	// API v1
	api := r.Group("/api/v1")

	// ─── Auth Routes (No Auth Required) ───────────────────────────────────
	auth := api.Group("/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
	}

	// ─── Protected Routes (Auth Required) ──────────────────────────────────
	protected := api.Group("")
	protected.Use(middleware.AuthRequired())
	{
		// Auth
		authProtected := protected.Group("/auth")
		{
			authProtected.GET("/me", handlers.GetProfile)
			authProtected.PUT("/me", handlers.UpdateProfile)
		}

		// Courses
		coursesGroup := protected.Group("/courses")
		{
			coursesGroup.GET("", handlers.GetCourses)        // Public list
			coursesGroup.POST("", handlers.CreateCourse)     // Create
			coursesGroup.GET("/:id", handlers.GetCourse)     // Get single
			coursesGroup.PUT("/:id", handlers.UpdateCourse)  // Update
			coursesGroup.DELETE("/:id", handlers.DeleteCourse) // Delete
		}

		// Notes
		notesGroup := protected.Group("/notes")
		{
			notesGroup.GET("", handlers.GetNotes)            // Public list
			notesGroup.GET("/my", handlers.GetMyNotes)       // My notes
			notesGroup.POST("", handlers.UploadNote)         // Upload
			notesGroup.GET("/:id", handlers.GetNote)         // Get single
			notesGroup.GET("/:id/download", handlers.DownloadNote) // Download
			notesGroup.PUT("/:id", handlers.UpdateNote)      // Update
			notesGroup.DELETE("/:id", handlers.DeleteNote)   // Delete

			// Todo Lists for a specific note
			notesGroup.GET("/:id/todos", handlers.GetTodoLists)            // Get todos for note
			notesGroup.POST("/:id/todos/generate", handlers.GenerateTodoList) // AI generate todos
		}

		// Todo Lists
		todosGroup := protected.Group("/todos")
		{
			todosGroup.GET("/my", handlers.GetMyTodoLists)   // My todos
			todosGroup.GET("/:id", handlers.GetTodoList)     // Get single
			todosGroup.PATCH("/:id/items", handlers.UpdateTodoItem) // Toggle item
			todosGroup.POST("/:id/items", handlers.AddTodoItem)   // Add item manually
			todosGroup.DELETE("/:id", handlers.DeleteTodoList)   // Delete list
		}
	}
}

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/Ponloe/cinemesh-core/internal/admin"
	"github.com/Ponloe/cinemesh-core/internal/auth"
	"github.com/Ponloe/cinemesh-core/internal/database"
	"github.com/Ponloe/cinemesh-core/internal/users"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not loaded, continuing with environment variables")
	}

	if m := os.Getenv("GIN_MODE"); m != "" {
		gin.SetMode(m)
	}

	if err := database.Connect(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// run migrations to create tables
	if err := database.Migrate(&users.User{}); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	r := gin.Default()

	// Load HTML templates
	r.LoadHTMLGlob("internal/admin/templates/*")

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth routes
	r.POST("/login", auth.LoginHandler)
	r.POST("/users", users.CreateUserHandler)
	r.GET("/users/:id", users.GetUserHandler)

	// Admin login (public)
	r.GET("/admin/login", admin.LoginFormHandler)
	r.POST("/admin/login", admin.LoginPostHandler)

	// Protected routes
	r.GET("/me", auth.RequireAuth(), auth.MeHandler)

	// Admin routes (protected by auth + admin role)
	adminGroup := r.Group("/admin", auth.RequireAuth(), auth.RequireAdmin())
	{
		adminGroup.GET("/", admin.DashboardHandler)
		adminGroup.GET("/users", users.ListUsersHandler)
		adminGroup.GET("/users/new", users.NewUserFormHandler)
		adminGroup.POST("/users", users.CreateUserAdminHandler)
		adminGroup.GET("/users/:id/edit", users.EditUserFormHandler)
		adminGroup.POST("/users/:id", users.UpdateUserHandler)        // For updates (PUT via _method)
		adminGroup.POST("/users/:id/delete", users.DeleteUserHandler) // For deletes
	}

	r.POST("/admin/logout", func(c *gin.Context) {
		c.SetCookie("token", "", -1, "/", "", false, true)
		c.Redirect(http.StatusFound, "/admin/login")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}

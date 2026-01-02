package admin

import (
	"net/http"

	"github.com/Ponloe/cinemesh-core/internal/auth"
	"github.com/Ponloe/cinemesh-core/internal/database"
	"github.com/Ponloe/cinemesh-core/internal/users"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func DashboardHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.html", gin.H{"title": "Admin Dashboard"})
}

func LoginFormHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{"title": "Admin Login"})
}

func LoginPostHandler(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	if email == "" || password == "" {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{"error": "Email and password required", "title": "Admin Login"})
		return
	}

	var u users.User
	if err := database.DB.First(&u, "email = ?", email).Error; err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"error": "Invalid credentials", "title": "Admin Login"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"error": "Invalid credentials", "title": "Admin Login"})
		return
	}

	token, err := auth.GenerateToken(&u)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{"error": "Failed to generate token", "title": "Admin Login"})
		return
	}

	c.SetCookie("token", token, 86400, "/", "", false, true) // 24 hours
	c.Redirect(http.StatusFound, "/admin")
}

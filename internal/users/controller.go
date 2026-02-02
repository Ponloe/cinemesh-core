package users

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"time"

	"github.com/Ponloe/cinemesh-core/internal/database"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type CreateUserDTO struct {
	Username string `json:"username" binding:"required,min=3"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type UserResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toResponse(u *User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		AvatarURL: u.AvatarURL,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func HashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(b), err
}

func CreateUserHandler(c *gin.Context) {
	var input struct {
		Username string `gorm:"size:50;unique;not null"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Role     string `json:"role"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	// Set default role if not provided
	if input.Role == "" {
		input.Role = "user"
	}

	// Validate role enum
	if input.Role != "user" && input.Role != "admin" {
		c.JSON(400, gin.H{"error": "role must be 'user' or 'admin'"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to hash password"})
		return
	}

	user := User{
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
		Role:         input.Role,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			c.JSON(409, gin.H{"error": "user with this email already exists"})
			return
		}
		c.JSON(500, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(201, gin.H{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	})
}

func GetUserHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var user User
	if err := database.DB.First(&user, uint(id)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toResponse(&user))
}

func ListUsersHandler(c *gin.Context) {
	var users []User
	if err := database.DB.Order("id ASC").Find(&users).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}
	c.HTML(http.StatusOK, "users.html", gin.H{"users": users})
}

func NewUserFormHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "user_form.html", gin.H{"user": User{}, "action": "/admin/users", "method": "POST"})
}

func CreateUserAdminHandler(c *gin.Context) {
	username := c.PostForm("username")
	email := c.PostForm("email")
	password := c.PostForm("password")
	role := c.PostForm("role")
	if role == "" {
		role = "user"
	}

	hashed, err := HashPassword(password)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "failed to hash password"})
		return
	}

	user := User{
		Username:     username,
		Email:        email,
		PasswordHash: hashed,
		Role:         role,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/admin/users")
}

func EditUserFormHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "invalid id"})
		return
	}

	var user User
	if err := database.DB.First(&user, uint(id)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "user not found"})
			return
		}
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "user_form.html", gin.H{"user": user, "action": "/admin/users/" + idStr, "method": "POST"})
}

func UpdateUserHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "invalid id"})
		return
	}

	var user User
	if err := database.DB.First(&user, uint(id)).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "user not found"})
		return
	}

	username := c.PostForm("username")
	email := c.PostForm("email")
	role := c.PostForm("role")
	password := c.PostForm("password") // optional

	user.Username = username
	user.Email = email
	user.Role = role
	if password != "" {
		hashed, err := HashPassword(password)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "failed to hash password"})
			return
		}
		user.PasswordHash = hashed
	}

	if err := database.DB.Save(&user).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/admin/users")
}

func DeleteUserHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "invalid id"})
		return
	}

	if err := database.DB.Delete(&User{}, uint(id)).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/admin/users")
}

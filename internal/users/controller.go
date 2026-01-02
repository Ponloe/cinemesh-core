package users

import (
	"errors"
	"net/http"
	"strconv"

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
	var body CreateUserDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashed, err := HashPassword(body.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user := User{
		Username:     body.Username,
		Email:        body.Email,
		PasswordHash: hashed,
		Role:         "user",
	}

	if err := database.DB.Create(&user).Error; err != nil {
		// handle unique constraint / validation errors simply
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toResponse(&user))
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
	if err := database.DB.Find(&users).Error; err != nil {
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

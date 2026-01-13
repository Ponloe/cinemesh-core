package admin

import (
	"net/http"
	"strconv"

	"github.com/Ponloe/cinemesh-core/internal/auth"
	"github.com/Ponloe/cinemesh-core/internal/database"
	"github.com/Ponloe/cinemesh-core/internal/movies"
	"github.com/Ponloe/cinemesh-core/internal/tmdb"
	"github.com/Ponloe/cinemesh-core/internal/users"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var (
	tmdbClient  *tmdb.Client
	tmdbFetcher *tmdb.MovieFetcher
)

func init() {
	// Initialize TMDb client
	tmdbConfig := tmdb.NewConfig()
	tmdbClient = tmdb.NewClient(tmdbConfig)
	tmdbFetcher = tmdb.NewMovieFetcher(tmdbClient)
}

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

func TMDbSearchHandler(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter required"})
		return
	}

	results, err := tmdbClient.SearchMovies(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

func ImportFromTMDbHandler(c *gin.Context) {
	var req struct {
		TMDbID int `json:"tmdb_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	movie, err := tmdbFetcher.FetchMovieByTMDbID(req.TMDbID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch from TMDb: " + err.Error()})
		return
	}

	var existing movies.Movie
	if err := database.DB.Where("slug = ?", movie.Slug).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":    "movie already exists",
			"movie_id": existing.ID,
		})
		return
	}
	var dbGenres []movies.Genre
	for _, tmdbGenre := range movie.Genres {
		var genre movies.Genre
		if err := database.DB.Where("LOWER(name) = LOWER(?)", tmdbGenre.Name).First(&genre).Error; err == nil {
			dbGenres = append(dbGenres, genre)
		} else {
			genre = movies.Genre{Name: tmdbGenre.Name}
			if err := database.DB.Create(&genre).Error; err == nil {
				dbGenres = append(dbGenres, genre)
			}
		}
	}
	movie.Genres = dbGenres

	if err := database.DB.Create(movie).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save movie: " + err.Error()})
		return
	}

	database.DB.Preload("Genres").First(movie, movie.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "movie imported successfully",
		"movie":   movie,
	})
}

func PrefillFromTMDbHandler(c *gin.Context) {
	tmdbIDStr := c.Query("tmdb_id")
	if tmdbIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tmdb_id required"})
		return
	}

	tmdbID, err := strconv.Atoi(tmdbIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tmdb_id"})
		return
	}

	movie, err := tmdbFetcher.FetchMovieByTMDbID(tmdbID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, movie)
}

func TMDbSearchPageHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "tmdb_search.html", gin.H{
		"title": "Import from TMDb",
	})
}

package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Ponloe/cinemesh-core/internal/database"
	"github.com/Ponloe/cinemesh-core/internal/movies"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)


// ================================
// MOVIES
// ================================

// ListMoviesPublicHandler returns paginated list of movies
func ListMoviesPublicHandler(c *gin.Context) {

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	genre := c.Query("genre")

	if page < 1 {
		page = 1
	}

	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	query := database.DB.Preload("Genres").Preload("Cast.Person")

	if search != "" {
		query = query.Where("title ILIKE ?", "%"+search+"%")
	}

	if genre != "" {
		query = query.Joins("JOIN movie_genres ON movie_genres.movie_id = movies.id").
			Joins("JOIN genres ON genres.id = movie_genres.genre_id").
			Where("LOWER(genres.name) = LOWER(?)", genre)
	}

	var total int64
	query.Model(&movies.Movie{}).Count(&total)

	var movieList []movies.Movie

	if err := query.
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&movieList).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": movieList,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}


// GetMoviePublicHandler returns movie by ID or slug
func GetMoviePublicHandler(c *gin.Context) {

	identifier := c.Param("id")

	var movie movies.Movie
	var err error

	if id, parseErr := strconv.Atoi(identifier); parseErr == nil {

		err = database.DB.
			Preload("Genres").
			Preload("Cast.Person").
			First(&movie, id).Error

	} else {

		err = database.DB.
			Preload("Genres").
			Preload("Cast.Person").
			Where("slug = ?", identifier).
			First(&movie).Error
	}

	if err != nil {

		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": movie})
}


// ================================
// SHOWTIMES PROXY
// ================================

// GetMovieShowtimesPublicHandler calls the ticketing subsystem
func GetMovieShowtimesPublicHandler(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id"})
		return
	}

	var movie movies.Movie

	if err := database.DB.First(&movie, id).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		return
	}

	// Encode movie title (important for spaces)
	encodedTitle := url.QueryEscape(movie.Title)

	url := "http://localhost:8000/showtimes?movie_title=" + encodedTitle

	resp, err := http.Get(url)

	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "ticketing service unavailable"})
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read ticketing response"})
		return
	}

	var result interface{}

	if err := json.Unmarshal(body, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid ticketing response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}


// ================================
// GENRES
// ================================

func ListGenresPublicHandler(c *gin.Context) {

	var genreList []movies.Genre

	if err := database.DB.Order("name ASC").Find(&genreList).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": genreList})
}


func GetGenrePublicHandler(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid genre id"})
		return
	}

	var genre movies.Genre

	if err := database.DB.First(&genre, id).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "genre not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": genre})
}


// ================================
// PEOPLE
// ================================

func ListPeoplePublicHandler(c *gin.Context) {

	var people []movies.Person

	if err := database.DB.Order("name ASC").Find(&people).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": people})
}


func GetPersonPublicHandler(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid person id"})
		return
	}

	var person movies.Person

	if err := database.DB.First(&person, id).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "person not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		return
	}

	c.JSON(http.StatusOK, gin.H{"data": person})
}


// ================================
// SEARCH
// ================================

func SearchPublicHandler(c *gin.Context) {

	query := c.Query("q")

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter required"})
		return
	}

	searchPattern := "%" + query + "%"

	var movieResults []movies.Movie

	if err := database.DB.
		Preload("Genres").
		Where("title ILIKE ?", searchPattern).
		Limit(10).
		Find(&movieResults).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": movieResults,
	})
}


// ================================
// STATS
// ================================

func GetStatsPublicHandler(c *gin.Context) {

	var movieCount int64
	var genreCount int64
	var peopleCount int64

	database.DB.Model(&movies.Movie{}).Count(&movieCount)
	database.DB.Model(&movies.Genre{}).Count(&genreCount)
	database.DB.Model(&movies.Person{}).Count(&peopleCount)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"total_movies": movieCount,
			"total_genres": genreCount,
			"total_people": peopleCount,
		},
	})
}
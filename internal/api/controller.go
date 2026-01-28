package api

import (
	"net/http"
	"strconv"

	"github.com/Ponloe/cinemesh-core/internal/database"
	"github.com/Ponloe/cinemesh-core/internal/movies"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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

	// Search by title
	if search != "" {
		query = query.Where("title ILIKE ?", "%"+search+"%")
	}

	// Filter by genre
	if genre != "" {
		query = query.Joins("JOIN movie_genres ON movie_genres.movie_id = movies.id").
			Joins("JOIN genres ON genres.id = movie_genres.genre_id").
			Where("LOWER(genres.name) = LOWER(?)", genre)
	}

	var total int64
	query.Model(&movies.Movie{}).Count(&total)

	var movieList []movies.Movie
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&movieList).Error; err != nil {
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

// GetMoviePublicHandler returns a single movie by ID or slug
func GetMoviePublicHandler(c *gin.Context) {
	identifier := c.Param("id")

	var movie movies.Movie
	var err error

	// Try to parse as ID first
	if id, parseErr := strconv.Atoi(identifier); parseErr == nil {
		err = database.DB.Preload("Genres").Preload("Cast.Person").First(&movie, id).Error
	} else {
		// Otherwise treat as slug
		err = database.DB.Preload("Genres").Preload("Cast.Person").Where("slug = ?", identifier).First(&movie).Error
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

// ListGenresPublicHandler returns all genres
func ListGenresPublicHandler(c *gin.Context) {
	var genreList []movies.Genre
	if err := database.DB.Order("name ASC").Find(&genreList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": genreList})
}

// GetGenrePublicHandler returns a genre with its movies
func GetGenrePublicHandler(c *gin.Context) {
	identifier := c.Param("id")

	var genre movies.Genre
	var err error

	// Try to parse as ID first
	if id, parseErr := strconv.Atoi(identifier); parseErr == nil {
		err = database.DB.First(&genre, id).Error
	} else {
		// Otherwise treat as name
		err = database.DB.Where("LOWER(name) = LOWER(?)", identifier).First(&genre).Error
	}

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "genre not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// Get movies for this genre
	var movieList []movies.Movie
	database.DB.Preload("Genres").
		Joins("JOIN movie_genres ON movie_genres.movie_id = movies.id").
		Where("movie_genres.genre_id = ?", genre.ID).
		Order("title ASC").
		Find(&movieList)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"genre":  genre,
			"movies": movieList,
		},
	})
}

// ListPeoplePublicHandler returns paginated list of people
func ListPeoplePublicHandler(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	query := database.DB.Model(&movies.Person{})

	// Search by name
	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	var total int64
	query.Count(&total)

	var people []movies.Person
	if err := query.Offset(offset).Limit(limit).Order("name ASC").Find(&people).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": people,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetPersonPublicHandler returns a person with their movies
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

	// Get movies for this person
	type MovieWithRole struct {
		movies.Movie
		Role          string `json:"role"`
		CharacterName string `json:"character_name,omitempty"`
		CastOrder     *int   `json:"cast_order,omitempty"`
	}

	var movieRoles []MovieWithRole
	database.DB.Table("movies").
		Select("movies.*, movie_people.role, movie_people.character_name, movie_people.cast_order").
		Joins("JOIN movie_people ON movie_people.movie_id = movies.id").
		Where("movie_people.person_id = ?", id).
		Order("movies.release_date DESC").
		Scan(&movieRoles)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"person": person,
			"movies": movieRoles,
		},
	})
}

// SearchPublicHandler performs global search across movies, people, and genres
func SearchPublicHandler(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' required"})
		return
	}

	searchPattern := "%" + query + "%"

	// Search movies
	var movieResults []movies.Movie
	database.DB.Preload("Genres").
		Where("title ILIKE ? OR synopsis ILIKE ?", searchPattern, searchPattern).
		Limit(10).
		Find(&movieResults)

	// Search people
	var peopleResults []movies.Person
	database.DB.Where("name ILIKE ?", searchPattern).
		Limit(10).
		Find(&peopleResults)

	// Search genres
	var genreResults []movies.Genre
	database.DB.Where("name ILIKE ?", searchPattern).
		Limit(5).
		Find(&genreResults)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"movies": movieResults,
			"people": peopleResults,
			"genres": genreResults,
		},
	})
}

// GetStatsPublicHandler returns public statistics
func GetStatsPublicHandler(c *gin.Context) {
	var movieCount, genreCount, peopleCount int64

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

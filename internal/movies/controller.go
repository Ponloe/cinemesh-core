package movies

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Ponloe/cinemesh-core/internal/database"
	"github.com/gin-gonic/gin"
)

func ListMoviesHandler(c *gin.Context) {
	var movies []Movie
	if err := database.DB.Preload("Genres").Find(&movies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, movies)
}

func CreateMovieHandler(c *gin.Context) {
	var movie Movie
	if err := c.ShouldBindJSON(&movie); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := database.DB.Create(&movie).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, movie)
}

func ListGenresHandler(c *gin.Context) {
	var genres []Genre
	if err := database.DB.Find(&genres).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, genres)
}

// Admin Handlers for Movies
func ListMoviesAdminHandler(c *gin.Context) {
	sort := c.DefaultQuery("sort", "id")
	order := c.DefaultQuery("order", "desc")

	allowedSorts := map[string]bool{"id": true, "title": true, "release_date": true, "average_rating": true}
	if !allowedSorts[sort] {
		sort = "id"
	}
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	var movies []Movie
	query := database.DB.Preload("Genres").Order(sort + " " + order)
	if err := query.Find(&movies).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}
	c.HTML(http.StatusOK, "movies.html", gin.H{"movies": movies, "sort": sort, "order": order})
}

func NewMovieFormHandler(c *gin.Context) {
	var genres []Genre
	database.DB.Find(&genres)
	c.HTML(http.StatusOK, "movie_form.html", gin.H{
		"movie":          Movie{},
		"genres":         genres,
		"selectedGenres": make(map[uint]bool),
		"action":         "/admin/movies",
		"method":         "POST",
	})
}

func CreateMovieAdminHandler(c *gin.Context) {
	title := c.PostForm("title")
	slug := c.PostForm("slug")
	synopsis := c.PostForm("synopsis")
	posterURL := c.PostForm("poster_url")
	backdropURL := c.PostForm("backdrop_url")
	mpaaRating := c.PostForm("mpaa_rating")
	genreIDs := c.PostFormArray("genres")
	releaseDateStr := c.PostForm("release_date")
	durationStr := c.PostForm("duration_minutes")

	movie := Movie{
		Title:       title,
		Slug:        slug,
		Synopsis:    synopsis,
		PosterURL:   posterURL,
		BackdropURL: backdropURL,
		MPAARating:  mpaaRating,
	}

	if releaseDateStr != "" {
		releaseDate, err := time.Parse("2006-01-02", releaseDateStr)
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "invalid release date"})
			return
		}
		movie.ReleaseDate = &releaseDate
	}

	if durationStr != "" {
		duration, err := strconv.Atoi(durationStr)
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "invalid duration"})
			return
		}
		movie.DurationMinutes = &duration
	}
	// Associate genres
	for _, gidStr := range genreIDs {
		gid, _ := strconv.Atoi(gidStr)
		movie.Genres = append(movie.Genres, Genre{ID: uint(gid)})
	}

	if err := database.DB.Create(&movie).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}
	c.Redirect(http.StatusFound, "/admin/movies")
}

func EditMovieFormHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "invalid id"})
		return
	}

	var movie Movie
	if err := database.DB.Preload("Genres").First(&movie, uint(id)).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "movie not found"})
		return
	}

	var genres []Genre
	database.DB.Find(&genres)

	selectedGenres := make(map[uint]bool)
	for _, g := range movie.Genres {
		selectedGenres[g.ID] = true
	}

	// Add selectedGenres to the template context
	c.HTML(http.StatusOK, "movie_form.html", gin.H{
		"movie":          movie,
		"genres":         genres,
		"selectedGenres": selectedGenres,
		"action":         "/admin/movies/" + idStr,
		"method":         "POST",
	})
}

func UpdateMovieHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "invalid id"})
		return
	}

	var movie Movie
	if err := database.DB.Preload("Genres").First(&movie, uint(id)).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "movie not found"})
		return
	}

	// Update fields
	movie.Title = c.PostForm("title")
	movie.Slug = c.PostForm("slug")
	movie.Synopsis = c.PostForm("synopsis")
	movie.PosterURL = c.PostForm("poster_url")
	movie.BackdropURL = c.PostForm("backdrop_url")
	movie.MPAARating = c.PostForm("mpaa_rating")

	releaseDateStr := c.PostForm("release_date")
	if releaseDateStr != "" {
		releaseDate, err := time.Parse("2006-01-02", releaseDateStr)
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "invalid release date"})
			return
		}
		movie.ReleaseDate = &releaseDate
	} else {
		movie.ReleaseDate = nil
	}

	durationStr := c.PostForm("duration_minutes")
	if durationStr != "" {
		duration, err := strconv.Atoi(durationStr)
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "invalid duration"})
			return
		}
		movie.DurationMinutes = &duration
	} else {
		movie.DurationMinutes = nil
	}

	// Save basic movie fields first
	if err := database.DB.Save(&movie).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	// Update genres using Association Replace
	genreIDs := c.PostFormArray("genres")
	var genres []Genre
	for _, gidStr := range genreIDs {
		gid, _ := strconv.Atoi(gidStr)
		genres = append(genres, Genre{ID: uint(gid)})
	}

	// Replace the genres association
	if err := database.DB.Model(&movie).Association("Genres").Replace(genres); err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/admin/movies")
}
func DeleteMovieHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "invalid id"})
		return
	}

	if err := database.DB.Delete(&Movie{}, uint(id)).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}
	c.Redirect(http.StatusFound, "/admin/movies")
}

// Admin Handlers for Genres
func ListGenresAdminHandler(c *gin.Context) {
	var genres []Genre
	if err := database.DB.Order("name ASC").Find(&genres).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}
	c.HTML(http.StatusOK, "genres.html", gin.H{"genres": genres})
}

func NewGenreFormHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "genre_form.html", gin.H{"genre": Genre{}, "action": "/admin/genres", "method": "POST"})
}

func CreateGenreAdminHandler(c *gin.Context) {
	name := c.PostForm("name")
	if name == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "name is required"})
		return
	}

	genre := Genre{Name: name}
	if err := database.DB.Create(&genre).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}
	c.Redirect(http.StatusFound, "/admin/genres")
}

func EditGenreFormHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "invalid id"})
		return
	}

	var genre Genre
	if err := database.DB.First(&genre, uint(id)).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "genre not found"})
		return
	}

	c.HTML(http.StatusOK, "genre_form.html", gin.H{"genre": genre, "action": "/admin/genres/" + idStr, "method": "POST"})
}

func UpdateGenreHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "invalid id"})
		return
	}

	var genre Genre
	if err := database.DB.First(&genre, uint(id)).Error; err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "genre not found"})
		return
	}

	genre.Name = c.PostForm("name")
	if err := database.DB.Save(&genre).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}
	c.Redirect(http.StatusFound, "/admin/genres")
}

func DeleteGenreHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "invalid id"})
		return
	}

	if err := database.DB.Delete(&Genre{}, uint(id)).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}
	c.Redirect(http.StatusFound, "/admin/genres")
}

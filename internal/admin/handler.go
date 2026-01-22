package admin

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/Ponloe/cinemesh-core/internal/auth"
	"github.com/Ponloe/cinemesh-core/internal/database"
	"github.com/Ponloe/cinemesh-core/internal/movies"
	"github.com/Ponloe/cinemesh-core/internal/tmdb"
	"github.com/Ponloe/cinemesh-core/internal/users"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	tmdbClient  *tmdb.Client
	tmdbFetcher *tmdb.MovieFetcher
)

func InitializeTMDb() {
	log.Println("Initializing TMDb client...")
	tmdbConfig := tmdb.NewConfig()

	if tmdbConfig.APIKey == "" {
		log.Println("ERROR: TMDB_API_KEY is empty!")
	} else if len(tmdbConfig.APIKey) >= 8 {
		log.Printf("TMDb API Key loaded: %s...%s (length: %d)",
			tmdbConfig.APIKey[:4],
			tmdbConfig.APIKey[len(tmdbConfig.APIKey)-4:],
			len(tmdbConfig.APIKey))
	} else {
		log.Printf("WARNING: TMDB_API_KEY seems too short: %d characters", len(tmdbConfig.APIKey))
	}

	tmdbClient = tmdb.NewClient(tmdbConfig)
	tmdbFetcher = tmdb.NewMovieFetcher(tmdbClient)
	log.Println("TMDb client initialized successfully")
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
	log.Printf("TMDb search requested for query: %s", query)

	if query == "" {
		log.Println("ERROR: Empty query parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter required"})
		return
	}

	if tmdbClient == nil {
		log.Println("ERROR: TMDb client is nil")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "TMDb client not initialized",
			"hint":  "Check server logs and TMDB_API_KEY in .env",
		})
		return
	}

	log.Printf("Calling TMDb API to search for: %s", query)
	results, err := tmdbClient.SearchMovies(query)
	if err != nil {
		log.Printf("ERROR: TMDb search failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("TMDb API error: %v", err),
			"query": query,
		})
		return
	}

	log.Printf("TMDb search successful: found %d results", len(results.Results))
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

	log.Printf("Starting import for TMDb ID: %d", req.TMDbID)

	// Fetch movie using tmdbFetcher
	movie, err := tmdbFetcher.FetchMovieByTMDbID(req.TMDbID)
	if err != nil {
		log.Printf("ERROR: Failed to fetch movie from TMDb: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch from TMDb: " + err.Error()})
		return
	}

	log.Printf("Fetched movie: %s (TMDb ID: %d)", movie.Title, req.TMDbID)

	// Check if movie already exists
	var existing movies.Movie
	if err := database.DB.Where("slug = ?", movie.Slug).First(&existing).Error; err == nil {
		log.Printf("Movie already exists: %s (ID: %d)", existing.Title, existing.ID)
		c.JSON(http.StatusConflict, gin.H{
			"error":    "movie already exists",
			"movie_id": existing.ID,
		})
		return
	}

	// Start transaction with proper cleanup
	tx := database.DB.Begin()
	if tx.Error != nil {
		log.Printf("ERROR: Failed to begin transaction: %v", tx.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// Ensure rollback on panic or error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("PANIC during import, rolled back: %v", r)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "import failed due to panic"})
		}
	}()

	// Helper function to rollback and return error
	rollbackAndError := func(message string, err error) {
		tx.Rollback()
		log.Printf("ERROR: %s: %v", message, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": message + ": " + err.Error()})
	}

	// Handle genres
	var dbGenres []movies.Genre
	for _, tmdbGenre := range movie.Genres {
		var genre movies.Genre
		// Try to find existing genre by name
		if err := tx.Where("LOWER(name) = LOWER(?)", tmdbGenre.Name).First(&genre).Error; err == nil {
			log.Printf("Found existing genre: %s (ID: %d)", genre.Name, genre.ID)
			dbGenres = append(dbGenres, genre)
		} else if err == gorm.ErrRecordNotFound {
			// Create new genre with TMDb ID
			genreTMDbID := int(tmdbGenre.ID)
			genre = movies.Genre{
				Name:   tmdbGenre.Name,
				TMDbID: &genreTMDbID,
			}

			// Use upsert for genres too
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "name"}},
				DoNothing: true,
			}).Create(&genre).Error; err != nil {
				// If conflict, query again
				if tx.Where("LOWER(name) = LOWER(?)", tmdbGenre.Name).First(&genre).Error != nil {
					rollbackAndError("failed to create genre", err)
					return
				}
			}
			log.Printf("Created/found genre: %s (ID: %d)", genre.Name, genre.ID)
			dbGenres = append(dbGenres, genre)
		} else {
			rollbackAndError("failed to query genre", err)
			return
		}
	}
	movie.Genres = dbGenres

	// Create the movie
	log.Printf("Creating movie: %s with TMDb ID: %d", movie.Title, req.TMDbID)
	if err := tx.Create(movie).Error; err != nil {
		rollbackAndError("failed to save movie", err)
		return
	}
	log.Printf("Movie created successfully with ID: %d", movie.ID)

	// Import cast and crew
	log.Printf("Starting cast import for movie ID: %d", movie.ID)
	if err := importMovieCredits(tx, movie, req.TMDbID); err != nil {
		rollbackAndError("failed to import cast", err)
		return
	}
	log.Printf("Cast import completed")

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("ERROR: Failed to commit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit: " + err.Error()})
		return
	}

	database.DB.Preload("Genres").Preload("Cast.Person").First(movie, movie.ID)

	log.Printf("✅ Movie import completed successfully: %s (ID: %d)", movie.Title, movie.ID)
	c.JSON(http.StatusOK, gin.H{
		"message": "movie imported successfully with cast",
		"movie":   movie,
	})
}

func importMovieCredits(tx *gorm.DB, movie *movies.Movie, tmdbID int) error {
	log.Printf("Fetching credits for TMDb ID: %d", tmdbID)
	credits, err := tmdbClient.FetchMovieCredits(tmdbID)
	if err != nil {
		return fmt.Errorf("fetch credits: %w", err)
	}

	log.Printf("Found %d cast and %d crew members", len(credits.Cast), len(credits.Crew))

	// Import top 10 cast members
	for i, castMember := range credits.Cast {
		if i >= 10 {
			break
		}

		log.Printf("Processing cast member %d: %s (TMDb ID: %d)", i+1, castMember.Name, castMember.ID)
		person, err := getOrCreatePerson(tx, castMember.ID, castMember.Name, castMember.ProfilePath)
		if err != nil {
			return fmt.Errorf("get/create person %s: %w", castMember.Name, err)
		}

		order := castMember.Order
		moviePerson := movies.MoviePerson{
			MovieID:       movie.ID,
			PersonID:      person.ID,
			Role:          "Actor",
			CharacterName: castMember.Character,
			CastOrder:     &order,
		}

		// Use upsert to prevent duplicate key errors
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "movie_id"}, {Name: "person_id"}, {Name: "role"}},
			DoNothing: true,
		}).Create(&moviePerson).Error; err != nil {
			return fmt.Errorf("create movie_person for %s: %w", castMember.Name, err)
		}
		log.Printf("✓ Added cast: %s as %s", person.Name, castMember.Character)
	}

	// Import directors and key crew
	importedCrew := make(map[int]bool)
	for _, crewMember := range credits.Crew {
		if importedCrew[crewMember.ID] {
			continue
		}

		var role string
		switch crewMember.Job {
		case "Director":
			role = "Director"
		case "Screenplay", "Writer":
			role = "Writer"
		case "Producer":
			role = "Producer"
		default:
			continue
		}

		log.Printf("Processing crew: %s (%s, TMDb ID: %d)", crewMember.Name, role, crewMember.ID)
		person, err := getOrCreatePerson(tx, crewMember.ID, crewMember.Name, crewMember.ProfilePath)
		if err != nil {
			return fmt.Errorf("get/create crew person %s: %w", crewMember.Name, err)
		}

		moviePerson := movies.MoviePerson{
			MovieID:  movie.ID,
			PersonID: person.ID,
			Role:     role,
		}

		// Use upsert to prevent duplicate key errors
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "movie_id"}, {Name: "person_id"}, {Name: "role"}},
			DoNothing: true,
		}).Create(&moviePerson).Error; err != nil {
			return fmt.Errorf("create movie_person for crew %s: %w", crewMember.Name, err)
		}

		log.Printf("✓ Added crew: %s as %s", person.Name, role)
		importedCrew[crewMember.ID] = true
	}

	return nil
}

func getOrCreatePerson(tx *gorm.DB, tmdbID int, name string, profilePath string) (*movies.Person, error) {
	var person movies.Person
	err := tx.Unscoped().Where("tmdb_id = ?", tmdbID).First(&person).Error

	if err == nil {
		if person.DeletedAt.Valid {
			log.Printf("Restoring soft-deleted person: %s (ID: %d, TMDb ID: %d)", person.Name, person.ID, tmdbID)
			if err := tx.Model(&person).Update("deleted_at", nil).Error; err != nil {
				return nil, fmt.Errorf("restore person: %w", err)
			}
		} else {
			log.Printf("Found existing person: %s (ID: %d, TMDb ID: %d)", person.Name, person.ID, tmdbID)
		}
		return &person, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("query person tmdb_id=%d: %w", tmdbID, err)
	}

	log.Printf("Creating new person: %s (TMDb ID: %d)", name, tmdbID)
	person = movies.Person{
		Name:            name,
		ProfileImageURL: tmdbClient.GetFullImageURL(profilePath),
		TMDbID:          &tmdbID,
	}

	err = tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "tmdb_id"}},
		DoNothing: true, // If conflict, do nothing and retry query
	}).Create(&person).Error

	if err != nil {
		// If conflict happened, query again to get the existing record
		if tx.Unscoped().Where("tmdb_id = ?", tmdbID).First(&person).Error == nil {
			log.Printf("Person created by concurrent transaction: %s (ID: %d)", person.Name, person.ID)
			if person.DeletedAt.Valid {
				if err := tx.Model(&person).Update("deleted_at", nil).Error; err != nil {
					return nil, fmt.Errorf("restore concurrent person: %w", err)
				}
			}
			return &person, nil
		}
		return nil, fmt.Errorf("create person %s (tmdb_id=%d): %w", name, tmdbID, err)
	}

	log.Printf("Created person: %s (ID: %d, TMDb ID: %d)", person.Name, person.ID, tmdbID)
	return &person, nil
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

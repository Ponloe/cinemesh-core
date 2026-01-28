# Cinemesh-Core AI Agent Instructions

## Architecture Overview

**Cinemesh-Core** is a Go/Gin REST API that serves as the central metadata hub for a distributed movie platform. It exposes both **public APIs** (no auth) for external integrations and **protected admin APIs** (JWT auth) for content management.

```
┌─────────────────────────────────────────────────┐
│  Public API (/api/public/*)     │  Admin UI     │
│  - Movies, Genres, People       │  - TMDb Import│
│  - Search, Stats                │  - User Mgmt  │
│  - NO AUTH REQUIRED             │  - JWT Auth   │
└─────────────────────────────────────────────────┘
         ▼                                  ▼
┌─────────────────────────────────────────────────┐
│           Database Layer (PostgreSQL)            │
│  movies • genres • people • movie_people         │
└─────────────────────────────────────────────────┘
```

## Critical GORM Patterns

### 1. Column Name Override (CRITICAL)
GORM converts `TMDbID` → `tm_db_id` by default. **Always use explicit column tags:**

```go
TMDbID *int `gorm:"column:tmdb_id;uniqueIndex" json:"tmdb_id"`
```

### 2. Soft Deletes with Unique Indexes
Unique indexes MUST exclude soft-deleted rows:

```sql
CREATE UNIQUE INDEX idx_people_tmdb_id 
  ON people(tmdb_id) 
  WHERE tmdb_id IS NOT NULL AND deleted_at IS NULL;
```

### 3. Upsert Pattern (Transaction Safety)
Use `clause.OnConflict` to prevent duplicate key violations in concurrent scenarios:

```go
import "gorm.io/gorm/clause"

tx.Clauses(clause.OnConflict{
    Columns:   []clause.Column{{Name: "tmdb_id"}},
    DoNothing: true, // Then query again to get existing record
}).Create(&person)
```

## Project Structure

```
cmd/server/main.go           # Entry point, route registration
internal/
  ├── admin/                 # Admin handlers, TMDb import logic
  │   ├── handler.go         # ImportFromTMDbHandler, transaction management
  │   └── templates/         # Gin HTML templates (movies.html, api_docs.html)
  ├── api/                   # Public API handlers (no auth)
  │   └── public_controller.go
  ├── auth/                  # JWT middleware, token generation
  ├── database/              # GORM connection, migrations
  ├── movies/                # Models (Movie, Genre, Person, MoviePerson)
  │   ├── model.go           # All database models
  │   ├── controller.go      # Admin CRUD handlers
  │   └── people_controller.go
  ├── tmdb/                  # TMDb API client
  │   ├── client.go          # HTTP client, rate limiting
  │   └── movie_fetcher.go   # FetchMovieByTMDbID, FetchMovieCredits
  └── users/                 # User model, auth handlers
```

## Key Developer Workflows

### Running Locally
```bash
# Prerequisites: PostgreSQL running, .env configured
go run cmd/server/main.go

# Endpoints:
# http://localhost:8080              → Public API docs
# http://localhost:8080/admin        → Admin dashboard (requires JWT)
# http://localhost:8080/api/public/* → Public REST API
```

### Database Migrations
**Auto-migration runs on startup** via `database.Migrate()` in `main.go`:
```go
database.Migrate(&users.User{}, &movies.Movie{}, &movies.Genre{}, 
                 &movies.MovieGenre{}, &movies.Person{}, &movies.MoviePerson{})
```

To fix schema issues (column mismatches, indexes):
```bash
psql -U postgres -d cinemesh
ALTER TABLE people DROP COLUMN IF EXISTS tm_db_id; # Clean up GORM mistakes
\d people  # Verify schema
```

### TMDb Import Flow
1. User searches TMDb via `/admin/tmdb/search` (frontend)
2. POST `/admin/tmdb/import` with `{"tmdb_id": 27205}`
3. Backend:
   - Fetches movie details from TMDb API
   - Starts transaction (`tx := DB.Begin()`)
   - Creates/updates genres (upsert by name)
   - Creates movie with slug, poster URLs
   - Imports top 10 cast + key crew via `importMovieCredits()`
   - Uses `getOrCreatePerson()` with upsert pattern
   - Commits or rolls back entire transaction

**Critical**: Transaction errors cascade. If person import fails, entire movie import rolls back.

## Project-Specific Conventions

### 1. Custom Template Functions
Register custom Go template functions before `LoadHTMLGlob()`:
```go
r.SetFuncMap(template.FuncMap{
    "add": func(a, b int) int { return a + b },
})
r.LoadHTMLGlob("internal/admin/templates/*")
```

### 2. Pagination Pattern
All list endpoints use consistent pagination:
```go
page := c.DefaultQuery("page", "1")
limit := c.DefaultQuery("limit", "20")
if limit > 100 { limit = 20 } // Max 100 items
offset := (page - 1) * limit

query.Offset(offset).Limit(limit).Find(&results)
// Return pagination metadata in response
```

### 3. Slug Generation
Movies use URL-safe slugs via `gosimple/slug`:
```go
import "github.com/gosimple/slug"

movie.Slug = slug.Make(movie.Title) // "Inception" → "inception"
```

### 4. Error Handling in Transactions
Always use defer + recover for transaction cleanup:
```go
tx := database.DB.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
        c.JSON(500, gin.H{"error": "transaction failed"})
    }
}()

// Do work...

if err := tx.Commit().Error; err != nil {
    // No need to rollback, already failed
    c.JSON(500, gin.H{"error": err.Error()})
    return
}
```

## External Dependencies

- **TMDb API**: Movie metadata source (`TMDB_API_KEY` required)
  - Base URL: `https://api.themoviedb.org/3`
  - Rate limit: ~40 requests/10 seconds (handled in `client.go`)
  - Image URLs: `https://image.tmdb.org/t/p/w500{path}`

- **PostgreSQL**: Primary data store
  - Connection pooling: 25 max open/idle connections
  - GORM logger enabled for debugging SQL

- **JWT**: Auth tokens (`JWT_SECRET` required)
  - Expiry: 24 hours
  - Stored in cookies with `httpOnly=true`

## Common Pitfalls

1. **GORM Column Naming**: Always verify column names with `\d table_name` in psql
2. **Transaction Poisoning**: One error aborts all subsequent queries. Use separate transactions or savepoints
3. **Preload Depth**: Loading `Cast.Person` requires chaining: `.Preload("Cast.Person")`
4. **Template Functions**: Must register before `LoadHTMLGlob()`
5. **Unique Indexes**: Must exclude soft-deleted rows for Person/Genre/Movie TMDbID

## Testing Public API

```bash
# List movies with search
curl http://localhost:8080/api/public/movies?search=inception&limit=5

# Get specific movie (by ID or slug)
curl http://localhost:8080/api/public/movies/inception

# Global search
curl http://localhost:8080/api/public/search?q=nolan

# Stats
curl http://localhost:8080/api/public/stats
```

## Future Integrations

This core system is designed to be consumed by:
- **Forum App**: User discussions, ratings integration
- **Theater App**: Showtimes, ticketing (uses movie metadata)
- **Movie Finder**: Search aggregator across streaming platforms

All consume the public API at `/api/public/*` (no auth required).

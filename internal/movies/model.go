package movies

import (
	"time"
)

type Movie struct {
	ID              uint   `gorm:"primaryKey"`
	Title           string `gorm:"not null"`
	Slug            string `gorm:"unique;not null"`
	ReleaseDate     *time.Time
	DurationMinutes *int
	Synopsis        string
	PosterURL       string
	BackdropURL     string
	AverageRating   float64 `gorm:"type:decimal(3,2);default:0.0"`
	MPAARating      string
	CreatedAt       time.Time
	Genres          []Genre `gorm:"many2many:movie_genres;"`
}

type Genre struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"unique;not null"`
}

type MovieGenre struct {
	MovieID uint `gorm:"primaryKey"`
	GenreID uint `gorm:"primaryKey"`
}

package movies

import (
	"time"

	"gorm.io/gorm"
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
	TMDbID          *int `gorm:"column:tmdb_id;uniqueIndex" json:"tmdb_id"`
	CreatedAt       time.Time

	Genres []Genre       `gorm:"many2many:movie_genres;"`
	Cast   []MoviePerson `gorm:"foreignKey:MovieID;constraint:OnDelete:CASCADE"`
}

type Genre struct {
	ID     uint   `gorm:"primaryKey"`
	Name   string `gorm:"unique;not null"`
	TMDbID *int   `gorm:"column:tmdb_id;uniqueIndex" json:"tmdb_id"`
}

type MovieGenre struct {
	MovieID uint `gorm:"primaryKey"`
	GenreID uint `gorm:"primaryKey"`
}

type Person struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Name            string         `gorm:"size:100;not null;index" json:"name"`
	Biography       string         `gorm:"type:text" json:"biography"`
	BirthDate       *time.Time     `json:"birth_date"`
	ProfileImageURL string         `gorm:"size:255" json:"profile_image_url"`
	TMDbID          *int           `gorm:"column:tmdb_id;uniqueIndex" json:"tmdb_id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

type MoviePerson struct {
	MovieID       uint   `gorm:"primaryKey" json:"movie_id"`
	PersonID      uint   `gorm:"primaryKey" json:"person_id"`
	Role          string `gorm:"primaryKey;size:50;not null;index" json:"role"`
	CharacterName string `gorm:"size:100" json:"character_name,omitempty"`
	CastOrder     *int   `json:"cast_order,omitempty"`

	Movie  Movie  `gorm:"foreignKey:MovieID;constraint:OnDelete:CASCADE" json:"-"`
	Person Person `gorm:"foreignKey:PersonID;constraint:OnDelete:CASCADE" json:"person"`
}

func (MoviePerson) TableName() string {
	return "movie_people"
}

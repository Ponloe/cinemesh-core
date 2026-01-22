package tmdb

import (
	"time"
)

type MovieSearchResponse struct {
	Page         int           `json:"page"`
	Results      []MovieResult `json:"results"`
	TotalPages   int           `json:"total_pages"`
	TotalResults int           `json:"total_results"`
}

type MovieResult struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	OriginalTitle string  `json:"original_title"`
	Overview      string  `json:"overview"`
	PosterPath    string  `json:"poster_path"`
	BackdropPath  string  `json:"backdrop_path"`
	ReleaseDate   string  `json:"release_date"`
	VoteAverage   float64 `json:"vote_average"`
	Adult         bool    `json:"adult"`
	GenreIDs      []int   `json:"genre_ids"`
}

type MovieDetails struct {
	ID            int                  `json:"id"`
	Title         string               `json:"title"`
	OriginalTitle string               `json:"original_title"`
	Overview      string               `json:"overview"`
	PosterPath    string               `json:"poster_path"`
	BackdropPath  string               `json:"backdrop_path"`
	ReleaseDate   string               `json:"release_date"`
	Runtime       int                  `json:"runtime"`
	VoteAverage   float64              `json:"vote_average"`
	Status        string               `json:"status"`
	Tagline       string               `json:"tagline"`
	Genres        []Genre              `json:"genres"`
	ReleaseDates  ReleaseDatesResponse `json:"release_dates"`
}

type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ReleaseDatesResponse struct {
	Results []ReleaseDateCountry `json:"results"`
}

type ReleaseDateCountry struct {
	ISO31661     string        `json:"iso_3166_1"`
	ReleaseDates []ReleaseDate `json:"release_dates"`
}

type ReleaseDate struct {
	Certification string    `json:"certification"`
	ReleaseDate   time.Time `json:"release_date"`
	Type          int       `json:"type"`
}

type GenreListResponse struct {
	Genres []Genre `json:"genres"`
}

type TMDbCastMember struct {
	ID                 int    `json:"id"`
	Name               string `json:"name"`
	Character          string `json:"character"`
	Order              int    `json:"order"`
	ProfilePath        string `json:"profile_path"`
	KnownForDepartment string `json:"known_for_department"`
}

type TMDbCrewMember struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Job         string `json:"job"`
	Department  string `json:"department"`
	ProfilePath string `json:"profile_path"`
}

type TMDbCredits struct {
	Cast []TMDbCastMember `json:"cast"`
	Crew []TMDbCrewMember `json:"crew"`
}

type TMDbPersonDetail struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Biography   string  `json:"biography"`
	Birthday    *string `json:"birthday"`
	ProfilePath string  `json:"profile_path"`
}

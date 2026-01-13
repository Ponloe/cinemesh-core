package tmdb

import (
	"fmt"
	"time"

	"github.com/Ponloe/cinemesh-core/internal/movies"
	"github.com/gosimple/slug"
)

type MovieFetcher struct {
	client *Client
}

func NewMovieFetcher(client *Client) *MovieFetcher {
	return &MovieFetcher{client: client}
}

func (f *MovieFetcher) FetchMovieByTMDbID(tmdbID int) (*movies.Movie, error) {
	details, err := f.client.GetMovieDetails(tmdbID)
	if err != nil {
		return nil, fmt.Errorf("get movie details: %w", err)
	}

	return f.convertToMovie(details), nil
}

func (f *MovieFetcher) SearchAndConvert(query string) (*movies.Movie, error) {
	results, err := f.client.SearchMovies(query)
	if err != nil {
		return nil, fmt.Errorf("search movies: %w", err)
	}

	if len(results.Results) == 0 {
		return nil, fmt.Errorf("no results found for: %s", query)
	}

	return f.FetchMovieByTMDbID(results.Results[0].ID)
}

func (f *MovieFetcher) convertToMovie(details *MovieDetails) *movies.Movie {
	movie := &movies.Movie{
		Title:         details.Title,
		Slug:          slug.Make(details.Title),
		Synopsis:      details.Overview,
		PosterURL:     BuildPosterURL(details.PosterPath),
		BackdropURL:   BuildBackdropURL(details.BackdropPath),
		AverageRating: details.VoteAverage,
	}

	if details.ReleaseDate != "" {
		if releaseDate, err := time.Parse("2006-01-02", details.ReleaseDate); err == nil {
			movie.ReleaseDate = &releaseDate
		}
	}

	if details.Runtime > 0 {
		movie.DurationMinutes = &details.Runtime
	}

	movie.MPAARating = extractMPAARating(details)

	for _, g := range details.Genres {
		movie.Genres = append(movie.Genres, movies.Genre{
			ID:   uint(g.ID),
			Name: g.Name,
		})
	}

	return movie
}

func extractMPAARating(details *MovieDetails) string {
	for _, country := range details.ReleaseDates.Results {
		if country.ISO31661 == "US" {
			for _, rd := range country.ReleaseDates {
				if rd.Certification != "" {
					return rd.Certification
				}
			}
		}
	}
	return ""
}

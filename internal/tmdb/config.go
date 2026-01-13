package tmdb

import (
	"fmt"
	"os"
)

const (
	BaseURL      = "https://api.themoviedb.org/3"
	ImageBaseURL = "https://image.tmdb.org/t/p/"
)

const (
	SizePosterW92     = "w92"
	SizePosterW154    = "w154"
	SizePosterW185    = "w185"
	SizePosterW342    = "w342"
	SizePosterW500    = "w500"
	SizePosterW780    = "w780"
	SizeBackdropW300  = "w300"
	SizeBackdropW780  = "w780"
	SizeBackdropW1280 = "w1280"
	SizeOriginal      = "original"
)

type Config struct {
	APIKey string
}

func NewConfig() *Config {
	return &Config{
		APIKey: os.Getenv("TMDB_API_KEY"),
	}
}

func BuildImageURL(size, path string) string {
	if path == "" {
		return ""
	}
	return fmt.Sprintf("%s%s%s", ImageBaseURL, size, path)
}

func BuildPosterURL(path string) string {
	return BuildImageURL(SizePosterW500, path)
}

func BuildBackdropURL(path string) string {
	return BuildImageURL(SizeBackdropW1280, path)
}

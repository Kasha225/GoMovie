package config

import (
	"os"
)

func GetOMDBApiKey() string {
	apiKey := os.Getenv("OMDB_API_KEY")
	if apiKey == "" {
		panic("OMDB_API_KEY is not set")
	}
	return apiKey
}

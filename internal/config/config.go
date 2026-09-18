// Package config loads TYPESAFE_API_KEY and related settings from the environment.
package config

import (
	"fmt"
	"os"
	"strings"
)

const (
	DefaultModel   = "jev-latest"
	DefaultBaseURL = "https://api.typesafe.ai"
)

type Config struct {
	APIKey  string
	Model   string
	BaseURL string
}

func Load() (Config, error) {
	key := strings.TrimSpace(os.Getenv("TYPESAFE_API_KEY"))
	if key == "" {
		return Config{}, fmt.Errorf("TYPESAFE_API_KEY is not set")
	}
	model := os.Getenv("JEV_MCP_MODEL")
	if model == "" {
		model = DefaultModel
	}
	base := os.Getenv("TYPESAFE_BASE_URL")
	if base == "" {
		base = DefaultBaseURL
	}
	return Config{APIKey: key, Model: model, BaseURL: strings.TrimRight(base, "/")}, nil
}

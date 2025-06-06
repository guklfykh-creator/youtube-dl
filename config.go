package main

import (
	"fmt"
	"os"
)

type Config struct {
	BotToken      string
	GitHubToken   string
	RepoOwner     string
	RepoName      string
	DefaultBranch string
}

func LoadConfig() *Config {
	return &Config{
		BotToken:      os.Getenv("BOT_TOKEN"),
		GitHubToken:   os.Getenv("GH_TOKEN"),
		RepoOwner:     os.Getenv("GH_REPO_OWNER"),
		RepoName:      os.Getenv("GH_REPO_NAME"),
		DefaultBranch: envOr("GH_DEFAULT_BRANCH", "main"),
	}
}

func (c *Config) Validate() error {
	if c.BotToken == "" {
		return fmt.Errorf("BOT_TOKEN env is required")
	}
	if c.GitHubToken == "" {
		return fmt.Errorf("GH_TOKEN env is required")
	}
	if c.RepoOwner == "" {
		return fmt.Errorf("GH_REPO_OWNER env is required")
	}
	if c.RepoName == "" {
		return fmt.Errorf("GH_REPO_NAME env is required")
	}
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	BotToken      string
	GitHubToken   string
	RepoOwner     string
	RepoName      string
	WorkflowFile  string
	DefaultBranch string
}

func LoadConfig() *Config {
	_ = loadDotEnv(".env")

	return &Config{
		BotToken:      os.Getenv("BOT_TOKEN"),
		GitHubToken:   os.Getenv("GH_TOKEN"),
		RepoOwner:     os.Getenv("GH_REPO_OWNER"),
		RepoName:      os.Getenv("GH_REPO_NAME"),
		WorkflowFile:  envOr("GH_WORKFLOW_FILE", "download.yml"),
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
	if c.WorkflowFile == "" {
		return fmt.Errorf("GH_WORKFLOW_FILE env is required")
	}
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(stripInlineComment(value))
		value = strings.Trim(value, `"'`)

		if key == "" || os.Getenv(key) != "" {
			continue
		}
		_ = os.Setenv(key, value)
	}
	return scanner.Err()
}

func stripInlineComment(value string) string {
	inSingle := false
	inDouble := false

	for i, r := range value {
		switch r {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case '#':
			if !inSingle && !inDouble && (i == 0 || value[i-1] == ' ' || value[i-1] == '\t') {
				return strings.TrimSpace(value[:i])
			}
		}
	}
	return value
}

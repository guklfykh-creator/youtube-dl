package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type WorkflowPayload struct {
	Ref    string            `json:"ref"`
	Inputs map[string]string `json:"inputs"`
}

func TriggerWorkflow(cfg *Config, url, formatType, quality, chatID string) error {
	payload := WorkflowPayload{
		Ref: cfg.DefaultBranch,
		Inputs: map[string]string{
			"url":         url,
			"format_type": formatType,
			"quality":     quality,
			"chat_id":     chatID,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/dispatches",
		cfg.RepoOwner, cfg.RepoName)

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+cfg.GitHubToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "youtube-dl-bot")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("api request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 204 {
		return nil
	}

	respBody, _ := io.ReadAll(resp.Body)

	switch resp.StatusCode {
	case 401:
		return fmt.Errorf("unauthorized: check GH_TOKEN — response: %s", string(respBody))
	case 404:
		return fmt.Errorf("repo or workflow not found: check GH_REPO_OWNER/GH_REPO_NAME — response: %s", string(respBody))
	case 422:
		return fmt.Errorf("validation failed: workflow not found in repo or inputs mismatch — response: %s", string(respBody))
	default:
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}
}
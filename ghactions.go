package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"
	"strings"
	"time"

	"golang.org/x/crypto/nacl/box"
)

type WorkflowPayload struct {
	Ref    string            `json:"ref"`
	Inputs map[string]string `json:"inputs"`
}

func TriggerWorkflow(cfg *Config, url, formatType, quality, chatID, username string) error {
	client := &http.Client{Timeout: 30 * time.Second}

	if err := SyncWorkflowSecrets(cfg, client); err != nil {
		return fmt.Errorf("sync workflow secrets: %w", err)
	}

	inputs := map[string]string{
		"url":         url,
		"format_type": formatType,
		"quality":     quality,
		"chat_id":     chatID,
	}
	if username != "" {
		inputs["username"] = username
	}

	payload := WorkflowPayload{
		Ref:    cfg.DefaultBranch,
		Inputs: inputs,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/workflows/%s/dispatches",
		urlpkg.PathEscape(cfg.RepoOwner), urlpkg.PathEscape(cfg.RepoName), urlpkg.PathEscape(cfg.WorkflowFile))

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+cfg.GitHubToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "youtube-dl-bot")

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
		return fmt.Errorf("repo or workflow not found: check GH_REPO_OWNER/GH_REPO_NAME/GH_WORKFLOW_FILE — response: %s", string(respBody))
	case 422:
		return fmt.Errorf("validation failed: workflow not found in repo or inputs mismatch — response: %s", string(respBody))
	default:
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}
}

type repoPublicKey struct {
	KeyID string `json:"key_id"`
	Key   string `json:"key"`
}

type encryptedSecretPayload struct {
	EncryptedValue string `json:"encrypted_value"`
	KeyID          string `json:"key_id"`
}

func SyncWorkflowSecrets(cfg *Config, client *http.Client) error {
	required := map[string]string{
		"BOT_TOKEN":   cfg.BotToken,
		"TG_APP_ID":   cfg.TGAppID,
		"TG_APP_HASH": cfg.TGAppHash,
		"TG_SESSION":  cfg.TGSession,
	}

	for name, value := range required {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s env is required for GitHub Actions upload", name)
		}
	}

	publicKey, err := getRepoPublicKey(cfg, client)
	if err != nil {
		return err
	}

	for name, value := range required {
		if err := putRepoSecret(cfg, client, publicKey, name, value); err != nil {
			return err
		}
	}

	return nil
}

func getRepoPublicKey(cfg *Config, client *http.Client) (*repoPublicKey, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/secrets/public-key",
		urlpkg.PathEscape(cfg.RepoOwner), urlpkg.PathEscape(cfg.RepoName))

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create public key request: %w", err)
	}
	setGitHubHeaders(req, cfg)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("public key request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read public key response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("get repository public key failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var publicKey repoPublicKey
	if err := json.Unmarshal(respBody, &publicKey); err != nil {
		return nil, fmt.Errorf("decode repository public key: %w", err)
	}
	if publicKey.KeyID == "" || publicKey.Key == "" {
		return nil, fmt.Errorf("repository public key response is incomplete")
	}

	return &publicKey, nil
}

func putRepoSecret(cfg *Config, client *http.Client, publicKey *repoPublicKey, name, value string) error {
	encryptedValue, err := encryptGitHubSecret(publicKey.Key, value)
	if err != nil {
		return fmt.Errorf("encrypt %s: %w", name, err)
	}

	payload := encryptedSecretPayload{
		EncryptedValue: encryptedValue,
		KeyID:          publicKey.KeyID,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal %s secret payload: %w", name, err)
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/secrets/%s",
		urlpkg.PathEscape(cfg.RepoOwner), urlpkg.PathEscape(cfg.RepoName), urlpkg.PathEscape(name))

	req, err := http.NewRequest(http.MethodPut, apiURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create %s secret request: %w", name, err)
	}
	setGitHubHeaders(req, cfg)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%s secret request: %w", name, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read %s secret response: %w", name, err)
	}
	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusNoContent {
		return nil
	}

	return fmt.Errorf("put %s secret failed with status %d: %s", name, resp.StatusCode, string(respBody))
}

func encryptGitHubSecret(publicKeyB64, value string) (string, error) {
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		return "", fmt.Errorf("decode public key: %w", err)
	}
	if len(publicKeyBytes) != 32 {
		return "", fmt.Errorf("github public key length is %d, expected 32", len(publicKeyBytes))
	}

	var publicKey [32]byte
	copy(publicKey[:], publicKeyBytes)

	encrypted, err := box.SealAnonymous(nil, []byte(value), &publicKey, rand.Reader)
	if err != nil {
		return "", fmt.Errorf("seal secret: %w", err)
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func setGitHubHeaders(req *http.Request, cfg *Config) {
	req.Header.Set("Authorization", "Bearer "+cfg.GitHubToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "youtube-dl-bot")
}

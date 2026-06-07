package main

import (
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestSyncWorkflowSecretsRejectsInvalidYouTubeCookiesBase64(t *testing.T) {
	cfg := &Config{
		BotToken:     "bot-token",
		GitHubToken:  "github-token",
		RepoOwner:    "owner",
		RepoName:     "repo",
		TGAppID:      "12345",
		TGAppHash:    "hash",
		TGSession:    "session",
		YTCookiesB64: "not-valid-base64!",
	}
	client := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			t.Fatalf("unexpected HTTP request to %s", req.URL)
			return nil, nil
		}),
	}

	err := SyncWorkflowSecrets(cfg, client)
	if err == nil {
		t.Fatal("expected invalid YT_COOKIES_B64 error")
	}
	if !strings.Contains(err.Error(), "YT_COOKIES_B64 must be valid base64") {
		t.Fatalf("unexpected error: %v", err)
	}
}

package trackers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"time"
)

// ClaudeCodeTracker tracks usage for Claude Code
type ClaudeCodeTracker struct {
	toolName       string
	toolType       ToolType
	cachedUsage    *UsageResponse
	cacheFetchedAt time.Time
}

const usageCacheTTL = 5 * time.Second

type Credentials struct {
	ClaudeAiOauth struct {
		AccessToken      string   `json:"accessToken"`
		RefreshToken     string   `json:"refreshToken"`
		ExpiresAt        int64    `json:"expiresAt"`
		Scopes           []string `json:"scopes"`
		SubscriptionType string   `json:"subscriptionType"`
		RateLimitTier    string   `json:"rateLimitTier"`
	} `json:"claudeAiOauth"`
}

type UsageResponse struct {
	FiveHour UsageWindow `json:"five_hour"`
	SevenDay UsageWindow `json:"seven_day"`
}

type UsageWindow struct {
	Utilization float64 `json:"utilization"`
	ResetsAt    string  `json:"resets_at"`
}

// NewClaudeCodeTracker creates a new tracker for Claude Code
func NewClaudeCodeTracker() *ClaudeCodeTracker {
	return &ClaudeCodeTracker{
		toolName: "Claude Code",
		toolType: ClaudeCodeTool,
	}
}

// NewClaudeCodeTrackerWithLimit creates a tracker with a specific cost limit
func NewClaudeCodeTrackerWithLimit(costLimit float64) *ClaudeCodeTracker {
	// costLimit parameter kept for API compatibility but ignored
	// (Anthropic API returns utilization percentage directly)
	return NewClaudeCodeTracker()
}

func (t *ClaudeCodeTracker) GetAvailablePercentage() (float64, error) {
	usage, err := t.getUsage()
	if err != nil {
		return 0, err
	}

	available := 100.0 - usage.FiveHour.Utilization
	if available < 0 {
		available = 0
	}
	if available > 100 {
		available = 100
	}

	return available, nil
}

func (t *ClaudeCodeTracker) GetRemainingTime() (int, error) {
	usage, err := t.getUsage()
	if err != nil {
		return 0, err
	}

	resetTime, err := time.Parse(time.RFC3339Nano, usage.FiveHour.ResetsAt)
	if err != nil {
		return 0, fmt.Errorf("failed to parse five_hour resets_at: %w", err)
	}

	remaining := time.Until(resetTime)
	if remaining <= 0 {
		return 0, nil
	}

	return int(remaining.Minutes()), nil
}

func (t *ClaudeCodeTracker) GetTotalCost5hWindow() (float64, error) {
	return 0, nil
}

func (t *ClaudeCodeTracker) GetToolName() string {
	return t.toolName
}

func (t *ClaudeCodeTracker) GetToolType() ToolType {
	return t.toolType
}

func (t *ClaudeCodeTracker) IsAvailable() (bool, error) {
	available, err := t.GetAvailablePercentage()
	if err != nil {
		return false, err
	}
	return available >= AvailabilityThreshold, nil
}

func (t *ClaudeCodeTracker) getUsage() (*UsageResponse, error) {
	if t.cachedUsage != nil && time.Since(t.cacheFetchedAt) < usageCacheTTL {
		return t.cachedUsage, nil
	}

	token, err := getClaudeAccessToken()
	if err != nil {
		return nil, err
	}

	usage, err := fetchClaudeUsage(token)
	if err != nil {
		return nil, err
	}

	t.cachedUsage = usage
	t.cacheFetchedAt = time.Now()
	return usage, nil
}

func getClaudeAccessToken() (string, error) {
	creds, err := readCredentialsFromKeychain()
	if err != nil {
		return "", err
	}
	// Check if token is expired
	if isTokenExpired(creds.ClaudeAiOauth.ExpiresAt) {
		log.Printf("Access token expired, refreshing via Claude PTY")
		if err := refreshTokenViaPTY(); err != nil && err.Error() != "signal: killed" {
			log.Printf("Warning: failed to refresh token via PTY: %v", err)
		}

		// Re-read credentials from Keychain after refresh
		creds, err = readCredentialsFromKeychain()
		if err != nil {
			return "", fmt.Errorf("failed to read refreshed credentials: %w", err)
		}
	}

	return creds.ClaudeAiOauth.AccessToken, nil
}

func readCredentialsFromKeychain() (*Credentials, error) {
	cmd := exec.Command("security", "find-generic-password", "-s", "Claude Code-credentials", "-w")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error executing security command: %v", err)
		return nil, fmt.Errorf("failed to retrieve Claude Code credentials: %v", err)
	}
	var creds Credentials
	if err := json.Unmarshal(output, &creds); err != nil {
		log.Printf("Error parsing credentials JSON: %v", err)
		return nil, fmt.Errorf("failed to parse Claude Code credentials: %v", err)
	}

	return &creds, nil
}

func isTokenExpired(expiresAt int64) bool {
	expiresAtSeconds := expiresAt / 1000 // Convert milliseconds to seconds
	return time.Now().Unix() > expiresAtSeconds
}

func refreshTokenViaPTY() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "claude")
	return cmd.Run()
}

func fetchClaudeUsage(accessToken string) (*UsageResponse, error) {
	request, err := http.NewRequest("GET", "https://api.anthropic.com/api/oauth/usage", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build usage request: %w", err)
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	request.Header.Set("anthropic-beta", "oauth-2025-04-20")

	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("usage request failed: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read usage response: %w", err)
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("usage request failed: status %d: %s", response.StatusCode, string(body))
	}
	var usage UsageResponse
	if err := json.Unmarshal(body, &usage); err != nil {
		return nil, fmt.Errorf("failed to parse usage response: %w", err)
	}

	return &usage, nil
}

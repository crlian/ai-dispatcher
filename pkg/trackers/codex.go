package trackers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type CodexTracker struct {
	toolName       string
	toolType       ToolType
	cachedUsage    *ChatGPTUsageResponse
	cacheFetchedAt time.Time
	accessToken    string
	credentials    *CodexCredentials
}

// NewCodexTracker creates a new tracker for Codex
func NewCodexTracker() *CodexTracker {
	return &CodexTracker{
		toolName: "Codex",
		toolType: CodexTool,
	}
}

// GetAvailablePercentage returns the percentage of available capacity
func (t *CodexTracker) GetAvailablePercentage() (float64, error) {
	usage, err := t.getUsage()
	if err != nil {
		return 0, err
	}

	// Usa el primary_window (la ventana actual de 5 horas)
	usedPercent := usage.RateLimit.PrimaryWindow.UsedPercent
	available := 100.0 - usedPercent

	if available < 0 {
		available = 0
	}
	if available > 100 {
		available = 100
	}

	return available, nil
}

// GetRemainingTime returns remaining time in minutes
func (t *CodexTracker) GetRemainingTime() (int, error) {
	usage, err := t.getUsage()
	if err != nil {
		return 0, err
	}

	// Usa el reset_at timestamp (en segundos desde epoch)
	resetAt := time.Unix(usage.RateLimit.PrimaryWindow.ResetAt, 0)
	remaining := time.Until(resetAt)

	if remaining <= 0 {
		return 0, nil
	}

	return int(remaining.Minutes()), nil
}

// GetTotalCost5hWindow returns the total cost in the 5-hour window
// Note: ChatGPT API doesn't return cost information, only rate limit usage
func (t *CodexTracker) GetTotalCost5hWindow() (float64, error) {
	_, err := t.getUsage()
	if err != nil {
		return 0, err
	}

	// La API de ChatGPT no retorna información de costos
	// Solo rate limiting. Retornamos 0.
	return 0, nil
}

// IsAvailable returns true if tool has more than 5% capacity
func (t *CodexTracker) IsAvailable() (bool, error) {
	available, err := t.GetAvailablePercentage()
	if err != nil {
		return false, err
	}
	return available >= AvailabilityThreshold, nil
}

// GetToolName returns the tool name
func (t *CodexTracker) GetToolName() string {
	return t.toolName
}

// GetToolType returns the tool type
func (t *CodexTracker) GetToolType() ToolType {
	return t.toolType
}

// getUsage fetches usage data from ChatGPT API
func (t *CodexTracker) getUsage() (*ChatGPTUsageResponse, error) {
	// Retorna cached si aún es válido (5 segundos)
	if t.cachedUsage != nil && time.Since(t.cacheFetchedAt) < 5*time.Second {
		return t.cachedUsage, nil
	}

	// Obtiene access token (con refresh si es necesario)
	token, err := t.getAccessToken()
	if err != nil {
		return nil, err
	}

	// Fetch usage desde ChatGPT API
	usage, err := t.fetchCodexUsage(token)
	if err != nil {
		return nil, err
	}

	t.cachedUsage = usage
	t.cacheFetchedAt = time.Now()
	return usage, nil
}

// getAccessToken obtiene el access token del archivo ~/.codex/auth.json
// y hace refresh si last_refresh es más viejo que 8 días
func (t *CodexTracker) getAccessToken() (string, error) {
	// Si ya tenemos el token en cache, retorna
	if t.accessToken != "" {
		return t.accessToken, nil
	}

	// Lee archivo de credenciales
	creds, err := t.readCodexCredentials()
	if err != nil {
		return "", err
	}

	// Verifica si necesita refresh (> 8 días)
	lastRefresh, err := time.Parse(time.RFC3339, creds.LastRefresh)
	if err != nil {
		return "", fmt.Errorf("failed to parse last_refresh: %w", err)
	}

	eightyDaysAgo := time.Now().AddDate(0, 0, -8)
	if lastRefresh.Before(eightyDaysAgo) {
		log.Printf("Codex token older than 8 days, refreshing...")
		newToken, err := t.refreshAccessToken(creds.Tokens.RefreshToken)
		if err != nil {
			return "", fmt.Errorf("failed to refresh token: %w", err)
		}
		t.accessToken = newToken
		return newToken, nil
	}

	t.accessToken = creds.Tokens.AccessToken
	return creds.Tokens.AccessToken, nil
}

// readCodexCredentials lee el archivo ~/.codex/auth.json
func (t *CodexTracker) readCodexCredentials() (*CodexCredentials, error) {
	// Expande ~
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	authPath := filepath.Join(homeDir, ".codex", "auth.json")

	// Lee archivo
	data, err := os.ReadFile(authPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Codex credentials from %s: %w", authPath, err)
	}

	// Parsea JSON
	var creds CodexCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse Codex credentials JSON: %w", err)
	}

	// Valida que tenga access token
	if creds.Tokens.AccessToken == "" {
		return nil, fmt.Errorf("codex credentials missing access_token")
	}

	return &creds, nil
}

// refreshAccessToken usa el refresh_token para obtener uno nuevo
func (t *CodexTracker) refreshAccessToken(refreshToken string) (string, error) {
	// OpenAI token refresh endpoint
	url := "https://auth.openai.com/oauth/token"

	payload := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
	}

	// Crea request
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal refresh request: %w", err)
	}

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create refresh request: %w", err)
	}

	// Usa jsonPayload en el body del request
	req.Body = io.NopCloser(bytes.NewReader(jsonPayload))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("refresh token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read refresh response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("refresh token failed: status %d: %s", resp.StatusCode, string(body))
	}

	var refreshResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &refreshResp); err != nil {
		return "", fmt.Errorf("failed to parse refresh response: %w", err)
	}

	if refreshResp.AccessToken == "" {
		return "", fmt.Errorf("refresh response missing access_token")
	}

	return refreshResp.AccessToken, nil
}

// fetchCodexUsage obtiene datos de uso desde la API de ChatGPT
func (t *CodexTracker) fetchCodexUsage(accessToken string) (*ChatGPTUsageResponse, error) {
	// Endpoint de ChatGPT para obtener información de uso
	url := "https://chatgpt.com/backend-api/wham/usage"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create usage request: %w", err)
	}

	// Headers necesarios
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("User-Agent", "ai-dispatcher")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("usage request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read usage response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("usage request failed: status %d: %s", resp.StatusCode, string(body))
	}

	var usage ChatGPTUsageResponse
	if err := json.Unmarshal(body, &usage); err != nil {
		return nil, fmt.Errorf("failed to parse usage response: %w", err)
	}

	return &usage, nil
}

package trackers

// CodexCredentials representa las credenciales almacenadas en ~/.codex/auth.json
type CodexCredentials struct {
	OpenAIAPIKey string `json:"OPENAI_API_KEY"`
	Tokens       struct {
		IDToken      string `json:"id_token"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		AccountID    string `json:"account_id"`
	} `json:"tokens"`
	LastRefresh string `json:"last_refresh"`
}

// ChatGPTUsageResponse representa la respuesta de la API de uso de ChatGPT
// Endpoint: https://chatgpt.com/backend-api/wham/usage
type ChatGPTUsageResponse struct {
	PlanType string `json:"plan_type"`
	RateLimit struct {
		Allowed      bool   `json:"allowed"`
		LimitReached bool   `json:"limit_reached"`
		PrimaryWindow struct {
			UsedPercent        float64 `json:"used_percent"`
			LimitWindowSeconds int     `json:"limit_window_seconds"`
			ResetAfterSeconds  int     `json:"reset_after_seconds"`
			ResetAt            int64   `json:"reset_at"`
		} `json:"primary_window"`
		SecondaryWindow *struct {
			UsedPercent        float64 `json:"used_percent"`
			LimitWindowSeconds int     `json:"limit_window_seconds"`
			ResetAfterSeconds  int     `json:"reset_after_seconds"`
			ResetAt            int64   `json:"reset_at"`
		} `json:"secondary_window"`
	} `json:"rate_limit"`
	Credits struct {
		HasCredits       bool   `json:"has_credits"`
		Unlimited        bool   `json:"unlimited"`
		Balance          string `json:"balance"`
		ApproxLocalMessages []int `json:"approx_local_messages"`
		ApproxCloudMessages []int `json:"approx_cloud_messages"`
	} `json:"credits"`
}

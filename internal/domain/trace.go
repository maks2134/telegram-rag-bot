package domain

// Step is one visible stage of the RAG pipeline (for the web Trace UI).
type Step struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // ok | skip | error
	Detail  string `json:"detail"`
	Ms      int64  `json:"ms,omitempty"`
	Payload any    `json:"payload,omitempty"`
}

// Result is the full ask outcome for Telegram and the web UI.
type Result struct {
	Answer    string   `json:"answer"`
	Sources   []Source `json:"sources"`
	Refused   bool     `json:"refused"`
	Model     string   `json:"model"`
	Trace     []Step   `json:"trace"`
	LatencyMs int64    `json:"latency_ms"`
}

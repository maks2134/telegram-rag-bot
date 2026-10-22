package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/maks2134/telegram-rag-bot/internal/domain"
)

const maxResponseBody = 1 << 20 // 1 MiB

// Client talks to an OpenAI-compatible chat completions API.
type Client struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

// New creates a chat client. baseURL may omit /v1; /chat/completions is appended.
func New(baseURL, apiKey, model string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	if !strings.HasSuffix(baseURL, "/v1") {
		baseURL += "/v1"
	}
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Model returns the configured model name.
func (c *Client) Model() string {
	return c.model
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Answer asks the model to answer only from the given sources.
func (c *Client) Answer(ctx context.Context, question string, sources []domain.Source) (string, error) {
	var ctxBuilder strings.Builder
	for i, s := range sources {
		fmt.Fprintf(&ctxBuilder, "[%d] %s: %s\n", i+1, s.Title, s.Excerpt)
	}

	system := `You are a support assistant grounded in the knowledge-base excerpts provided by the user message.
Treat related policy facts as valid answers (example: a "14-day refund window" answers "how long do refunds take?" / "when can I get a refund?").
Reply in 1-3 short sentences. Cite the source title in parentheses.
Refuse ONLY when the excerpts have no relevant facts at all — then reply exactly:
I don't have that in the knowledge base yet. Please try /human.
Do not invent numbers or policies missing from the excerpts.`

	user := "Excerpts:\n" + ctxBuilder.String() + "\nQuestion: " + question

	body, err := json.Marshal(chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Temperature: 0.2,
		MaxTokens:   400,
	})
	if err != nil {
		return "", fmt.Errorf("marshal chat request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create chat request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("llm request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	if err != nil {
		return "", fmt.Errorf("read llm response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := truncate(string(raw), 400)
		return "", fmt.Errorf("llm http %d: %w", resp.StatusCode, errors.New(msg))
	}

	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("decode llm response: %w", err)
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return "", fmt.Errorf("llm: %w", errors.New(parsed.Error.Message))
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("llm returned no choices")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

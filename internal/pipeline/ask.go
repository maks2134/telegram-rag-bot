package pipeline

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/maks2134/telegram-rag-bot/internal/llm"
	"github.com/maks2134/telegram-rag-bot/internal/store"
)

const topK = 3

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
	Answer    string         `json:"answer"`
	Sources   []store.Source `json:"sources"`
	Refused   bool           `json:"refused"`
	Model     string         `json:"model"`
	Trace     []Step         `json:"trace"`
	LatencyMs int64          `json:"latency_ms"`
}

// Asker runs retrieve → LLM grounded answer.
type Asker struct {
	Store *store.Store
	LLM   *llm.Client
}

const refuseMsg = "I don't have that in the knowledge base yet. Please try /human."

// Ask runs the RAG pipeline for a question.
func (a *Asker) Ask(ctx context.Context, question string) Result {
	start := time.Now()
	question = strings.TrimSpace(question)
	res := Result{
		Model: a.LLM.Model(),
		Trace: make([]Step, 0, 4),
	}

	if question == "" {
		res.Answer = "Send a question about your docs, or type /human."
		res.Refused = true
		res.Trace = append(res.Trace, Step{Name: "validate", Status: "error", Detail: "empty question"})
		res.LatencyMs = time.Since(start).Milliseconds()
		return res
	}

	t0 := time.Now()
	sources := a.Store.Retrieve(question, topK)
	res.Sources = sources
	res.Trace = append(res.Trace, Step{
		Name:   "retrieve",
		Status: statusOK(len(sources) > 0),
		Detail: fmt.Sprintf("top-%d keyword chunks", topK),
		Ms:     time.Since(t0).Milliseconds(),
		Payload: map[string]any{
			"count":   len(sources),
			"sources": sources,
		},
	})

	if len(sources) == 0 {
		res.Answer = refuseMsg
		res.Refused = true
		res.Trace = append(res.Trace,
			Step{Name: "prompt", Status: "skip", Detail: "no sources — LLM skipped"},
			Step{Name: "llm", Status: "skip", Detail: "no sources — LLM skipped"},
			Step{Name: "answer", Status: "ok", Detail: "refusal"},
		)
		res.LatencyMs = time.Since(start).Milliseconds()
		return res
	}

	var preview strings.Builder
	for i, s := range sources {
		fmt.Fprintf(&preview, "%d. %s (score %d)\n", i+1, s.Title, s.Score)
	}
	res.Trace = append(res.Trace, Step{
		Name:    "prompt",
		Status:  "ok",
		Detail:  "grounded system + excerpts",
		Payload: map[string]any{"context_preview": strings.TrimSpace(preview.String())},
	})

	t1 := time.Now()
	answer, err := a.LLM.Answer(ctx, question, sources)
	llmMs := time.Since(t1).Milliseconds()
	if err != nil {
		res.Answer = "Sorry — the model failed. Please try again or /human."
		res.Refused = true
		res.Trace = append(res.Trace, Step{
			Name:   "llm",
			Status: "error",
			Detail: err.Error(),
			Ms:     llmMs,
		})
		res.LatencyMs = time.Since(start).Milliseconds()
		return res
	}

	res.Trace = append(res.Trace, Step{
		Name:   "llm",
		Status: "ok",
		Detail: a.LLM.Model(),
		Ms:     llmMs,
	})

	res.Answer = answer
	res.Refused = strings.Contains(strings.ToLower(answer), "don't have that in the knowledge base")
	res.Trace = append(res.Trace, Step{
		Name:   "answer",
		Status: "ok",
		Detail: fmt.Sprintf("refused=%v", res.Refused),
	})
	res.LatencyMs = time.Since(start).Milliseconds()
	return res
}

// FormatTelegram builds a Telegram message with sources footer.
func FormatTelegram(res Result) string {
	var b strings.Builder
	b.WriteString(res.Answer)
	if len(res.Sources) > 0 && !res.Refused {
		b.WriteString("\n\nSources:")
		seen := map[string]bool{}
		for _, s := range res.Sources {
			if seen[s.Title] {
				continue
			}
			seen[s.Title] = true
			b.WriteString("\n• ")
			b.WriteString(s.Title)
		}
	}
	return b.String()
}

func statusOK(ok bool) string {
	if ok {
		return "ok"
	}
	return "skip"
}

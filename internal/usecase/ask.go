package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/maks2134/telegram-rag-bot/internal/domain"
)

const topK = 3

const refuseMsg = "I don't have that in the knowledge base yet. Please try /human."

const maxTraceDetailLen = 400

// Asker runs retrieve → LLM grounded answer.
type Asker struct {
	Retriever Retriever
	Answerer  Answerer
}

// Ask runs the RAG pipeline for a question.
func (a *Asker) Ask(ctx context.Context, question string) domain.Result {
	start := time.Now()
	question = strings.TrimSpace(question)
	res := domain.Result{
		Trace:   make([]domain.Step, 0, 4),
		Sources: []domain.Source{},
	}

	if a == nil || a.Retriever == nil || a.Answerer == nil {
		res.Answer = "Sorry — the assistant is not configured. Please try /human."
		res.Refused = true
		res.Trace = append(res.Trace, domain.Step{Name: "validate", Status: "error", Detail: "missing retriever or answerer"})
		res.LatencyMs = time.Since(start).Milliseconds()
		return res
	}

	res.Model = a.Answerer.Model()

	if question == "" {
		res.Answer = "Send a question about your docs, or type /human."
		res.Refused = true
		res.Trace = append(res.Trace, domain.Step{Name: "validate", Status: "error", Detail: "empty question"})
		res.LatencyMs = time.Since(start).Milliseconds()
		return res
	}

	t0 := time.Now()
	sources := a.Retriever.Retrieve(question, topK)
	res.Sources = sources
	res.Trace = append(res.Trace, domain.Step{
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
			domain.Step{Name: "prompt", Status: "skip", Detail: "no sources — LLM skipped"},
			domain.Step{Name: "llm", Status: "skip", Detail: "no sources — LLM skipped"},
			domain.Step{Name: "answer", Status: "ok", Detail: "refusal"},
		)
		res.LatencyMs = time.Since(start).Milliseconds()
		return res
	}

	var preview strings.Builder
	for i, s := range sources {
		fmt.Fprintf(&preview, "%d. %s (score %d)\n", i+1, s.Title, s.Score)
	}
	res.Trace = append(res.Trace, domain.Step{
		Name:    "prompt",
		Status:  "ok",
		Detail:  "grounded system + excerpts",
		Payload: map[string]any{"context_preview": strings.TrimSpace(preview.String())},
	})

	t1 := time.Now()
	answer, err := a.Answerer.Answer(ctx, question, sources)
	llmMs := time.Since(t1).Milliseconds()
	if err != nil {
		res.Answer = "Sorry — the model failed. Please try again or /human."
		res.Refused = true
		res.Trace = append(res.Trace, domain.Step{
			Name:   "llm",
			Status: "error",
			Detail: truncateDetail(err.Error(), maxTraceDetailLen),
			Ms:     llmMs,
		})
		res.LatencyMs = time.Since(start).Milliseconds()
		return res
	}

	res.Trace = append(res.Trace, domain.Step{
		Name:   "llm",
		Status: "ok",
		Detail: a.Answerer.Model(),
		Ms:     llmMs,
	})

	res.Answer = answer
	res.Refused = strings.Contains(answer, refuseMsg)
	res.Trace = append(res.Trace, domain.Step{
		Name:   "answer",
		Status: "ok",
		Detail: fmt.Sprintf("refused=%v", res.Refused),
	})
	res.LatencyMs = time.Since(start).Milliseconds()
	return res
}

func statusOK(ok bool) string {
	if ok {
		return "ok"
	}
	return "skip"
}

func truncateDetail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

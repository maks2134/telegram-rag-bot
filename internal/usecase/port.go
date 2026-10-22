package usecase

import (
	"context"

	"github.com/maks2134/telegram-rag-bot/internal/domain"
)

// Retriever returns top-k keyword-matched sources for a question.
type Retriever interface {
	Retrieve(question string, k int) []domain.Source
}

// Answerer produces a grounded answer from retrieved sources.
type Answerer interface {
	Answer(ctx context.Context, question string, sources []domain.Source) (string, error)
	Model() string
}

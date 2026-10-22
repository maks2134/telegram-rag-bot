package usecase

import (
	"context"
	"testing"

	"github.com/maks2134/telegram-rag-bot/internal/domain"
)

type stubRetriever struct {
	sources []domain.Source
}

func (s stubRetriever) Retrieve(string, int) []domain.Source {
	return s.sources
}

type stubAnswerer struct {
	answer string
	err    error
	model  string
}

func (s stubAnswerer) Answer(context.Context, string, []domain.Source) (string, error) {
	return s.answer, s.err
}

func (s stubAnswerer) Model() string {
	if s.model == "" {
		return "stub"
	}
	return s.model
}

func TestAskEmptyQuestion(t *testing.T) {
	a := &Asker{
		Retriever: stubRetriever{},
		Answerer:  stubAnswerer{model: "m"},
	}

	res := a.Ask(context.Background(), "   ")
	if !res.Refused {
		t.Fatal("expected refused")
	}
	if res.Answer == "" {
		t.Fatal("expected answer")
	}
}

func TestAskNilGuard(t *testing.T) {
	var a *Asker
	res := a.Ask(context.Background(), "hello")
	if !res.Refused || res.Answer == "" {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestAskWithSources(t *testing.T) {
	a := &Asker{
		Retriever: stubRetriever{sources: []domain.Source{
			{Title: "Refund Policy", Excerpt: "within 14 days", Score: 3},
		}},
		Answerer: stubAnswerer{answer: "Refunds take 14 days (Refund Policy).", model: "stub"},
	}
	res := a.Ask(context.Background(), "How long do refunds take?")
	if res.Refused {
		t.Fatal("expected grounded answer")
	}
	if res.Answer == "" || len(res.Sources) != 1 {
		t.Fatalf("unexpected: %+v", res)
	}
}

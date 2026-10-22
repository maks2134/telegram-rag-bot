package telegram

import (
	"strings"
	"testing"

	"github.com/maks2134/telegram-rag-bot/internal/domain"
)

func TestFormatTelegramUniqueSources(t *testing.T) {
	res := domain.Result{
		Answer:  "Refunds take 14 days.",
		Refused: false,
		Sources: []domain.Source{
			{Title: "Refunds", Excerpt: "a", Score: 2},
			{Title: "Refunds", Excerpt: "b", Score: 1},
			{Title: "SLA", Excerpt: "c", Score: 1},
		},
	}
	body := FormatTelegram(res)
	if strings.Count(body, "\n• Refunds") != 1 {
		t.Fatalf("expected one Refunds bullet, body=%q", body)
	}
	if !strings.Contains(body, "SLA") {
		t.Fatalf("missing SLA: %q", body)
	}
}

func TestFormatTelegramNilSources(t *testing.T) {
	res := domain.Result{Answer: "ok", Refused: true, Sources: nil}
	body := FormatTelegram(res)
	if body != "ok" {
		t.Fatalf("body=%q", body)
	}
}

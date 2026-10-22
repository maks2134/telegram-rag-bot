package keyword

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/maks2134/telegram-rag-bot/internal/domain"
)

func TestLoadDirAndRetrieve(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "refunds.md")
	content := "# Refund Policy\n\nCustomers can request a refund within 14 days of purchase.\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	s := New()
	if err := s.LoadDir(dir); err != nil {
		t.Fatal(err)
	}
	if len(s.Docs()) != 1 {
		t.Fatalf("docs=%d", len(s.Docs()))
	}

	sources := s.Retrieve("How do refunds work?", 3)
	if len(sources) == 0 {
		t.Fatal("expected keyword hits for refunds")
	}
	if sources[0].Title != "Refund Policy" {
		t.Fatalf("title=%q", sources[0].Title)
	}
}

func TestSplitSentencesTrimMarkdownHeading(t *testing.T) {
	got := splitSentences("## Refund Policy\n\nCustomers may refund within 14 days.")
	if len(got) == 0 {
		t.Fatal("expected chunks")
	}
	if got[0] == "Refund Policy" || got[0] == "## Refund Policy" {
		// first chunk may include heading line + body depending on split; heading prefix must be gone
	}
	for _, part := range got {
		if strings.HasPrefix(part, "# ") || strings.HasPrefix(part, "## ") {
			t.Fatalf("chunk still has heading prefix: %q", part)
		}
	}

	parts := splitSentences("# Title only")
	for _, p := range parts {
		if p == "itle only" {
			t.Fatalf("cutset mangled chunk: %q", p)
		}
	}
}

func TestRetrieveRanking(t *testing.T) {
	s := New()
	s.Add(domain.Document{
		ID:      "a",
		Title:   "Refunds",
		Content: "Refunds are processed in 14 days.",
	})
	s.Add(domain.Document{
		ID:      "b",
		Title:   "Onboarding",
		Content: "Onboarding takes one week for new teams.",
	})

	sources := s.Retrieve("refund processing time", 2)
	if len(sources) < 1 {
		t.Fatal("expected hits")
	}
	if sources[0].Title != "Refunds" {
		t.Fatalf("best match title=%q score=%d", sources[0].Title, sources[0].Score)
	}
	if sources[0].Score < sources[len(sources)-1].Score {
		t.Fatalf("sources not sorted by score desc: %v", sources)
	}
}

func TestRetrieveEmptyReturnsNonNil(t *testing.T) {
	s := New()
	s.Add(domain.Document{ID: "x", Title: "T", Content: "Unrelated topic only."})
	out := s.Retrieve("zzzznomatch", 3)
	if out == nil {
		t.Fatal("expected non-nil empty slice")
	}
	if len(out) != 0 {
		t.Fatalf("len=%d", len(out))
	}
}

func TestLoadDirEmpty(t *testing.T) {
	dir := t.TempDir()
	s := New()
	err := s.LoadDir(dir)
	if err == nil {
		t.Fatal("expected error for empty docs dir")
	}
}

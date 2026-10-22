package store

import (
	"os"
	"path/filepath"
	"testing"
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

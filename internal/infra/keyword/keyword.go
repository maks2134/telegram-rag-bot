package keyword

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/google/uuid"

	"github.com/maks2134/telegram-rag-bot/internal/domain"
)

type chunk struct {
	DocID   string
	Title   string
	Content string
	Tokens  []string
}

// Index is an in-memory keyword retrieval index.
type Index struct {
	mu     sync.RWMutex
	docs   []domain.Document
	chunks []chunk
}

// New returns an empty index.
func New() *Index {
	return &Index{}
}

// Docs returns a copy of loaded documents.
func (s *Index) Docs() []domain.Document {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Document, len(s.docs))
	copy(out, s.docs)
	return out
}

// LoadDir indexes all *.md files under dir.
func (s *Index) LoadDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read docs dir: %w", err)
	}

	loaded := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		title := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
		content := strings.TrimSpace(string(raw))
		if content == "" {
			continue
		}
		if h := firstHeading(content); h != "" {
			title = h
		}
		s.Add(domain.Document{
			ID:      uuid.NewString(),
			Title:   title,
			Content: content,
		})
		loaded++
	}
	if loaded == 0 {
		return fmt.Errorf("no markdown files found in %s", dir)
	}
	return nil
}

// Add chunks and indexes a document.
func (s *Index) Add(d domain.Document) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.docs = append(s.docs, d)
	for _, sentence := range splitSentences(d.Content) {
		s.chunks = append(s.chunks, chunk{
			DocID:   d.ID,
			Title:   d.Title,
			Content: sentence,
			Tokens:  tokenize(sentence + " " + d.Title),
		})
	}
}

func splitSentences(content string) []string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, ".\n", ". ")
	var out []string
	for _, part := range strings.Split(content, ". ") {
		part = trimMarkdownHeading(strings.TrimSpace(part))
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func trimMarkdownHeading(s string) string {
	s = strings.TrimSpace(s)
	for strings.HasPrefix(s, "## ") {
		s = strings.TrimSpace(strings.TrimPrefix(s, "## "))
	}
	for strings.HasPrefix(s, "# ") {
		s = strings.TrimSpace(strings.TrimPrefix(s, "# "))
	}
	return s
}

// Retrieve returns the top-k keyword-matched sources.
func (s *Index) Retrieve(question string, k int) []domain.Source {
	qTokens := tokenize(question)
	qSet := make(map[string]struct{}, len(qTokens))
	for _, qt := range qTokens {
		qSet[qt] = struct{}{}
	}

	type scored struct {
		src   domain.Source
		score int
	}
	var ranked []scored

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, c := range s.chunks {
		score := 0
		for _, ct := range c.Tokens {
			if _, ok := qSet[ct]; ok {
				score++
			}
		}
		if score > 0 {
			ranked = append(ranked, scored{
				src: domain.Source{
					Title:   c.Title,
					Excerpt: c.Content,
					Score:   score,
				},
				score: score,
			})
		}
	}

	slices.SortFunc(ranked, func(a, b scored) int {
		return cmp.Compare(b.score, a.score)
	})

	if k > len(ranked) {
		k = len(ranked)
	}
	out := make([]domain.Source, 0, k)
	for i := 0; i < k; i++ {
		out = append(out, ranked[i].src)
	}
	return out
}

func firstHeading(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}

func stem(p string) string {
	if len(p) > 4 && strings.HasSuffix(p, "s") && !strings.HasSuffix(p, "ss") {
		return p[:len(p)-1]
	}
	return p
}

func tokenize(s string) []string {
	s = strings.ToLower(s)
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9')
	})
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if len(p) > 2 {
			out = append(out, stem(p))
		}
	}
	return out
}

package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"
)

// Document is a loaded knowledge-base file.
type Document struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// Source is a retrieved excerpt with a relevance score.
type Source struct {
	Title   string `json:"title"`
	Excerpt string `json:"excerpt"`
	Score   int    `json:"score"`
}

type chunk struct {
	DocID   string
	Title   string
	Content string
	Tokens  []string
}

// Store is an in-memory keyword retrieval index.
type Store struct {
	mu     sync.RWMutex
	docs   []Document
	chunks []chunk
}

// New returns an empty store.
func New() *Store {
	return &Store{}
}

// Docs returns a copy of loaded documents.
func (s *Store) Docs() []Document {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Document, len(s.docs))
	copy(out, s.docs)
	return out
}

// LoadDir indexes all *.md files under dir.
func (s *Store) LoadDir(dir string) error {
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
		s.Add(Document{
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
func (s *Store) Add(d Document) {
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
	// Normalize sentence boundaries so ".\n" still chunks.
	content = strings.ReplaceAll(content, ".\n", ". ")
	var out []string
	for _, part := range strings.Split(content, ". ") {
		part = strings.TrimSpace(part)
		part = strings.TrimLeft(part, "# ")
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

// Retrieve returns the top-k keyword-matched sources.
func (s *Store) Retrieve(question string, k int) []Source {
	qTokens := tokenize(question)
	type scored struct {
		src   Source
		score int
	}
	var ranked []scored

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, c := range s.chunks {
		score := 0
		for _, qt := range qTokens {
			for _, ct := range c.Tokens {
				if qt == ct {
					score++
				}
			}
		}
		if score > 0 {
			ranked = append(ranked, scored{
				src: Source{
					Title:   c.Title,
					Excerpt: c.Content,
					Score:   score,
				},
				score: score,
			})
		}
	}

	for i := 0; i < len(ranked); i++ {
		best := i
		for j := i + 1; j < len(ranked); j++ {
			if ranked[j].score > ranked[best].score {
				best = j
			}
		}
		ranked[i], ranked[best] = ranked[best], ranked[i]
	}

	if k > len(ranked) {
		k = len(ranked)
	}
	out := make([]Source, 0, k)
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

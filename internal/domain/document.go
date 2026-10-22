package domain

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

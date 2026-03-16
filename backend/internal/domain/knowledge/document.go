package knowledge

import "time"

type Document struct {
	ID        string            `json:"id"`
	PageID    string            `json:"page_id"`
	Title     string            `json:"title"`
	Content   string            `json:"content"`
	Chunk     int               `json:"chunk"`
	Metadata  map[string]string `json:"metadata"`
	Vector    []float64         `json:"vector"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type SearchResult struct {
	Document   Document `json:"document"`
	Similarity float64  `json:"similarity"`
}

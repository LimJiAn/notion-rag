package documents

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/jian1990/notion-rag/backend/internal/domain/knowledge"
)

func TestSearch(t *testing.T) {
	s, err := New(filepath.Join(t.TempDir(), "documents.json"))
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	now := time.Now().UTC()
	err = s.Replace(context.Background(), []knowledge.Document{
		{
			ID:        "doc-1",
			Title:     "alpha",
			Content:   "hello world",
			Vector:    []float64{1, 0},
			UpdatedAt: now,
		},
		{
			ID:        "doc-2",
			Title:     "beta",
			Content:   "other",
			Vector:    []float64{0, 1},
			UpdatedAt: now,
		},
	})
	if err != nil {
		t.Fatalf("replace: %v", err)
	}

	results, err := s.Search(context.Background(), []float64{0.9, 0.1}, 1, 0.1)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Document.ID != "doc-1" {
		t.Fatalf("expected doc-1, got %s", results[0].Document.ID)
	}
}

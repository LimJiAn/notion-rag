package store

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/jian1990/notion-rag/backend/internal/models"
)

type Store struct {
	path      string
	mu        sync.RWMutex
	documents []models.Document
}

func New(path string) (*Store, error) {
	s := &Store{path: path}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) Replace(ctx context.Context, documents []models.Document) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.documents = append([]models.Document(nil), documents...)
	return s.persistLocked()
}

func (s *Store) Search(ctx context.Context, query []float64, topK int, minSimilarity float64) ([]models.SearchResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]models.SearchResult, 0, len(s.documents))
	for _, doc := range s.documents {
		if len(doc.Vector) == 0 {
			continue
		}
		score := cosineSimilarity(query, doc.Vector)
		if score < minSimilarity {
			continue
		}
		results = append(results, models.SearchResult{
			Document:   doc,
			Similarity: score,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	if topK > 0 && len(results) > topK {
		results = results[:topK]
	}

	return results, nil
}

func (s *Store) Stats() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lastUpdated := ""
	if n := len(s.documents); n > 0 {
		lastUpdated = s.documents[n-1].UpdatedAt.Format(time.RFC3339)
	}

	return map[string]any{
		"documents":    len(s.documents),
		"last_updated": lastUpdated,
	}
}

func (s *Store) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		s.documents = nil
		return nil
	}
	if err != nil {
		return err
	}
	if len(data) == 0 {
		s.documents = nil
		return nil
	}
	return json.Unmarshal(data, &s.documents)
}

func (s *Store) persistLocked() error {
	data, err := json.MarshalIndent(s.documents, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

func cosineSimilarity(a, b []float64) float64 {
	if len(a) == 0 || len(a) != len(b) {
		return 0
	}

	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

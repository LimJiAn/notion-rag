package weaviate

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"

	"github.com/jian1990/notion-rag/backend/internal/domain/knowledge"
)

const batchSize = 100

type Store struct {
	baseURL   string
	className string
	client    *http.Client
	timeout   time.Duration
}

type statusError struct {
	code int
	body string
}

func (e statusError) Error() string {
	if e.body == "" {
		return fmt.Sprintf("weaviate returned status %d", e.code)
	}
	return fmt.Sprintf("weaviate returned status %d: %s", e.code, e.body)
}

func New(baseURL, className string, timeout time.Duration) (*Store, error) {
	parsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return nil, err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("invalid WEAVIATE_URL %q", baseURL)
	}
	if className == "" {
		className = "NotionChunk"
	}
	if !isGraphQLName(className) {
		return nil, fmt.Errorf("invalid WEAVIATE_CLASS_NAME %q", className)
	}

	return &Store{
		baseURL:   strings.TrimRight(parsed.String(), "/"),
		className: className,
		client:    &http.Client{Timeout: timeout},
		timeout:   timeout,
	}, nil
}

func (s *Store) Replace(ctx context.Context, documents []knowledge.Document) error {
	if err := s.resetClass(ctx); err != nil {
		return err
	}
	if len(documents) == 0 {
		return nil
	}

	for start := 0; start < len(documents); start += batchSize {
		end := start + batchSize
		if end > len(documents) {
			end = len(documents)
		}
		if err := s.createBatch(ctx, documents[start:end]); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) Search(ctx context.Context, query []float64, topK int, minSimilarity float64) ([]knowledge.SearchResult, error) {
	if len(query) == 0 {
		return nil, nil
	}
	if err := s.ensureClass(ctx); err != nil {
		return nil, err
	}
	if topK <= 0 {
		topK = 6
	}

	vectorLiteral, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	maxDistance := math.Max(0, 1-minSimilarity)
	graphQL := fmt.Sprintf(`{
  Get {
    %s(
      nearVector: { vector: %s, distance: %.6f },
      limit: %d
    ) {
      documentId
      pageId
      title
      content
      chunk
      source
      updatedAt
      _additional {
        distance
      }
    }
  }
}`, s.className, string(vectorLiteral), maxDistance, topK)

	var payload graphQLResponse
	if err := s.graphQL(ctx, graphQL, &payload); err != nil {
		return nil, err
	}
	raw, ok := payload.Data.Get[s.className]
	if !ok {
		return nil, nil
	}

	var rows []objectRow
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, err
	}

	results := make([]knowledge.SearchResult, 0, len(rows))
	for _, row := range rows {
		results = append(results, knowledge.SearchResult{
			Document:   row.toDocument(),
			Similarity: distanceToSimilarity(row.Additional.Distance),
		})
	}
	return results, nil
}

func (s *Store) List(ctx context.Context, limit int) ([]knowledge.Document, error) {
	if err := s.ensureClass(ctx); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 25
	}

	graphQL := fmt.Sprintf(`{
  Get {
    %s(limit: %d) {
      documentId
      pageId
      title
      content
      chunk
      source
      updatedAt
    }
  }
}`, s.className, limit)

	var payload graphQLResponse
	if err := s.graphQL(ctx, graphQL, &payload); err != nil {
		return nil, err
	}
	raw, ok := payload.Data.Get[s.className]
	if !ok {
		return nil, nil
	}

	var rows []objectRow
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, err
	}

	documents := make([]knowledge.Document, 0, len(rows))
	for _, row := range rows {
		documents = append(documents, row.toDocument())
	}
	return documents, nil
}

func (s *Store) Stats() map[string]any {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	if err := s.ensureClass(ctx); err != nil {
		return map[string]any{
			"documents":    0,
			"last_updated": "",
			"vector_store": "weaviate",
			"collection":   s.className,
			"weaviate_url": s.baseURL,
			"status":       "unavailable",
			"error":        err.Error(),
		}
	}

	graphQL := fmt.Sprintf(`{
  Aggregate {
    %s {
      meta {
        count
      }
    }
  }
}`, s.className)

	var payload graphQLResponse
	if err := s.graphQL(ctx, graphQL, &payload); err != nil {
		return map[string]any{
			"documents":    0,
			"last_updated": "",
			"vector_store": "weaviate",
			"collection":   s.className,
			"weaviate_url": s.baseURL,
			"status":       "unavailable",
			"error":        err.Error(),
		}
	}

	count := 0
	if raw, ok := payload.Data.Aggregate[s.className]; ok {
		var rows []aggregateRow
		if err := json.Unmarshal(raw, &rows); err == nil && len(rows) > 0 {
			count = rows[0].Meta.Count
		}
	}

	return map[string]any{
		"documents":    count,
		"last_updated": "",
		"vector_store": "weaviate",
		"collection":   s.className,
		"weaviate_url": s.baseURL,
		"status":       "ready",
	}
}

func (s *Store) ensureClass(ctx context.Context) error {
	err := s.request(ctx, http.MethodGet, "/v1/schema/"+s.className, nil, nil)
	var status statusError
	if err != nil && errors.As(err, &status) && status.code == http.StatusNotFound {
		return s.createClass(ctx)
	}
	return err
}

func (s *Store) resetClass(ctx context.Context) error {
	err := s.request(ctx, http.MethodDelete, "/v1/schema/"+s.className, nil, nil)
	var status statusError
	if err != nil && !(errors.As(err, &status) && status.code == http.StatusNotFound) {
		return err
	}

	return s.createClass(ctx)
}

func (s *Store) createClass(ctx context.Context) error {
	body := map[string]any{
		"class":       s.className,
		"description": "Notion chunks indexed by notion-rag",
		"vectorizer":  "none",
		"vectorIndexConfig": map[string]any{
			"distance": "cosine",
		},
		"properties": []map[string]any{
			{"name": "documentId", "dataType": []string{"text"}},
			{"name": "pageId", "dataType": []string{"text"}},
			{"name": "title", "dataType": []string{"text"}},
			{"name": "content", "dataType": []string{"text"}},
			{"name": "chunk", "dataType": []string{"int"}},
			{"name": "source", "dataType": []string{"text"}},
			{"name": "updatedAt", "dataType": []string{"date"}},
		},
	}

	return s.request(ctx, http.MethodPost, "/v1/schema", body, nil)
}

func (s *Store) createBatch(ctx context.Context, documents []knowledge.Document) error {
	objects := make([]map[string]any, 0, len(documents))
	for _, doc := range documents {
		properties := map[string]any{
			"documentId": doc.ID,
			"pageId":     doc.PageID,
			"title":      doc.Title,
			"content":    doc.Content,
			"chunk":      doc.Chunk,
			"source":     doc.Metadata["source"],
		}
		if !doc.UpdatedAt.IsZero() {
			properties["updatedAt"] = doc.UpdatedAt.UTC().Format(time.RFC3339)
		}
		objects = append(objects, map[string]any{
			"class":      s.className,
			"id":         deterministicUUID(doc.ID),
			"properties": properties,
			"vector":     doc.Vector,
		})
	}

	var response []batchObjectResponse
	if err := s.request(ctx, http.MethodPost, "/v1/batch/objects", map[string]any{"objects": objects}, &response); err != nil {
		return err
	}
	for _, item := range response {
		if len(item.Result.Errors.Error) == 0 {
			continue
		}
		return fmt.Errorf("weaviate batch object %s failed: %s", item.ID, item.Result.Errors.Error[0].Message)
	}
	return nil
}

func (s *Store) graphQL(ctx context.Context, query string, out *graphQLResponse) error {
	if err := s.request(ctx, http.MethodPost, "/v1/graphql", map[string]string{"query": query}, out); err != nil {
		return err
	}
	if len(out.Errors) > 0 {
		return fmt.Errorf("weaviate graphql error: %s", out.Errors[0].Message)
	}
	return nil
}

func (s *Store) request(ctx context.Context, method, path string, body any, out any) error {
	var payload io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payload = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, s.baseURL+path, payload)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		data, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
		return statusError{code: res.StatusCode, body: string(data)}
	}
	if out == nil {
		io.Copy(io.Discard, res.Body)
		return nil
	}
	return json.NewDecoder(res.Body).Decode(out)
}

type graphQLResponse struct {
	Data struct {
		Get       map[string]json.RawMessage `json:"Get"`
		Aggregate map[string]json.RawMessage `json:"Aggregate"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type objectRow struct {
	DocumentID string `json:"documentId"`
	PageID     string `json:"pageId"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Chunk      int    `json:"chunk"`
	Source     string `json:"source"`
	UpdatedAt  string `json:"updatedAt"`
	Additional struct {
		Distance float64 `json:"distance"`
	} `json:"_additional"`
}

func (r objectRow) toDocument() knowledge.Document {
	updatedAt, _ := time.Parse(time.RFC3339, r.UpdatedAt)
	metadata := map[string]string{}
	if r.Source != "" {
		metadata["source"] = r.Source
	}
	return knowledge.Document{
		ID:        r.DocumentID,
		PageID:    r.PageID,
		Title:     r.Title,
		Content:   r.Content,
		Chunk:     r.Chunk,
		Metadata:  metadata,
		UpdatedAt: updatedAt,
	}
}

type aggregateRow struct {
	Meta struct {
		Count int `json:"count"`
	} `json:"meta"`
}

type batchObjectResponse struct {
	ID     string `json:"id"`
	Result struct {
		Errors struct {
			Error []struct {
				Message string `json:"message"`
			} `json:"error"`
		} `json:"errors"`
	} `json:"result"`
}

func deterministicUUID(value string) string {
	sum := md5.Sum([]byte(value))
	sum[6] = (sum[6] & 0x0f) | 0x30
	sum[8] = (sum[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", sum[0:4], sum[4:6], sum[6:8], sum[8:10], sum[10:16])
}

func distanceToSimilarity(distance float64) float64 {
	score := 1 - distance
	if score < 0 {
		return 0
	}
	if score > 1 {
		return 1
	}
	return score
}

func isGraphQLName(value string) bool {
	for i, r := range value {
		if i == 0 && !(unicode.IsLetter(r) || r == '_') {
			return false
		}
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_') {
			return false
		}
	}
	return value != ""
}

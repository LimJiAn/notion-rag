package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jian1990/notion-rag/backend/internal/settings"
)

type Client struct {
	settings   *settings.Store
	baseURL    string
	httpClient *http.Client
}

func NewClient(settingsStore *settings.Store, timeout time.Duration) *Client {
	return &Client{
		settings:   settingsStore,
		baseURL:    "https://generativelanguage.googleapis.com/v1beta",
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *Client) Embed(ctx context.Context, text string) ([]float64, error) {
	current := c.settings.Snapshot()

	payload := map[string]any{
		"model": fmt.Sprintf("models/%s", current.EmbeddingModel),
		"content": map[string]any{
			"parts": []map[string]string{
				{"text": text},
			},
		},
		"taskType": "RETRIEVAL_DOCUMENT",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/models/%s:embedContent?key=%s", c.baseURL, current.EmbeddingModel, current.GeminiAPIKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("embed request failed: status=%d body=%s", resp.StatusCode, string(data))
	}

	var parsed struct {
		Embedding struct {
			Values []float64 `json:"values"`
		} `json:"embedding"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, err
	}
	return parsed.Embedding.Values, nil
}

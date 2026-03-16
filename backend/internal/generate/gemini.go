package generate

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

func (c *Client) Answer(ctx context.Context, prompt string) (string, error) {
	current := c.settings.Snapshot()
	payload := map[string]any{
		"contents": []map[string]any{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.baseURL, current.GenerationModel, current.GeminiAPIKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("generate request failed: status=%d body=%s", resp.StatusCode, string(data))
	}

	var parsed struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty generation response")
	}
	return parsed.Candidates[0].Content.Parts[0].Text, nil
}

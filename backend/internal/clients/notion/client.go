package notion

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/jian1990/notion-rag/backend/internal/settings"
)

type Client struct {
	settings   *settings.Store
	httpClient *http.Client
	baseURL    string
}

type Page struct {
	ID      string
	Title   string
	Content string
}

type blocksResponse struct {
	Results    []block `json:"results"`
	HasMore    bool    `json:"has_more"`
	NextCursor string  `json:"next_cursor"`
}

type block struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	HasChild  bool           `json:"has_children"`
	Paragraph *richTextBlock `json:"paragraph,omitempty"`
	Heading1  *richTextBlock `json:"heading_1,omitempty"`
	Heading2  *richTextBlock `json:"heading_2,omitempty"`
	Heading3  *richTextBlock `json:"heading_3,omitempty"`
	Bulleted  *richTextBlock `json:"bulleted_list_item,omitempty"`
	Numbered  *richTextBlock `json:"numbered_list_item,omitempty"`
	ToDo      *richTextBlock `json:"to_do,omitempty"`
	Toggle    *richTextBlock `json:"toggle,omitempty"`
	Quote     *richTextBlock `json:"quote,omitempty"`
	Callout   *richTextBlock `json:"callout,omitempty"`
	Code      *richTextBlock `json:"code,omitempty"`
	ChildPage *struct {
		Title string `json:"title"`
	} `json:"child_page,omitempty"`
}

func NewClient(settingsStore *settings.Store, timeout time.Duration) *Client {
	return &Client{
		settings:   settingsStore,
		httpClient: &http.Client{Timeout: timeout},
		baseURL:    "https://api.notion.com/v1",
	}
}

func (c *Client) Crawl(ctx context.Context, rootPageIDs []string) ([]Page, error) {
	visited := make(map[string]struct{})
	pages := make([]Page, 0, len(rootPageIDs))

	for _, pageID := range rootPageIDs {
		if err := c.walkPage(ctx, pageID, visited, &pages); err != nil {
			return nil, err
		}
	}

	return pages, nil
}

func (c *Client) walkPage(ctx context.Context, pageID string, visited map[string]struct{}, out *[]Page) error {
	if _, ok := visited[pageID]; ok {
		return nil
	}
	visited[pageID] = struct{}{}

	title, err := c.getPageTitle(ctx, pageID)
	if err != nil {
		return err
	}

	content, childPageIDs, err := c.getBlockChildren(ctx, pageID)
	if err != nil {
		return err
	}

	*out = append(*out, Page{
		ID:      pageID,
		Title:   title,
		Content: strings.TrimSpace(content),
	})

	for _, childPageID := range childPageIDs {
		if err := c.walkPage(ctx, childPageID, visited, out); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) getPageTitle(ctx context.Context, pageID string) (string, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("%s/pages/%s", c.baseURL, pageID), nil)
	if err != nil {
		return "", err
	}

	var parsed struct {
		Properties map[string]struct {
			Type  string `json:"type"`
			Title []struct {
				PlainText string `json:"plain_text"`
			} `json:"title"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(resp, &parsed); err != nil {
		return "", err
	}

	for _, property := range parsed.Properties {
		if property.Type != "title" {
			continue
		}
		var parts []string
		for _, title := range property.Title {
			parts = append(parts, title.PlainText)
		}
		text := strings.TrimSpace(strings.Join(parts, " "))
		if text != "" {
			return text, nil
		}
	}

	return pageID, nil
}

func (c *Client) getBlockChildren(ctx context.Context, blockID string) (string, []string, error) {
	return c.getBlockChildrenPage(ctx, blockID, "")
}

func (c *Client) getBlockChildrenPage(ctx context.Context, blockID, cursor string) (string, []string, error) {
	url := fmt.Sprintf("%s/blocks/%s/children?page_size=100", c.baseURL, blockID)
	if cursor != "" {
		url += "&start_cursor=" + cursor
	}

	resp, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", nil, err
	}

	var parsed blocksResponse
	if err := json.Unmarshal(resp, &parsed); err != nil {
		return "", nil, err
	}

	var parts []string
	var childPages []string
	for _, block := range parsed.Results {
		if block.Type == "child_page" {
			childPages = append(childPages, block.ID)
			if block.ChildPage != nil && strings.TrimSpace(block.ChildPage.Title) != "" {
				parts = append(parts, block.ChildPage.Title)
			}
			continue
		}
		if text := extractBlockText(block); text != "" {
			parts = append(parts, text)
		}
	}
	if parsed.HasMore && parsed.NextCursor != "" {
		nextContent, nextPages, err := c.getBlockChildrenPage(ctx, blockID, parsed.NextCursor)
		if err != nil {
			return "", nil, err
		}
		if nextContent != "" {
			parts = append(parts, nextContent)
		}
		childPages = append(childPages, nextPages...)
	}

	return strings.Join(parts, "\n"), childPages, nil
}

type richTextBlock struct {
	RichText []struct {
		PlainText string `json:"plain_text"`
	} `json:"rich_text"`
}

func extractBlockText(value any) string {
	switch v := value.(type) {
	case block:
		return joinRichText(v.Paragraph, v.Heading1, v.Heading2, v.Heading3, v.Bulleted, v.Numbered, v.ToDo, v.Toggle, v.Quote, v.Callout, v.Code)
	default:
		return ""
	}
}

func joinRichText(blocks ...*richTextBlock) string {
	var parts []string
	for _, block := range blocks {
		if block == nil {
			continue
		}
		for _, item := range block.RichText {
			if text := strings.TrimSpace(item.PlainText); text != "" {
				parts = append(parts, text)
			}
		}
	}
	return strings.Join(parts, " ")
}

func (c *Client) doRequest(ctx context.Context, method, url string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	current := c.settings.Snapshot()
	req.Header.Set("Authorization", "Bearer "+current.NotionToken)
	req.Header.Set("Notion-Version", current.NotionVersion)
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
		return nil, fmt.Errorf("notion request failed: status=%d body=%s", resp.StatusCode, string(data))
	}
	return data, nil
}

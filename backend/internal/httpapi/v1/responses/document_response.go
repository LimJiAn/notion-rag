package responses

import (
	"fmt"
	"strings"
	"time"

	"github.com/jian1990/notion-rag/backend/internal/domain/knowledge"
)

type DocumentResponse struct {
	Title     string `json:"title"`
	PageID    string `json:"page_id"`
	Content   string `json:"content"`
	Chunk     int    `json:"chunk"`
	NotionURL string `json:"notion_url,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

func NewDocumentResponse(doc knowledge.Document) DocumentResponse {
	return DocumentResponse{
		Title:     doc.Title,
		PageID:    doc.PageID,
		Content:   doc.Content,
		Chunk:     doc.Chunk,
		NotionURL: buildNotionURL(doc.PageID),
		UpdatedAt: formatUpdatedAt(doc.UpdatedAt),
	}
}

func buildNotionURL(pageID string) string {
	normalized := strings.ReplaceAll(strings.TrimSpace(pageID), "-", "")
	if normalized == "" {
		return ""
	}
	return fmt.Sprintf("https://www.notion.so/%s", normalized)
}

func formatUpdatedAt(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

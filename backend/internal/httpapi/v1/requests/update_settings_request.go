package requests

import "strings"

type UpdateSettingsRequest struct {
	NotionToken       string `json:"notion_token"`
	NotionVersion     string `json:"notion_version"`
	NotionRootPageIDs string `json:"notion_root_page_ids"`
	GeminiAPIKey      string `json:"gemini_api_key"`
	EmbeddingModel    string `json:"embedding_model"`
	GenerationModel   string `json:"generation_model"`
}

func (r UpdateSettingsRequest) RootPageIDs() []string {
	if strings.TrimSpace(r.NotionRootPageIDs) == "" {
		return nil
	}

	parts := strings.Split(r.NotionRootPageIDs, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}

	return out
}

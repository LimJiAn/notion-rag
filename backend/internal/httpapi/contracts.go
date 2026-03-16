package httpapi

type queryRequest struct {
	Question string `json:"question"`
}

type updateSettingsRequest struct {
	NotionToken       string `json:"notion_token"`
	NotionVersion     string `json:"notion_version"`
	NotionRootPageIDs string `json:"notion_root_page_ids"`
	GeminiAPIKey      string `json:"gemini_api_key"`
	EmbeddingModel    string `json:"embedding_model"`
	GenerationModel   string `json:"generation_model"`
}

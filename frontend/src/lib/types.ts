export type Stats = {
  documents?: number;
  last_updated?: string;
};

export type SearchResult = {
  similarity: number;
  document: {
    title: string;
    page_id: string;
    content: string;
  };
};

export type QueryResponse = {
  question: string;
  answer: string;
  results: SearchResult[];
};

export type ChatMessage = {
  id: string;
  role: "user" | "assistant";
  text: string;
  createdAt: string;
  results?: SearchResult[];
};

export type HealthStatus = "loading" | "online" | "offline";

export type SettingsResponse = {
  notion_token_set: boolean;
  notion_version: string;
  notion_root_page_ids: string[];
  gemini_api_key_set: boolean;
  embedding_model: string;
  generation_model: string;
};

export type SettingsForm = {
  notionToken: string;
  notionVersion: string;
  notionRootPageIDs: string;
  geminiAPIKey: string;
  embeddingModel: string;
  generationModel: string;
};

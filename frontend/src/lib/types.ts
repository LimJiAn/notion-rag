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
    chunk: number;
    notion_url?: string;
    updated_at?: string;
  };
};

export type Citation = {
  title: string;
  page_id: string;
  snippet: string;
  notion_url?: string;
  similarity: number;
};

export type QueryResponse = {
  question: string;
  answer: string;
  results: SearchResult[];
  citations: Citation[];
  confidence_score: number;
  confidence_label: "high" | "medium" | "low";
  used_context: boolean;
  follow_up_questions: string[];
};

export type ChatMessage = {
  id: string;
  role: "user" | "assistant";
  text: string;
  createdAt: string;
  results?: SearchResult[];
  citations?: Citation[];
  confidenceScore?: number;
  confidenceLabel?: "high" | "medium" | "low";
  usedContext?: boolean;
  followUpQuestions?: string[];
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

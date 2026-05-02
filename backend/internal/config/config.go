package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr          string
	DataDir           string
	StorePath         string
	VectorStore       string
	WeaviateURL       string
	WeaviateClassName string
	NotionToken       string
	NotionVersion     string
	NotionRootPageIDs []string
	GeminiAPIKey      string
	EmbeddingModel    string
	GenerationModel   string
	ChunkSize         int
	ChunkOverlap      int
	TopK              int
	SimilarityCutoff  float64
	WorkerCount       int
	RequestTimeout    time.Duration
	ShutdownTimeout   time.Duration
}

func Load() (Config, error) {
	dataDir := envOrDefault("DATA_DIR", "./data")

	cfg := Config{
		HTTPAddr:          envOrDefault("HTTP_ADDR", ":8080"),
		DataDir:           dataDir,
		StorePath:         filepath.Join(dataDir, "documents.json"),
		VectorStore:       strings.ToLower(envOrDefault("VECTOR_STORE", "file")),
		WeaviateURL:       envOrDefault("WEAVIATE_URL", "http://localhost:8081"),
		WeaviateClassName: envOrDefault("WEAVIATE_CLASS_NAME", "NotionChunk"),
		NotionToken:       strings.TrimSpace(os.Getenv("NOTION_TOKEN")),
		NotionVersion:     envOrDefault("NOTION_VERSION", "2026-03-11"),
		NotionRootPageIDs: splitCSV(os.Getenv("NOTION_ROOT_PAGE_IDS")),
		GeminiAPIKey:      strings.TrimSpace(os.Getenv("GEMINI_API_KEY")),
		EmbeddingModel:    envOrDefault("GEMINI_EMBEDDING_MODEL", "gemini-embedding-001"),
		GenerationModel:   envOrDefault("GEMINI_GENERATION_MODEL", "gemini-2.5-flash"),
		ChunkSize:         envInt("CHUNK_SIZE", 1000),
		ChunkOverlap:      envInt("CHUNK_OVERLAP", 150),
		TopK:              envInt("TOP_K", 6),
		SimilarityCutoff:  envFloat("SIMILARITY_CUTOFF", 0.65),
		WorkerCount:       envInt("WORKER_COUNT", 4),
		RequestTimeout:    envDuration("REQUEST_TIMEOUT", 60*time.Second),
		ShutdownTimeout:   envDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
	}

	if cfg.NotionToken == "" {
		return Config{}, errors.New("NOTION_TOKEN is required")
	}
	if len(cfg.NotionRootPageIDs) == 0 {
		return Config{}, errors.New("NOTION_ROOT_PAGE_IDS is required")
	}
	if cfg.GeminiAPIKey == "" {
		return Config{}, errors.New("GEMINI_API_KEY is required")
	}
	if cfg.ChunkOverlap >= cfg.ChunkSize {
		return Config{}, fmt.Errorf("CHUNK_OVERLAP must be smaller than CHUNK_SIZE")
	}
	if cfg.VectorStore != "file" && cfg.VectorStore != "weaviate" {
		return Config{}, fmt.Errorf("VECTOR_STORE must be either file or weaviate")
	}
	if cfg.VectorStore == "weaviate" && strings.TrimSpace(cfg.WeaviateURL) == "" {
		return Config{}, errors.New("WEAVIATE_URL is required when VECTOR_STORE=weaviate")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envFloat(key string, fallback float64) float64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

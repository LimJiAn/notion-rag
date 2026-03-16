package settings

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jian1990/notion-rag/backend/internal/config"
)

type Values struct {
	NotionToken       string
	NotionVersion     string
	NotionRootPageIDs []string
	GeminiAPIKey      string
	EmbeddingModel    string
	GenerationModel   string
}

type PublicValues struct {
	NotionTokenSet    bool     `json:"notion_token_set"`
	NotionVersion     string   `json:"notion_version"`
	NotionRootPageIDs []string `json:"notion_root_page_ids"`
	GeminiAPIKeySet   bool     `json:"gemini_api_key_set"`
	EmbeddingModel    string   `json:"embedding_model"`
	GenerationModel   string   `json:"generation_model"`
}

type UpdateInput struct {
	NotionToken       string
	NotionVersion     string
	NotionRootPageIDs []string
	GeminiAPIKey      string
	EmbeddingModel    string
	GenerationModel   string
}

type Store struct {
	path   string
	mu     sync.RWMutex
	values Values
}

func New(cfg config.Config) (*Store, error) {
	store := &Store{
		path: filepath.Join(cfg.DataDir, "runtime.env"),
		values: Values{
			NotionToken:       cfg.NotionToken,
			NotionVersion:     cfg.NotionVersion,
			NotionRootPageIDs: append([]string(nil), cfg.NotionRootPageIDs...),
			GeminiAPIKey:      cfg.GeminiAPIKey,
			EmbeddingModel:    cfg.EmbeddingModel,
			GenerationModel:   cfg.GenerationModel,
		},
	}

	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		return nil, err
	}
	if err := store.loadFromFile(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *Store) Snapshot() Values {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return Values{
		NotionToken:       s.values.NotionToken,
		NotionVersion:     s.values.NotionVersion,
		NotionRootPageIDs: append([]string(nil), s.values.NotionRootPageIDs...),
		GeminiAPIKey:      s.values.GeminiAPIKey,
		EmbeddingModel:    s.values.EmbeddingModel,
		GenerationModel:   s.values.GenerationModel,
	}
}

func (s *Store) Public() PublicValues {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return PublicValues{
		NotionTokenSet:    s.values.NotionToken != "",
		NotionVersion:     s.values.NotionVersion,
		NotionRootPageIDs: append([]string(nil), s.values.NotionRootPageIDs...),
		GeminiAPIKeySet:   s.values.GeminiAPIKey != "",
		EmbeddingModel:    s.values.EmbeddingModel,
		GenerationModel:   s.values.GenerationModel,
	}
}

func (s *Store) Update(input UpdateInput) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if trimmed := strings.TrimSpace(input.NotionToken); trimmed != "" {
		s.values.NotionToken = trimmed
	}
	if trimmed := strings.TrimSpace(input.NotionVersion); trimmed != "" {
		s.values.NotionVersion = trimmed
	}
	if len(input.NotionRootPageIDs) > 0 {
		s.values.NotionRootPageIDs = cleanIDs(input.NotionRootPageIDs)
	}
	if trimmed := strings.TrimSpace(input.GeminiAPIKey); trimmed != "" {
		s.values.GeminiAPIKey = trimmed
	}
	if trimmed := strings.TrimSpace(input.EmbeddingModel); trimmed != "" {
		s.values.EmbeddingModel = trimmed
	}
	if trimmed := strings.TrimSpace(input.GenerationModel); trimmed != "" {
		s.values.GenerationModel = trimmed
	}

	if err := validate(s.values); err != nil {
		return err
	}
	return s.persistLocked()
}

func validate(values Values) error {
	if strings.TrimSpace(values.NotionToken) == "" {
		return errors.New("NOTION_TOKEN is required")
	}
	if len(cleanIDs(values.NotionRootPageIDs)) == 0 {
		return errors.New("NOTION_ROOT_PAGE_IDS is required")
	}
	if strings.TrimSpace(values.GeminiAPIKey) == "" {
		return errors.New("GEMINI_API_KEY is required")
	}
	return nil
}

func (s *Store) loadFromFile() error {
	file, err := os.Open(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	overrides := map[string]string{}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		overrides[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	if value := overrides["NOTION_TOKEN"]; value != "" {
		s.values.NotionToken = value
	}
	if value := overrides["NOTION_VERSION"]; value != "" {
		s.values.NotionVersion = value
	}
	if value := overrides["NOTION_ROOT_PAGE_IDS"]; value != "" {
		s.values.NotionRootPageIDs = cleanIDs(strings.Split(value, ","))
	}
	if value := overrides["GEMINI_API_KEY"]; value != "" {
		s.values.GeminiAPIKey = value
	}
	if value := overrides["GEMINI_EMBEDDING_MODEL"]; value != "" {
		s.values.EmbeddingModel = value
	}
	if value := overrides["GEMINI_GENERATION_MODEL"]; value != "" {
		s.values.GenerationModel = value
	}

	return validate(s.values)
}

func (s *Store) persistLocked() error {
	lines := []string{
		fmt.Sprintf("NOTION_TOKEN=%s", s.values.NotionToken),
		fmt.Sprintf("NOTION_VERSION=%s", s.values.NotionVersion),
		fmt.Sprintf("NOTION_ROOT_PAGE_IDS=%s", strings.Join(cleanIDs(s.values.NotionRootPageIDs), ",")),
		fmt.Sprintf("GEMINI_API_KEY=%s", s.values.GeminiAPIKey),
		fmt.Sprintf("GEMINI_EMBEDDING_MODEL=%s", s.values.EmbeddingModel),
		fmt.Sprintf("GEMINI_GENERATION_MODEL=%s", s.values.GenerationModel),
	}
	return os.WriteFile(s.path, []byte(strings.Join(lines, "\n")+"\n"), 0o600)
}

func cleanIDs(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

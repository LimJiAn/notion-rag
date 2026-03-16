package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/jian1990/notion-rag/backend/internal/ingest"
	"github.com/jian1990/notion-rag/backend/internal/rag"
	"github.com/jian1990/notion-rag/backend/internal/settings"
	"github.com/jian1990/notion-rag/backend/internal/store"
)

type Server struct {
	store    *store.Store
	ingest   *ingest.Service
	rag      *rag.Service
	settings *settings.Store
}

func NewServer(store *store.Store, ingest *ingest.Service, rag *rag.Service, settingsStore *settings.Store) *Server {
	return &Server{
		store:    store,
		ingest:   ingest,
		rag:      rag,
		settings: settingsStore,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/api/v1/sync", s.handleSync)
	mux.HandleFunc("/api/v1/query", s.handleQuery)
	mux.HandleFunc("/api/v1/stats", s.handleStats)
	mux.HandleFunc("/api/v1/settings", s.handleSettings)
	return withCORS(mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleStats(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.store.Stats())
}

func (s *Server) handleSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()

	stats, err := s.ingest.Sync(ctx)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status": "synced",
		"stats":  stats,
	})
}

func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Question string `json:"question"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.Question == "" {
		writeError(w, http.StatusBadRequest, "question is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()

	answer, err := s.rag.Ask(ctx, req.Question)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, answer)
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.settings.Public())
	case http.MethodPut:
		var req struct {
			NotionToken       string `json:"notion_token"`
			NotionVersion     string `json:"notion_version"`
			NotionRootPageIDs string `json:"notion_root_page_ids"`
			GeminiAPIKey      string `json:"gemini_api_key"`
			EmbeddingModel    string `json:"embedding_model"`
			GenerationModel   string `json:"generation_model"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid json body")
			return
		}

		err := s.settings.Update(settings.UpdateInput{
			NotionToken:       req.NotionToken,
			NotionVersion:     req.NotionVersion,
			NotionRootPageIDs: splitCSV(req.NotionRootPageIDs),
			GeminiAPIKey:      req.GeminiAPIKey,
			EmbeddingModel:    req.EmbeddingModel,
			GenerationModel:   req.GenerationModel,
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"status":   "saved",
			"settings": s.settings.Public(),
		})
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
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

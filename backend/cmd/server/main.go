package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	_ "github.com/jian1990/notion-rag/backend/docs"
	"github.com/jian1990/notion-rag/backend/internal/app"
	"github.com/jian1990/notion-rag/backend/internal/config"
)

// @title Notion RAG API
// @version 1.0.0
// @description REST API for syncing Notion content, querying indexed knowledge, and managing runtime settings.
// @BasePath /
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	server, err := app.NewServer(cfg)
	if err != nil {
		log.Fatalf("init server: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("backend listening on %s", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
}

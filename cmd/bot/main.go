package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/maks2134/telegram-rag-bot/internal/config"
	"github.com/maks2134/telegram-rag-bot/internal/delivery/telegram"
	"github.com/maks2134/telegram-rag-bot/internal/delivery/web"
	"github.com/maks2134/telegram-rag-bot/internal/infra/keyword"
	"github.com/maks2134/telegram-rag-bot/internal/infra/llm"
	"github.com/maks2134/telegram-rag-bot/internal/usecase"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	kb := keyword.New()
	if err := kb.LoadDir(cfg.DocsDir); err != nil {
		log.Fatalf("docs: %v", err)
	}
	log.Printf("loaded %d docs from %s", len(kb.Docs()), cfg.DocsDir)

	client := llm.New(cfg.LLMBaseURL, cfg.LLMAPIKey, cfg.LLMModel)
	asker := &usecase.Asker{Retriever: kb, Answerer: client}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           (&web.Server{Service: asker, Catalog: kb, Model: cfg.LLMModel}).Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("web UI listening on http://localhost%s", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("http: %v", err)
			stop()
		}
	}()

	if cfg.TelegramToken != "" {
		tg, err := telegram.NewTelegram(cfg.TelegramToken, asker, cfg.AdminChatID)
		if err != nil {
			log.Fatalf("telegram: %v", err)
		}
		go func() {
			if err := tg.Run(ctx); err != nil && ctx.Err() == nil {
				log.Printf("telegram stopped: %v", err)
			}
		}()
	} else {
		log.Printf("TELEGRAM_TOKEN empty — web UI only (set token for polling)")
	}

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("http shutdown: %v", err)
	}
	log.Printf("shutdown complete")
}

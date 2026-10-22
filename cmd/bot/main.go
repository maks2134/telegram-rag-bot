package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/maks2134/telegram-rag-bot/internal/bot"
	"github.com/maks2134/telegram-rag-bot/internal/config"
	"github.com/maks2134/telegram-rag-bot/internal/llm"
	"github.com/maks2134/telegram-rag-bot/internal/pipeline"
	"github.com/maks2134/telegram-rag-bot/internal/store"
	"github.com/maks2134/telegram-rag-bot/internal/web"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	kb := store.New()
	if err := kb.LoadDir(cfg.DocsDir); err != nil {
		log.Fatalf("docs: %v", err)
	}
	log.Printf("loaded %d docs from %s", len(kb.Docs()), cfg.DocsDir)

	client := llm.New(cfg.LLMBaseURL, cfg.LLMAPIKey, cfg.LLMModel)
	asker := &pipeline.Asker{Store: kb, LLM: client}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           (&web.Server{Asker: asker, Store: kb, Model: cfg.LLMModel}).Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("web UI listening on http://localhost%s", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http: %v", err)
		}
	}()

	if cfg.TelegramToken != "" {
		tg, err := bot.NewTelegram(cfg.TelegramToken, asker, cfg.AdminChatID)
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
	_ = srv.Shutdown(shutdownCtx)
	log.Printf("shutdown complete")
}

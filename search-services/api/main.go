package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"yadro.com/course/api/adapters/aaa"
	"yadro.com/course/api/adapters/search"
	"yadro.com/course/api/adapters/update"
	"yadro.com/course/api/adapters/words"
	"yadro.com/course/api/config"
	"yadro.com/course/api/core"

	_ "yadro.com/course/api/docs"
)

type traceHandler struct {
	slog.Handler
}

// @title Search comics API
// @version 1.0
// @description Микросервис поиска комиксов XKCD
// @termsOfService http://swagger.io/terms/

// @contact.name Roman
// @contact.email r.garanin2014@yandex.ru

// @license.name MIT

// @host localhost:28080
// @BasePath /
func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()

	cfg := config.MustLoad(configPath)

	log := mustMakeLogger(cfg.LogLevel)

	log.Info("starting server")
	log.Debug("debug messages are enabled")

	// Клиент для сервиса update(Обновление БД)
	updateClient, err := update.NewUpdateClient(cfg.UpdateAddress, log)
	if err != nil {
		log.Error("cannot init update adapter", "error", err)
		os.Exit(1)
	}

	// Клиент для сервиса words(Номрализация слов)
	wordsClient, err := words.NewWordClient(cfg.WordsAddress, log)
	if err != nil {
		log.Error("cannot init words adapter", "error", err)
		os.Exit(1)
	}

	// Клиент для сервиса search(Поиск комиксов)
	searchClient, err := search.NewSearchClient(cfg.SearchAddress, log)
	if err != nil {
		log.Error("cannot init search adapter", "error", err)
		os.Exit(1)
	}

	// Клиент для сервиса aaa(Авторизация)
	aaaClient, err := aaa.NewAAAClient(cfg.DBAddress, cfg.AdminName, cfg.AdminPassword, cfg.TokenTTL, cfg.JWTSecret , log)
	if err != nil {
		log.Error("cannot init aaa adapter", "error", err)
		os.Exit(1)
	}

	mux := NewRouter(log, cfg, updateClient, wordsClient, searchClient, aaaClient, aaaClient)
	server := http.Server{
		Addr:        cfg.HTTPConfig.Address,
		ReadTimeout: cfg.HTTPConfig.Timeout,
		Handler:     mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		log.Debug("shutting down server")
		if err := server.Shutdown(context.Background()); err != nil {
			log.Error("erroneous shutdown", "error", err)
		}
	}()

	log.Info("Running HTTP server", "address", cfg.HTTPConfig.Address)
	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("server closed unexpectedly", "error", err)
			return
		}
	}
}

// Установка уровня логирования для slog
func mustMakeLogger(logLevel string) *slog.Logger {
	var level slog.Level
	switch logLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "ERROR":
		level = slog.LevelError
	default:
		panic("unknown log level: " + logLevel)
	}
	base := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	})
	return slog.New(&traceHandler{base})
}

// Handler для логирования, добавляет в лог trace_id для отслеживания
func (h *traceHandler) Handle(ctx context.Context, r slog.Record) error {
	if id := core.TraceIDFromContext(ctx); id != "" {
		r.AddAttrs(slog.String("trace_id", id))
	}
	return h.Handler.Handle(ctx, r)
}

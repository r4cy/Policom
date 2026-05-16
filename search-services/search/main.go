package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	searchpd "yadro.com/course/proto/search"
	"yadro.com/course/search/adapters/broker"
	db "yadro.com/course/search/adapters/db/postgres"
	searchgrpc "yadro.com/course/search/adapters/grpc"
	"yadro.com/course/search/adapters/initiator"
	"yadro.com/course/search/adapters/words"
	"yadro.com/course/search/config"
	"yadro.com/course/search/core"
)

func main() {

	// config
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()
	cfg := config.MustLoad(configPath)

	// logger
	log := mustMakeLogger(cfg.LogLevel)

	if err := run(cfg, log); err != nil {
		log.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func run(cfg config.Config, log *slog.Logger) error {
	log.Info("starting server")
	log.Debug("debug messages are enabled")

	// context for Ctrl-C
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Клиент для БД
	storage, err := db.NewDBClient(log, cfg.DBAddress)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %v", err)
	}

	index := core.NewMemoryIndex()

	// Клиент для сервиса words(Номрализация слов)
	words, err := words.NewWordClient(cfg.WordsAddress, log)
	if err != nil {
		return fmt.Errorf("failed create Words client: %v", err)
	}

	// Клиент для брокера сообщений(Nats)
	broker, err := broker.NewClientNats(cfg.BrokerAddress, log)
	if err != nil {
		return fmt.Errorf("failed create Broker client: %v", err)
	}

	// Основной сервис
	searcher, err := core.NewService(log, storage, words, index, broker)
	if err != nil {
		return fmt.Errorf("failed create Update service: %v", err)
	}

	// Запуск горутины с подпиской на update
	go func() {
		if err := broker.Subscribe(ctx, searcher.HandlerOptions); err != nil {
			log.Error("broker subscribe failed", "error", err)
		}
	}()

	initiat := initiator.NewInitiatorIndex(log, cfg.SearchServer.IndexTTL, searcher)
	initiat.Start(ctx)

	// grpc server
	listener, err := net.Listen("tcp", cfg.SearchServer.Address)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(searchgrpc.LoggingInterceptor(log)))
	searchpd.RegisterSearchServer(s, searchgrpc.NewServer(searcher))
	reflection.Register(s)

	go func() {
		<-ctx.Done()
		log.Debug("shutting down server")
		s.GracefulStop()
	}()

	if err := s.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}
	return nil
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
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	})
	return slog.New(handler)
}

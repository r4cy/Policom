package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	wordspb "yadro.com/course/proto/words"
)

type Config struct {
	LogLevel   string `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	Address    string `yaml:"words_address" env:"WORDS_ADDRESS" env-required:"true"`
	MaxSizeMsg int    `yaml:"maxSizeMsg" env:"MAX_SIZE_MSG" env-default:"20000"`
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "variant of config")
	flag.Parse()
	var cfg Config
	if _, err := os.Stat(configPath); err == nil {
		if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
			log.Fatalf("Read config from file %q failed with err: %v", configPath, err)
		}
	} else {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			log.Fatalf("Read config from environment failed with err: %v", err)
		}
	}

	log := mustMakeLogger(cfg.LogLevel)

	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		log.Error("failed to listen", "err", err)
		os.Exit(1)
	}

	s := grpc.NewServer(
		grpc.MaxRecvMsgSize(cfg.MaxSizeMsg),
		grpc.UnaryInterceptor(LoggingInterceptor(log)),
	)
	wordspb.RegisterWordsServer(s, NewServer(log))
	reflection.Register(s)

	serverErr := make(chan error, 1)
	go func() {
		err := s.Serve(listener)
		serverErr <- err
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctx.Done():
		log.Debug("shutting down server")
		stop()

		drained := make(chan struct{})
		go func() {
			s.GracefulStop()
			close(drained)
		}()

		select {
		case <-drained:
			log.Debug("Complete stopped the server")
		case <-time.After(10 * time.Second):
			log.Debug("Graceful shutdown timeout")
			s.Stop()
		}

	case err := <-serverErr:
		if err != nil {
			log.Error("serverErr in grpc server with", "err", err)
			log.Error("Server is shotdown")
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
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	})
	return slog.New(handler)
}

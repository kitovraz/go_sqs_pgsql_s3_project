package apiserver

import (
	"context"
	"go_sqs_pqsql_s3_project/config"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"
)

type ApiServer struct {
	config *config.Config
	logger *slog.Logger
}

func New(cfg *config.Config, logger *slog.Logger) *ApiServer {
	return &ApiServer{
		config: cfg,
		logger: logger,
	}
}

func (s *ApiServer) ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

func (s *ApiServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", s.ping)

	middleware := NewLoggerMiddleware(s.logger)

	server := &http.Server{
		Addr:    net.JoinHostPort(s.config.ApiServerHost, s.config.ApiServerPort),
		Handler: middleware(mux),
	}

	go func() {
		s.logger.Info("Apiserver running", "port", s.config.ApiServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Apiserver failed to listen and serve", "error", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()

		shutDownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutDownCtx); err != nil {
			s.logger.Error("Apiserver failed to shutdown", "error", err)
		}
	}()
	wg.Wait()

	return nil
}

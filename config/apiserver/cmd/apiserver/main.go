package main

import (
	"context"
	"go_sqs_pqsql_s3_project/config"
	"go_sqs_pqsql_s3_project/config/apiserver"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	config, err := config.New()
	if err != nil {
		return err
	}

	jsonHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(jsonHandler)
	apiserver := apiserver.New(config, logger)
	if err := apiserver.Start(ctx); err != nil {
		return err
	}
	return nil
}

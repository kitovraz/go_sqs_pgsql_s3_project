package main

import (
	"context"
	"go_sqs_pqsql_s3_project/config"
	"go_sqs_pqsql_s3_project/config/apiserver"
	"go_sqs_pqsql_s3_project/store"
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

	dataStore, err := dataStore(config)
	if err != nil {
		return err
	}

	apiserver := apiserver.New(config, logger(), dataStore)
	if err := apiserver.Start(ctx); err != nil {
		return err
	}
	return nil
}

func logger() *slog.Logger {
	jsonHandler := slog.NewJSONHandler(os.Stdout, nil)
	return slog.New(jsonHandler)
}

func dataStore(config *config.Config) (*store.Store, error) {
	db, err := store.NewPostgresDb(config)
	if err != nil {
		return nil, err
	}
	return store.New(db), nil
}

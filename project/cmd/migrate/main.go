package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"pharma-platform/internal/config"
	"pharma-platform/internal/postgres"
	"pharma-platform/internal/questdb"
	"pharma-platform/internal/store"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg, err := config.Load("config/bootstrap.yaml")
	if err != nil {
		log.Fatal(err)
	}

	if err := config.Validate(cfg); err != nil {
		log.Fatal(err)
	}

	if err := ensureDatabase(ctx, cfg.Postgres); err != nil {
		log.Fatal(err)
	}

	postgresClient := postgres.New(cfg.Postgres)
	if err := postgresClient.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer postgresClient.Close()

	if err := store.MigratePostgres(ctx, postgresClient,
		"deploy/postgres/init",
		"",
		false,
	); err != nil {
		log.Fatal(err)
	}
	log.Println("Postgres schema migrated")

	questClient := questdb.New(cfg.QuestDB)
	if err := questClient.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer questClient.Close()

	if err := store.MigrateQuestDB(ctx, questClient, "deploy/questdb/init"); err != nil {
		log.Fatal(err)
	}
	log.Println("QuestDB schema migrated")
}

func ensureDatabase(ctx context.Context, cfg postgres.Config) error {
	adminCfg := cfg
	adminCfg.Database = "postgres"

	adminClient := postgres.New(adminCfg)
	if err := adminClient.Connect(ctx); err != nil {
		return fmt.Errorf("connect to postgres admin: %w", err)
	}
	defer adminClient.Close()

	var exists bool
	err := adminClient.DB().QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", cfg.Database,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check database exists: %w", err)
	}

	if !exists {
		_, err := adminClient.DB().ExecContext(ctx,
			"CREATE DATABASE "+cfg.Database,
		)
		if err != nil {
			return fmt.Errorf("create database: %w", err)
		}
		log.Printf("Created database %q", cfg.Database)
	}

	return nil
}

package main

import (
	"context"
	"log"

	"pharma-platform/internal/config"
	"pharma-platform/internal/postgres"
	"pharma-platform/internal/store"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load("config/bootstrap.yaml")
	if err != nil {
		log.Fatal(err)
	}

	if err := config.Validate(cfg); err != nil {
		log.Fatal(err)
	}

	postgresClient := postgres.New(cfg.Postgres)
	if err := postgresClient.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer postgresClient.Close()

	if err := store.MigratePostgres(ctx, postgresClient,
		"deploy/postgres/init",
		"deploy/postgres/seed",
		true,
	); err != nil {
		log.Fatal(err)
	}

	log.Println("Seed complete — machines and tags populated from SQL files")
}

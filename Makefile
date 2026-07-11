SHELL := /bin/bash

PORT_CHECK = @if command -v ss >/dev/null 2>&1 && ss -tlnp 2>/dev/null | grep -q ':8081 '; then \
		echo "Error: port 8081 is already in use by:"; \
		ss -tlnp | grep ':8081 '; \
		exit 1; \
	fi

setup:
	mkdir -p persistent/postgres persistent/questdb

up: setup
	docker compose -f project/runtime/docker-compose.yml up -d postgres questdb

up-all: setup
	docker compose -f project/runtime/docker-compose.yml up --build -d

down:
	docker compose -f project/runtime/docker-compose.yml down

logs:
	docker compose -f project/runtime/docker-compose.yml logs -f

dev:
	$(PORT_CHECK)
	cd project && go build -o dev-mode cmd/dev-mode/main.go && exec ./dev-mode

api:
	$(PORT_CHECK)
	cd project && go build -o api cmd/api/main.go && exec ./api

sim:
	cd project && go build -o collector-sim cmd/collector-sim/collector-sim.go && exec ./collector-sim

migrate:
	cd project && go run cmd/migrate/main.go

seed:
	cd project && go run cmd/seed/main.go

build:
	cd project && go build ./...

prod:
	$(PORT_CHECK)
	cd project && go build -o pharma-platform cmd/pharma-platform/main.go && exec ./pharma-platform

.PHONY: setup up up-all down logs dev api sim migrate seed build prod

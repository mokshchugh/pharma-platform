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
	cd project && go run cmd/dev-mode/main.go

api:
	cd project && go run cmd/api/main.go

sim:
	cd project && go run cmd/collector-sim/main.go

seed:
	cd project && go run cmd/seed/main.go

build:
	cd project && go build ./...

prod:
	cd project && go run cmd/pharma-platform/main.go

.PHONY: setup up up-all down logs dev api sim seed build prod

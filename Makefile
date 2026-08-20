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

# Simulation mode: full dashboard + a built-in simulation engine that
# generates realistic telemetry for all seeded machines (no real PLCs
# needed). Use this to see OEE/production/analytics working end-to-end.
simulate:
	$(PORT_CHECK)
	cd project && go build -o dev-mode cmd/dev-mode/main.go
	cd project && { ./dev-mode & DEV_PID=$$!; }; \
		cd ../web && npm run dev; \
		STATUS=$$?; \
		kill $$DEV_PID 2>/dev/null; \
		wait $$DEV_PID 2>/dev/null; \
		exit $$STATUS

# Real mode: full dashboard + backend with no pre-configured machines or
# simulated data. Machines are added and wired to drivers from the
# dashboard itself.
real:
	$(PORT_CHECK)
	cd project && go build -o pharma-platform cmd/pharma-platform/main.go
	cd project && { ./pharma-platform & API_PID=$$!; }; \
		cd ../web && npm run dev; \
		STATUS=$$?; \
		kill $$API_PID 2>/dev/null; \
		wait $$API_PID 2>/dev/null; \
		exit $$STATUS

# Standalone fake-telemetry writer with no API/dashboard attached —
# a lower-level utility for exercising the collector/writer pipeline
# in isolation, not one of the two primary run modes above.
collector-sim:
	cd project && go build -o collector-sim cmd/collector-sim/collector-sim.go && exec ./collector-sim

migrate:
	cd project && go run cmd/migrate/main.go

seed:
	cd project && go run cmd/seed/main.go

build:
	cd project && go build ./...

.PHONY: setup up up-all down logs simulate real collector-sim migrate seed build

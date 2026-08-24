#!/usr/bin/env bash
# Single entrypoint: installs dependencies, ensures Docker is running,
# starts Postgres + QuestDB, and launches either the "simulate" or "real"
# run mode. See README.md for what each mode does.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT"

echo "==> Checking Docker..."
if ! command -v docker >/dev/null 2>&1; then
	echo "Docker is not installed. Install Docker (and the Compose plugin), then re-run this script." >&2
	exit 1
fi

if ! docker info >/dev/null 2>&1; then
	echo "Docker daemon is not running."
	if command -v systemctl >/dev/null 2>&1; then
		echo "Attempting to start it via systemctl (this may prompt for your password)..."
		if ! sudo systemctl start docker; then
			echo "Could not start Docker automatically. Start Docker Desktop / the docker service yourself and re-run this script." >&2
			exit 1
		fi
	else
		echo "Start Docker manually and re-run this script." >&2
		exit 1
	fi
fi

if ! command -v go >/dev/null 2>&1; then
	echo "Go is not installed. Install Go, then re-run this script." >&2
	exit 1
fi

if ! command -v npm >/dev/null 2>&1; then
	echo "npm is not installed. Install Node.js/npm, then re-run this script." >&2
	exit 1
fi

echo "==> Installing Go dependencies..."
(cd project && go mod download)

echo "==> Installing frontend dependencies..."
(cd web && npm install)

echo "==> Starting Postgres + QuestDB..."
echo "    (Postgres is mapped to host port 5433, not 5432, so it won't clash"
echo "     with a system-installed Postgres — see project/runtime/docker-compose.yml)"
make up

MODE="${1:-}"
if [ -z "$MODE" ]; then
	echo
	echo "Select a run mode:"
	echo "  1) simulate — full dashboard + built-in simulation engine (no real PLCs needed)"
	echo "  2) real     — full dashboard + backend, no pre-configured machines"
	read -rp "Enter 1 or 2: " choice
	case "$choice" in
	1) MODE="simulate" ;;
	2) MODE="real" ;;
	*)
		echo "Invalid choice." >&2
		exit 1
		;;
	esac
fi

case "$MODE" in
simulate | real)
	exec make "$MODE"
	;;
*)
	echo "Usage: $0 [simulate|real]" >&2
	exit 1
	;;
esac

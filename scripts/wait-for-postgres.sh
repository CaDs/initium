#!/usr/bin/env bash
# wait-for-postgres.sh — used by `make setup` to block until Postgres is
# ready, with a bounded timeout and clear diagnostics instead of the
# infinite polling loop the setup target used to contain.
#
# Env:
#   TIMEOUT_SECONDS — max wait (default 60)
#   POSTGRES_SERVICE — compose service name (default postgres)
#   POSTGRES_USER — Postgres role to check with pg_isready (default initium)

set -euo pipefail

TIMEOUT_SECONDS="${TIMEOUT_SECONDS:-60}"
POSTGRES_SERVICE="${POSTGRES_SERVICE:-postgres}"
POSTGRES_USER="${POSTGRES_USER:-initium}"

red() { printf "\033[31m%s\033[0m\n" "$*" >&2; }
yellow() { printf "\033[33m%s\033[0m\n" "$*"; }
green() { printf "\033[32m%s\033[0m\n" "$*"; }

# 1. Is Docker even reachable?
if ! docker info >/dev/null 2>&1; then
  red "Docker daemon is not reachable."
  red "Start Docker Desktop (or your Docker runtime) and re-run 'make setup'."
  exit 1
fi

# 2. Is the postgres container running? If not, try to bring it up.
if [[ -z "$(docker compose ps -q "$POSTGRES_SERVICE" 2>/dev/null)" ]]; then
  yellow "Postgres container not running — starting it."
  docker compose up -d postgres mailpit
fi

# 3. Poll pg_isready with a bounded timeout.
yellow "Waiting for Postgres (up to ${TIMEOUT_SECONDS}s)..."
elapsed=0
while ! docker compose exec -T "$POSTGRES_SERVICE" pg_isready -U "$POSTGRES_USER" >/dev/null 2>&1; do
  if [[ "$elapsed" -ge "$TIMEOUT_SECONDS" ]]; then
    red "Postgres did not become ready within ${TIMEOUT_SECONDS}s."
    red "Container status:"
    docker compose ps "$POSTGRES_SERVICE" >&2 || true
    red "Recent logs (tail):"
    docker compose logs --tail 30 "$POSTGRES_SERVICE" >&2 || true
    exit 1
  fi
  sleep 1
  elapsed=$((elapsed + 1))
done

green "Postgres is ready."

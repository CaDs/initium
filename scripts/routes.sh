#!/usr/bin/env bash
# routes.sh — curl /_debug/routes on the running backend and pretty-print.
# Requires `make dev:backend` (or `make dev`) to be running.

set -euo pipefail

BACKEND_URL="${BACKEND_URL:-http://localhost:8000}"
ENDPOINT="$BACKEND_URL/_debug/routes"

if ! response=$(curl -fsS "$ENDPOINT" 2>/dev/null); then
  echo "Failed to reach $ENDPOINT." >&2
  echo "Is the backend running? Try 'make dev:backend'." >&2
  exit 1
fi

if command -v jq >/dev/null 2>&1; then
  echo "$response" | jq -r '.routes | sort_by(.pattern, .method)[] | "\(.method)\t\(.pattern)"' | column -t
else
  echo "$response"
fi

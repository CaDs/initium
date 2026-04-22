#!/usr/bin/env bash
# status.sh — report which Initium services are reachable.

set -uo pipefail

check() {
  local name="$1" url="$2"
  if curl -fsS -o /dev/null "$url" 2>/dev/null; then
    printf "  \033[32m✓\033[0m %-20s %s\n" "$name" "$url"
  else
    printf "  \033[31m✗\033[0m %-20s %s\n" "$name" "$url"
  fi
}

echo "Initium services:"
check "Backend /healthz"   "http://localhost:8000/healthz"
check "Backend /readyz"    "http://localhost:8000/readyz"
check "Web (Next.js)"      "http://localhost:3000"
check "Mailpit UI"         "http://localhost:8025"

echo ""
echo "Docker compose:"
docker compose ps --format "  {{.Service}}\t{{.State}}" 2>/dev/null || echo "  (compose not running)"

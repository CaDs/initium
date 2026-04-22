#!/usr/bin/env bash
# schemathesis.sh — run Schemathesis contract tests against a running backend.
#
# Usage:
#   scripts/schemathesis.sh                          # against http://localhost:8000
#   BASE_URL=http://api:8000 scripts/schemathesis.sh # against custom URL
#
# Requires: docker. Backend must be running and reachable from the container.
#
# Notes:
#   - Magic-link and OAuth callback are excluded: they return 307 redirects
#     with non-JSON bodies and require a live email/Google side channel.
#   - Admin and /me require auth; set AUTH_HEADER to include them, e.g.:
#       AUTH_HEADER="Authorization: Bearer $DEV_TOKEN" scripts/schemathesis.sh
#
# This script is intentionally NOT wired into the default CI path. Contract
# tests at the Go layer (backend/internal/app/contract_test.go) catch
# route↔spec drift cheaply. Schemathesis is heavier; run it nightly, on
# main merges, or locally before a release.

set -euo pipefail

BASE_URL="${BASE_URL:-http://host.docker.internal:8000}"
SPEC_PATH="${SPEC_PATH:-/spec/openapi.yaml}"
AUTH_HEADER="${AUTH_HEADER:-}"

HEADER_ARGS=()
if [[ -n "$AUTH_HEADER" ]]; then
  HEADER_ARGS=(-H "$AUTH_HEADER")
fi

docker run --rm \
  -v "$(pwd)/backend/api:/spec:ro" \
  --add-host=host.docker.internal:host-gateway \
  schemathesis/schemathesis:stable run \
  --checks all \
  --url "$BASE_URL" \
  --exclude-path-regex '^/(api/auth/(google|google/callback|verify|magic-link)|_debug/)' \
  "${HEADER_ARGS[@]}" \
  "$SPEC_PATH"

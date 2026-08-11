#!/usr/bin/env bash
# schemathesis.sh — run Schemathesis contract tests against a running backend.
#
# Usage:
#   DEV_BYPASS_AUTH=false make dev/backend            # separate terminal
#   scripts/schemathesis.sh                          # against http://localhost:8000
#   BASE_URL=http://api:8000 scripts/schemathesis.sh # against custom URL
#   SCHEMATHESIS_IMAGE=... scripts/schemathesis.sh   # override pinned image
#
# Requires: docker. Backend must be running and reachable from the container.
#
# Notes:
#   - Auth endpoints are excluded because they are rate-limited and depend on
#     cookies, email, or Google. Their behavior is covered by Go tests.
#   - Admin and /me accept an optional auth header, e.g.:
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
SCHEMATHESIS_IMAGE="${SCHEMATHESIS_IMAGE:-schemathesis/schemathesis:v4.24.3}"

SCHEMATHESIS_ARGS=(
  run
  --checks all
  --exclude-checks ignored_auth,missing_required_header,unsupported_method
  --url "$BASE_URL"
  --exclude-path-regex '^/(api/auth/|_debug/)'
)
if [[ -n "$AUTH_HEADER" ]]; then
  SCHEMATHESIS_ARGS+=(-H "$AUTH_HEADER")
fi
SCHEMATHESIS_ARGS+=("$SPEC_PATH")

docker run --rm \
  -v "$(pwd)/backend/api:/spec:ro" \
  --add-host=host.docker.internal:host-gateway \
  "$SCHEMATHESIS_IMAGE" \
  "${SCHEMATHESIS_ARGS[@]}"

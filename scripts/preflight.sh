#!/usr/bin/env bash
# preflight.sh — all gates a PR must pass, same order as CI.
# Runs lint, fast tests, and OpenAPI contract check.

set -euo pipefail

red() { printf "\033[31m%s\033[0m\n" "$*"; }
green() { printf "\033[32m%s\033[0m\n" "$*"; }
step() { printf "\n\033[1;36m▸ %s\033[0m\n" "$*"; }

trap 'red "preflight FAILED"' ERR

step "lint (backend + web + mobile, parallel)"
make lint

step "test (backend + web + mobile, parallel)"
make test

step "check:openapi (dart DTO drift)"
make check:openapi

step "check:skills (exemplar path drift)"
bash scripts/check-skills.sh

green "preflight PASSED"

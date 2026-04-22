#!/usr/bin/env bash
# preflight.sh — every gate a PR must pass, same order CI runs them.
# Designed for agent self-check: run this before declaring a feature done.

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

step "check:parity (every /api/ spec path has a consumer)"
make check:parity

step "check:skills (exemplar path + symbol drift)"
bash scripts/check-skills.sh

step "check:staged (no untracked or unstaged files)"
bash scripts/check-staged.sh

green "preflight PASSED"

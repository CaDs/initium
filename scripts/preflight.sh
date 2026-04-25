#!/usr/bin/env bash
# preflight.sh — every gate a PR must pass, same order CI runs them.
# Designed for agent self-check: run this before declaring a feature done.
#
# Native mobile (iOS / Android) tests and linters run separately via
# `make test:ios`, `make test:android`, `make lint:ios`, `make lint:android`
# because they require Xcode + a simulator (iOS) or a JDK + Gradle
# (Android) that aren't guaranteed in every environment.

set -euo pipefail

red() { printf "\033[31m%s\033[0m\n" "$*"; }
green() { printf "\033[32m%s\033[0m\n" "$*"; }
step() { printf "\n\033[1;36m▸ %s\033[0m\n" "$*"; }

trap 'red "preflight FAILED"' ERR

step "lint (backend + web, parallel)"
make lint

step "test (backend + web, parallel)"
make test

step "check:parity (every /api/ spec path has a web consumer; mobile is paused)"
make check:parity

step "check:skills (exemplar path + symbol drift)"
bash scripts/check-skills.sh

step "check:staged (no untracked or unstaged files)"
bash scripts/check-staged.sh

green "preflight PASSED"

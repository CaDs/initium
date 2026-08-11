#!/usr/bin/env bash
# check-staged.sh — fail if git has untracked or unstaged changes.
# Catches the round-2 failure mode: builder creates new files but forgets
# `git add -A` before declaring "done". Runs as the last step of preflight
# so you know the PR you're about to push matches what you tested.

set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

fail=0

if ! git diff --quiet; then
  printf "\033[31munstaged changes:\033[0m\n"
  git diff --name-only | sed 's/^/  /'
  fail=1
fi

untracked=$(git ls-files --others --exclude-standard)
if [[ -n "$untracked" ]]; then
  printf "\033[31muntracked files:\033[0m\n"
  echo "$untracked" | sed 's/^/  /'
  fail=1
fi

if [[ "$fail" -ne 0 ]]; then
  printf "\n\033[31mstage every file you expect to ship (git add -A) and re-run preflight.\033[0m\n"
  exit 1
fi

echo "worktree changes fully staged: ok"

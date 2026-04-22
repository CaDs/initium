#!/usr/bin/env bash
# check-staged.sh — fail if git has untracked or unstaged changes.
# Catches the round-2 failure mode: builder creates new files but forgets
# `git add -A` before declaring "done". Runs as the last step of preflight
# so you know the PR you're about to push matches what you tested.

set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

status=$(git status --porcelain)
if [[ -z "$status" ]]; then
  echo "git tree clean: ok"
  exit 0
fi

printf "\033[31muntracked or unstaged changes:\033[0m\n"
echo "$status" | sed 's/^/  /'
printf "\n\033[31mstage every file you expect to ship (git add -A) and re-run preflight.\033[0m\n"
exit 1

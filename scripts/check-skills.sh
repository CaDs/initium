#!/usr/bin/env bash
# check-skills.sh — verify exemplar file paths referenced in each SKILL.md
# actually resolve. Protects against the skills drifting from the code over
# time (e.g., a referenced file gets renamed or deleted).

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

fail=0

check_skill() {
  local skill="$1"
  local dir="$ROOT/.claude/skills/$skill"

  if [[ ! -f "$dir/SKILL.md" ]]; then
    echo "MISSING: $dir/SKILL.md" >&2
    fail=1
    return
  fi

  # Extract backtick-quoted paths that look like repo paths (start with
  # backend/, web/, mobile/, or docs/). Strip URL fragments. Dedup.
  local refs
  refs=$(
    grep -hoE '`(backend|web|mobile|docs)/[^`[:space:]]+`' \
      "$dir/SKILL.md" "$dir/patterns"/*.md 2>/dev/null \
      | sed -E 's/^`//; s/`$//; s/#.*$//' \
      | sort -u
  )

  while IFS= read -r path; do
    [[ -z "$path" ]] && continue
    # Skip glob-ish references, trailing-slash dirs
    if [[ "$path" == *"*"* || "$path" == */ ]]; then
      continue
    fi
    if [[ ! -e "$ROOT/$path" ]]; then
      printf "\033[31mBROKEN\033[0m %s referenced in %s\n" "$path" "$skill" >&2
      fail=1
    fi
  done <<< "$refs"
}

for skill in initium-backend initium-web initium-mobile; do
  check_skill "$skill"
done

if [[ ! -f "$ROOT/.claude/skills/_shared/parity.md" ]]; then
  echo "MISSING: .claude/skills/_shared/parity.md" >&2
  fail=1
fi

if [[ "$fail" -eq 0 ]]; then
  echo "skills check: ok"
else
  exit 1
fi

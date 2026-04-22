#!/usr/bin/env bash
# check-skills.sh — verify exemplar file paths referenced in each SKILL.md
# actually resolve AND optionally contain the symbol the skill claims.
#
# Path references that look like `backend/...` / `web/...` / `mobile/...` /
# `docs/...` are checked for existence. To also grep for an expected symbol
# inside the file (catches renames that preserve the filename but rename the
# thing the skill points to), annotate with an HTML comment:
#
#     - Handler: `backend/internal/adapter/handler/user.go` <!-- expect: writeUser -->
#
# The script extracts each `path + expect` pair and greps the path for the
# symbol; failure produces a clear error.

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

  # Collect all .md files under the skill (SKILL.md + patterns/*.md).
  local files=()
  files+=("$dir/SKILL.md")
  if [[ -d "$dir/patterns" ]]; then
    while IFS= read -r f; do files+=("$f"); done < <(find "$dir/patterns" -type f -name "*.md")
  fi

  # 1. Path existence check: only against SKILL.md (patterns/*.md contains
  #    tutorial paths that are intentionally illustrative, e.g. `order.go`).
  #    Paths must start with a known prefix AND contain `/`.
  local refs
  refs=$(
    grep -hoE '`(backend|web|mobile|docs|scripts)/[a-zA-Z0-9_./:\-]+`' "$dir/SKILL.md" 2>/dev/null \
      | sed -E 's/^`//; s/`$//; s/#.*$//' \
      | sort -u
  )

  while IFS= read -r path; do
    [[ -z "$path" ]] && continue
    if [[ "$path" == *"*"* || "$path" == */ ]]; then
      continue
    fi
    if [[ ! -e "$ROOT/$path" ]]; then
      printf "\033[31mBROKEN\033[0m  %s referenced in %s\n" "$path" "$skill" >&2
      fail=1
    fi
  done <<< "$refs"

  # 2. Symbol check: lines matching `path` <!-- expect: symbol --> must grep
  #    the symbol inside the file. Runs on SKILL.md AND patterns/*.md —
  #    explicit annotations are always checked.
  while IFS= read -r line; do
    [[ -z "$line" ]] && continue
    local path symbol
    path=$(printf '%s' "$line" | sed -E 's/.*`([^`]+)` *<!-- *expect: *([^ ]+) *-->.*/\1/')
    symbol=$(printf '%s' "$line" | sed -E 's/.*`([^`]+)` *<!-- *expect: *([^ ]+) *-->.*/\2/')
    if [[ -z "$path" || -z "$symbol" || "$path" == "$line" ]]; then
      continue
    fi
    path="${path#./}"; path="${path%#*}"
    if [[ ! -f "$ROOT/$path" ]]; then
      # already flagged above as BROKEN; skip
      continue
    fi
    if ! grep -q -F -- "$symbol" "$ROOT/$path"; then
      printf "\033[31mSTALE\033[0m   %s no longer contains \"%s\" (referenced in %s)\n" \
        "$path" "$symbol" "$skill" >&2
      fail=1
    fi
  done < <(grep -hE '`[^`]+` *<!-- *expect: *[^ ]+ *-->' "${files[@]}" 2>/dev/null || true)
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

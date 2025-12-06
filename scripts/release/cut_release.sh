#!/usr/bin/env bash
set -euo pipefail

# Interactive release helper:
# - Suggests version (minor vs patch) via OpenAI
# - Generates release notes + changelog entry
# - Updates CHANGELOG.md
# - On confirmation: git commit, tag, and run goreleaser release --clean
#
# Requirements: OPENAI_API_KEY set; goreleaser configured; git clean/staged as desired.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
NOTES_DIR="$ROOT_DIR/notes"
CHANGELOG="$ROOT_DIR/CHANGELOG.md"

prev_tag="${PREV_TAG:-}"
version_override="${VERSION:-}"
extra_hint="${FEEDBACK:-}"

if [[ -z "$prev_tag" ]]; then
  prev_tag=$(git -C "$ROOT_DIR" describe --tags --abbrev=0 2>/dev/null || true)
fi

echo "Previous tag: ${prev_tag:-none}"

suggest_json=$(python "$ROOT_DIR/scripts/release/suggest_version.py" --prev-tag "$prev_tag" 2>/dev/null || true)
suggest_version=""
suggest_reason=""
if [[ -n "$suggest_json" ]]; then
  suggest_version=$(echo "$suggest_json" | python -c "import sys,json; d=json.load(sys.stdin); print(d.get('version',''))" 2>/dev/null || true)
  suggest_reason=$(echo "$suggest_json" | python -c "import sys,json; d=json.load(sys.stdin); print(d.get('reason',''))" 2>/dev/null || true)
fi

if [[ -z "$suggest_version" ]]; then
  suggest_version="v0.1.0"
  suggest_reason="LLM unavailable; defaulted to v0.1.0"
fi

echo "Suggested version: $suggest_version"
[[ -n "$suggest_reason" ]] && echo "Reason: $suggest_reason"

# Always preview release notes for the suggested version before choosing
preview=$(python "$ROOT_DIR/scripts/release/generate_notes.py" \
  --version "$suggest_version" \
  --prev-tag "$prev_tag" \
  ${extra_hint:+--extra "$extra_hint"} 2>/dev/null || true)
if [[ -n "$preview" ]]; then
  release_preview=$(echo "$preview" | awk '/===RELEASE_NOTES===/{flag=1;next}/===CHANGELOG_ENTRY===/{flag=0}flag')
  echo "=== Suggested Release Notes Preview ==="
  if [[ -n "$release_preview" ]]; then
    echo "$release_preview"
  else
    echo "$preview"
  fi
  echo "======================================="
else
  echo "Release notes preview unavailable (LLM error)."
fi

read -r -p "Use this version? [Y/n/other]: " ans
if [[ "${ans,,}" == "n" || "${ans,,}" == "no" ]]; then
  read -r -p "Enter version (e.g., v1.2.3): " version_override
elif [[ -n "$ans" && "${ans,,}" != "y" && "${ans,,}" != "yes" ]]; then
  version_override="$ans"
fi

VERSION="${version_override:-$suggest_version}"
echo "Releasing version: $VERSION"

mkdir -p "$NOTES_DIR"
release_out="$NOTES_DIR/release_${VERSION}.md"
changelog_snippet="$NOTES_DIR/changelog_${VERSION}.md"

python "$ROOT_DIR/scripts/release/generate_notes.py" \
  --version "$VERSION" \
  --prev-tag "$prev_tag" \
  --release-out "$release_out" \
  --changelog-out "$changelog_snippet" \
  ${extra_hint:+--extra "$extra_hint"}

echo "Generated:"
echo "  Release notes:    $release_out"
echo "  Changelog entry:  $changelog_snippet"
echo
sed -n '1,120p' "$release_out" || true
echo

read -r -p "Insert changelog entry into CHANGELOG.md? [Y/n]: " ins
if [[ -z "$ins" || "${ins,,}" == "y" || "${ins,,}" == "yes" ]]; then
  python - "$CHANGELOG" "$changelog_snippet" <<'PY'
import sys
from pathlib import Path

changelog = Path(sys.argv[1])
snippet = Path(sys.argv[2]).read_text()

if not changelog.exists():
    changelog.write_text("# Changelog\n\n" + snippet + "\n")
    sys.exit(0)

text = changelog.read_text()
marker = "## [Unreleased]"
if marker in text:
    parts = text.split(marker, 1)
    new_text = f"{parts[0]}{marker}\n\n{snippet}\n{parts[1]}"
else:
    if text.lstrip().startswith("# Changelog"):
        lines = text.splitlines()
        lines.insert(1, "")
        lines.insert(2, snippet)
        new_text = "\n".join(lines) + "\n"
    else:
        new_text = "# Changelog\n\n" + snippet + "\n\n" + text
changelog.write_text(new_text)
PY
  echo "CHANGELOG.md updated."
else
  echo "Skipped changelog insertion."
fi

echo
read -r -p "Commit, tag ($VERSION), and run goreleaser release --clean? [y/N]: " go
if [[ "${go,,}" == "y" || "${go,,}" == "yes" ]]; then
  git -C "$ROOT_DIR" add "$CHANGELOG" "$release_out" "$changelog_snippet" 2>/dev/null || true
  git -C "$ROOT_DIR" commit -m "Release $VERSION"
  git -C "$ROOT_DIR" tag "$VERSION"
  (cd "$ROOT_DIR" && goreleaser release --clean)
  echo "Release process kicked off. Verify goreleaser output."
else
  echo "Skipped commit/tag/release. Files remain for manual review."
fi

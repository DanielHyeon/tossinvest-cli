#!/usr/bin/env bash
set -uo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
QUERY="${1:-}"
[ -n "$QUERY" ] || { echo "usage: $0 \"<query>\"" >&2; exit 2; }

PYTHON_BIN="${PYTHON:-python3}"
echo "[memory-recall] file primary: \"$QUERY\""
"$PYTHON_BIN" "$ROOT/scripts/memory_index.py" search "$QUERY" || true

if command -v gbrain >/dev/null 2>&1; then
  echo "[memory-recall] gbrain advisory: \"$QUERY\""
  "$PYTHON_BIN" "$ROOT/tools/sdd/gbrain_project.py" search "$QUERY" 2>/dev/null || true
fi

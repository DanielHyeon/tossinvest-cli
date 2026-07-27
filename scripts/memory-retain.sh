#!/usr/bin/env bash
set -uo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
EPISODE="${1:-}"
[ -n "$EPISODE" ] || { echo "usage: $0 <episode.md>" >&2; exit 2; }
[ -f "$EPISODE" ] || { echo "episode not found: $EPISODE" >&2; exit 2; }

"${PYTHON:-python3}" "$ROOT/scripts/memory_index.py" retain "$EPISODE"

#!/bin/bash
set -euo pipefail
ROOT=/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/codex-tree-85-r2
label=$1
shift
cd "$ROOT/review-r2/$label"
GOCACHE="$ROOT/.gocache" GOTMPDIR="$ROOT/.gotmp" GOFLAGS=-mod=mod GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local TMPDIR="$ROOT/.gotmp" GOTELEMETRY=off go test "$@"

#!/usr/bin/env bash
# a113 뮤테이션 하네스 — 저장소가 아니라 사본에서만 변이한다.
# 사용: run.sh <name> <python-mutation-snippet-file|none>
set -euo pipefail
H0=$(cd "$(dirname "$0")" && pwd)
REPO=$(git -C "$H0" rev-parse --show-toplevel)
H=$(cd "$(dirname "$0")" && pwd)
name=$1; mut=$2
W=${TMPDIR:-/tmp}/a113-mut-$name-$$
rm -rf "$W"; mkdir -p "$W/internal"
cp "$REPO/go.mod" "$REPO/go.sum" "$W/"
cp -r "$REPO/internal/strategyprojection" "$REPO/internal/strategyprojectionrpc" "$W/internal/"
test "$(git -C "$REPO" rev-parse --show-toplevel)" = "$REPO"
if [ "$mut" != none ]; then
  python3 "$H/$mut" "$W/internal/strategyprojectionrpc"
fi
cd "$W"
set +e
GOFLAGS=-trimpath go test -count=1 ./internal/strategyprojectionrpc/ > "${TMPDIR:-/tmp}/a113-mut-$name.log" 2>&1
rc=$?
set -e
rm -rf "$W"
echo "$name rc=$rc"
grep -E "^(--- FAIL|    --- FAIL)" "${TMPDIR:-/tmp}/a113-mut-$name.log" | head -8 || true

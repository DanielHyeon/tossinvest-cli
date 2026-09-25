#!/usr/bin/env bash
# a114 뮤테이션 하네스 — 저장소가 아니라 사본(go.mod·cmd·internal)에서만 변이한다.
# 사용: run.sh <name> <mutation.py|none>   (TMPDIR 을 여유 있는 파일시스템에 두라)
set -euo pipefail
H=$(cd "$(dirname "$0")" && pwd)
REPO=$(git -C "$H" rev-parse --show-toplevel)
name=$1; mut=$2
W=${TMPDIR:-/tmp}/a114-mut-$name-$$
rm -rf "$W"; mkdir -p "$W"
cp "$REPO/go.mod" "$REPO/go.sum" "$W/"
cp -r "$REPO/cmd" "$REPO/internal" "$W/"
if [ "$mut" != none ]; then
  python3 "$H/$mut" "$W/cmd/tossctl"
fi
cd "$W"
set +e
GOFLAGS=-trimpath go test -count=1 -run "$(cat "$H/tests.txt")" ./cmd/tossctl/ > "${TMPDIR:-/tmp}/a114-mut-$name.log" 2>&1
rc=$?
set -e
rm -rf "$W"
echo "$name rc=$rc"
grep -E "^(--- FAIL|    --- FAIL)" "${TMPDIR:-/tmp}/a114-mut-$name.log" | head -6 || true

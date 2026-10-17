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
# 판정은 rc 가 아니라 출력이다(post-review P2-3): TMPDIR 정리 실패(`unlinkat … directory not empty`)가
# 무변이 대조군도 rc=1 로 만든 적이 있다. FAIL 줄이 있으면 CAUGHT, `ok` 줄이 있으면 SURVIVED, 둘 다 없으면 BROKEN.
log="${TMPDIR:-/tmp}/a114-mut-$name.log"
if grep -qE "^(--- FAIL|    --- FAIL)" "$log"; then verdict=CAUGHT
elif grep -qE "^ok[[:space:]]" "$log"; then verdict=SURVIVED
else verdict=BROKEN; fi
echo "$name rc=$rc verdict=$verdict"
grep -E "^(--- FAIL|    --- FAIL)" "$log" | head -6 || true

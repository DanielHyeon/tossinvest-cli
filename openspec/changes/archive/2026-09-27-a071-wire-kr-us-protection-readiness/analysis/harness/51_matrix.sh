#!/bin/bash
# task 5.1 하네스: 패키지 하나를 -cover 로 한 번 빌드하고 시험 함수마다 따로 돌려 커버 프로필을 남김.
# 목적: 분기 본문 블록을 **어느 시험이 실행하는지** 손으로 고르지 않고 측정으로 댐.
# 사용: [A071_REPO=<트리>] 51_matrix.sh <패키지 상대 경로> <출력 디렉터리>
#   예) 51_matrix.sh internal/execgw /somewhere/matrix
# 주의: TMPDIR 은 ext4 여야 함 — fuseblk(/mnt/D) 위 t.TempDir 은 chmod 0700 이 안 먹어
#       protectionreadiness 의 소유권 검사 시험 8개가 헛실패함(2026-09-25 실측).
set -euo pipefail
# 깨끗한 트리에서 재려면 A071_REPO 로 격리 워크트리를 지정함(남의 미추적 시험이 모집단에 섞이지 않게).
REPO=${A071_REPO:-$(git -C "$(dirname "$0")" rev-parse --show-toplevel)}
pkg=$1; tag=${pkg//\//_}
OUT=$2/$tag
rm -rf "$OUT"; mkdir -p "$OUT/prof"
cd "$REPO"
go test -c -cover -o "$OUT/test.bin" "./$pkg"
cd "$REPO/$pkg"
"$OUT/test.bin" -test.list '.*' 2>/dev/null | grep -E '^(Test|Example|Fuzz)' > "$OUT/tests.txt"
run_one() {
  t=$1
  if "$OUT/test.bin" -test.run "^${t}\$" -test.count=1 -test.coverprofile="$OUT/prof/$t.out" > "$OUT/prof/$t.log" 2>&1; then echo "PASS $t"; else echo "FAIL $t"; fi
}
export -f run_one; export OUT
xargs -P 4 -I{} bash -c 'run_one {}' < "$OUT/tests.txt" | sort > "$OUT/results.txt"
echo "tests=$(wc -l < "$OUT/tests.txt") pass=$(grep -c '^PASS' "$OUT/results.txt") fail=$(grep -c '^FAIL' "$OUT/results.txt" || true)"

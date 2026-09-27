#!/usr/bin/env bash
# a112 8.7.2 — 테스트별 커버리지 프로파일 수집 하네스.
#
# 쓰는 법: a872_pertest_cover.sh <engine.test 바이너리> <출력 디렉터리>
#   바이너리는 저장소 루트에서 다음으로 만든다(-trimpath 금지: 소스를 읽는 시험이 깨진다):
#     go test -c -tags tossos_testseams -covermode=count \
#       -coverpkg=./internal/app/engine,./internal/strategyrouter -o <바이너리> ./internal/app/engine/
#
# 하는 일: 바이너리의 전체 시험 목록을 읽어 시험마다 `-test.run '^<Test>$'` 로 하나씩 돌리고
# `<출력>/<Test>.cov` 를 남긴다. 스위트 전체 프로파일(`<출력>/suite.cov`)도 함께 뜬다 —
# 귀속 완전성(시험별 합 == 스위트)은 이 둘로 a872_attribute.py 가 등식으로 잰다.
#
# 메모리: 엔진 스위트가 묶지 않으면 커널 OOM 을 낸 적이 있어 cgroup 안에서 돈다.
set -euo pipefail
bin=$(realpath "$1")
out=$(realpath -m "$2")
repo=$(git -C "$(dirname "$0")" rev-parse --show-toplevel)
mkdir -p "$out"
cd "$repo/internal/app/engine"
run() { systemd-run --user --scope -q -p MemoryMax=16G -p MemorySwapMax=0 "$@"; }
run "$bin" -test.coverprofile="$out/suite.cov" -test.timeout=40m > "$out/suite.log" 2>&1 || { echo "suite FAIL"; exit 1; }
"$bin" -test.list '.*' 2>/dev/null | grep '^Test' > "$out/tests.txt"
fail=0
while read -r name; do
  if ! run "$bin" -test.run "^${name}\$" -test.coverprofile="$out/${name}.cov" -test.timeout=10m > "$out/${name}.log" 2>&1; then
    echo "FAIL $name"; fail=1
  fi
done < "$out/tests.txt"
echo "tests=$(wc -l < "$out/tests.txt") fail=$fail"
exit "$fail"

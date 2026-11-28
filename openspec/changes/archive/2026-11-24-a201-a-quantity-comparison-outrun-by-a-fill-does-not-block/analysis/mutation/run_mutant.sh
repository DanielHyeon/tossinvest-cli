#!/usr/bin/env bash
# a201 2.5 변이 하네스 — 사본 worktree 에서 a201 시험(세 패키지 TestA201)을 돌려 CAUGHT 와 reached 를 따로 잼.
#
#   CAUGHT  = 실패한 시험이 하나라도 있음(build failed 포함 — 그때는 reached 를 잴 수 없어 「?」).
#   reached = 변이 줄이 커버 프로파일(mode set, 세 패키지 합집합)에서 실행됨.
#
# 사용: run_mutant.sh <사본 worktree 절대경로> <이름> <변이 파일(저장소 상대)> <변이 줄> <출력 디렉터리>
# 사본은 한 번에 한 판만 씀(두 판이 한 사본을 쓰면 서로의 변이가 기준이 됨). 원복과 원복 확인은 부르는 쪽이 함.
set -uo pipefail
copy=$1; label=$2; file=$3; line=$4; out=$5
if [ "$(git -C "$copy" rev-parse --show-toplevel)" != "$copy" ]; then
  echo "not a worktree toplevel: $copy" >&2
  exit 2
fi
mkdir -p "$out"
log="$out/$label.txt"
prof="$out/$label.cover"
rm -f "$prof".*
{
  echo "# label=$label file=$file line=$line head=$(git -C "$copy" rev-parse HEAD)"
  git -C "$copy" diff --stat
  git -C "$copy" diff
} > "$log" 2>&1
rc_all=0
for pkg in ./internal/app/engine ./internal/reconcile ./internal/journal; do
  name=$(basename "$pkg")
  (cd "$copy" && rtk proxy go test -count=1 -timeout 30m -v -run 'TestA201' -covermode=set \
    -coverpkg=./internal/app/engine,./internal/reconcile,./internal/journal \
    -coverprofile="$prof.$name" "$pkg") >> "$log" 2>&1 || rc_all=1
done
fails=$(grep -c '^--- FAIL' "$log")
build=$(grep -c 'build failed\|setup failed' "$log")
if [ "$build" -gt 0 ]; then
  reached="?"
else
  reached=$(cat "$prof".* 2>/dev/null | awk -v f="/$file:" -v L="$line" '
    index($0, f) { split($1, a, ":"); split(a[2], r, ","); split(r[1], s, "."); split(r[2], e, ".");
      if (s[1] <= L && L <= e[1] && $3 > 0) hit = 1 }
    END { print (hit ? "yes" : "no") }')
fi
echo "$label rc=$rc_all fail_tests=$fails build_fail=$build reached=$reached"
grep '^--- FAIL' "$log" | sed 's/ (.*//' | sort -u | sed "s/^/  $label /"

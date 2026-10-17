#!/usr/bin/env bash
# a066 6.1 — 전체 스위트·race·속성/fuzz 를 격리 사본에서 **순차로** 돌리고 단계별 rc·시간을 원장에 남김.
#
# 왜 사본인가: 병행 로트의 미추적 파일이 같은 패키지를 깨뜨릴 수 있음(test_in_copy.sh 머리 주석).
# 왜 df 를 먼저 보나: 디스크 가득 참이 시험 실패·변이 CAUGHT 를 부풀린 선례(a122 7.5.35) — 여유가 모자라면 시작하지 않음.
#
# 사용: run_6_1.sh <scratch-dir> [min-free-GB]
# 결과: <scratch-dir>/run61-<pid>/ledger.tsv (단계 · rc · 초 · 로그 경로) + 단계별 로그(rtk 요약 없이 원문).
set -euo pipefail
scratch=${1:?scratch dir}
min_free=${2:-15}
free_gb=$(df -BG --output=avail / | tail -1 | tr -dc '0-9')
if [ "$free_gb" -lt "$min_free" ]; then
  echo "refuse: / has ${free_gb}G free, need ${min_free}G" >&2
  exit 3
fi
root=$(git rev-parse --show-toplevel)
harness="$root/openspec/changes/a066-add-multi-horizon-risk-buckets/analysis/harness"
out="$scratch/run61-$$"
copy="$out/copy"
mkdir -p "$out"
ledger="$out/ledger.tsv"
printf 'TREE\t%s\tuncommitted=%s\tfree_gb=%s\n' "$(git -C "$root" rev-parse --short HEAD)" \
  "$(git -C "$root" status --porcelain -- internal cmd tools go.mod go.sum | wc -l)" "$free_gb" > "$ledger"
# 사본만 만들고 시험은 돌리지 않음(-run 에 맞는 이름 없음).
bash "$harness/test_in_copy.sh" "$copy" -- -count=1 -run '^$' ./internal/riskbucket > "$out/copy.log" 2>&1

# 단계 하나를 사본에서 돌리고 rc·시간을 기록함. rc 는 파이프 없이 직접 받음.
step() {
  local name=$1; shift
  # STEPS 가 주어지면 그 이름(정규식)에 맞는 단계만 돌림 — 한 단계 재실행용.
  if [ -n "${STEPS:-}" ] && ! [[ $name =~ ^(${STEPS})$ ]]; then return 0; fi
  local log="$out/$name.log" start rc
  start=$(date +%s)
  set +e
  ( cd "$copy" && "$@" ) > "$log" 2>&1
  rc=$?
  set -e
  printf '%s\t%s\t%s\t%s\n' "$name" "$rc" "$(( $(date +%s) - start ))" "$log" >> "$ledger"
}

pkgs=(./internal/journal/... ./internal/execgw/... ./internal/riskbucket/... ./internal/officialfx/...)
focus='RiskBucket|QFinal|Reconcile|MigrationV|OfficialZero|OwnerRelease|OwnerBind|A066|EntryLossLock|FirstLeg|Fill'

step untagged        go test -count=1 "${pkgs[@]}"
step seams           go test -count=1 -tags tossos_testseams ./internal/execgw/... ./internal/riskbucket/... ./internal/officialfx/...
step property        go test -count=1 -v -run 'Property|Monoton|Fuzz' ./internal/riskbucket/
step fuzz-reserve    go test -run '^$' -fuzz '^FuzzReservationIsMonotone$' -fuzztime 60s ./internal/riskbucket/
step fuzz-fill       go test -run '^$' -fuzz '^FuzzApplyFillRetryIsPure$' -fuzztime 60s ./internal/riskbucket/
step race-pure       go test -count=1 -race ./internal/riskbucket/... ./internal/officialfx/...
# journal race 는 sqlite(순수 Go) 가 race 계측 아래 느려 기본 10m 를 넘음(2026-09-28 1회차: 시험 실패·race 보고 없이
# 600 s 시한 초과) — 시한만 늘림.
step race-focus-journal go test -count=1 -race -timeout 60m -run "$focus" ./internal/journal/
step race-focus-execgw  go test -count=1 -race -run "$focus" ./internal/execgw/
step race-focus-seam go test -count=1 -race -tags tossos_testseams -run "$focus" ./internal/execgw/
step vet             go vet ./internal/journal/... ./internal/execgw/... ./internal/riskbucket/... ./internal/officialfx/...
step vet-seams       go vet -tags tossos_testseams ./internal/execgw/... ./internal/riskbucket/... ./internal/officialfx/...
cat "$ledger"

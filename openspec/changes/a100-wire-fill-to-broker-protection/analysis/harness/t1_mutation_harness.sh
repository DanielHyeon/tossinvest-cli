#!/usr/bin/env bash
# a100 T1 가드 사슬 변이 하네스 — 작업 트리를 건드리지 않고 go test -overlay 로 소스 사본만 바꿔 잼.
# 사용: t1_mutation_harness.sh <repo-root>
# 출력: 변이별 기대(CAUGHT/SHADOWED/CONTROL)와 실측, 그리고 실패 시 다음 벽의 실제 문구.
# 한계: 변이 단위는 주로 가드 **전체**(조건을 false 로)이며 절 단위 생존은 명시한 변이(…_no_poslatch,
#       AF_B3_mismatch_only 등)만 센다. -cover 와 -overlay 는 함께 쓰지 않음(go1.26.5 에서 커버리지가
#       디스크 원본을 계측함 — 결함 기록 P2-6).
set -euo pipefail
ROOT=$(cd "${1:?repo root}" && pwd)
[ "$(git -C "$ROOT" rev-parse --show-toplevel)" = "$ROOT" ] || { echo "not a repo root: $ROOT"; exit 2; }
PKG="$ROOT/internal/protectionlifecycle"
SRC="$PKG/lifecycle.go"
STATE="$PKG/state.go"
HERE=$(cd "$(dirname "$0")" && pwd)
WORK=$(mktemp -d "${TMPDIR:-/tmp}/a100-t1-mut.$$.XXXX")
trap 'rm -rf "$WORK"' EXIT
START_SHA=$(sha256sum "$SRC" | cut -d' ' -f1)
STATE_SHA=$(sha256sum "$STATE" | cut -d' ' -f1)
[ "$START_SHA" = "de50441bc89c79ec5cfeb8308a837db1cffede7e3ab52c661eccb9515d1688e5" ] || { echo "lifecycle.go sha 가 번들과 다름: $START_SHA"; exit 2; }
[ "${STATE_SHA:0:12}" = "df5c5459c6d2" ] || { echo "state.go sha 가 t1-ast 와 다름: $STATE_SHA"; exit 2; }
cd "$ROOT"

run() { # name file runpattern expect [old new]...  — file 은 패키지 안 파일명, old/new 쌍은 여러 개 가능
  local name=$1 file=$2 pat=$3 expect=$4; shift 4
  local src="$PKG/$file" mut="$WORK/$name.go" ov="$WORK/$name.json" out got verdict
  python3 - "$src" "$mut" "$@" <<'PY'
import sys
src, dst, *pairs = sys.argv[1:]
text = open(src).read()
for old, new in zip(pairs[0::2], pairs[1::2]):
    if text.count(old) != 1:
        sys.exit(f"anchor count {text.count(old)}: {old!r}")
    text = text.replace(old, new, 1)
open(dst, "w").write(text)
PY
  printf '{"Replace":{"%s":"%s"}}' "$src" "$mut" > "$ov"
  if out=$(go test -overlay "$ov" -count=1 -run "$pat" ./internal/protectionlifecycle 2>&1); then got=PASS; else got=FAIL; fi
  case "$expect:$got" in CAUGHT:FAIL|SHADOWED:PASS|CONTROL:PASS) verdict=OK;; *) verdict=UNEXPECTED;; esac
  echo "[$verdict] $name expect=$expect got=$got"
  if [ "$got" = FAIL ]; then printf '%s\n' "$out" | { grep -E '_test\.go:[0-9]+:' || true; } | head -4 | sed 's/^/    /'; fi
}

ALL='^TestA100T1'
AF_B1_OLD='if err != nil && errorCode(err) != RefusalInvalidObservation {'
AF_B2_OLD='if !validState(state) {
		return state, FillResult{PreserveExit: true}, refuse(RefusalInvalidState, "state seal invalid")'
PR_B2_OLD='if !state.marketEntryOpen(key.Market) || position.EntryLatch != "" || (position.Phase != Unprotected && position.Phase != Terminal) {'
PR_B2_RET='return state, BrokerCommand{}, refuse(RefusalEntryLatched, "entry is closed")'

run control lifecycle.go "$ALL|^Test|^Fuzz" CONTROL
run AF_B1 lifecycle.go '^TestA100T1ApplyFillB1' CAUGHT "$AF_B1_OLD" 'if false && err != nil {'
run AF_B2 lifecycle.go '.' SHADOWED "$AF_B2_OLD" "${AF_B2_OLD/if !validState/if false \&\& !validState}"
# AF_B2 양성 대조: B1·B2 를 함께 끄면 무효 봉인이 통과해 B2 시나리오 시험이 잡아야 함.
run AF_B1_B2_off lifecycle.go '^TestA100T1ApplyFillB2' CAUGHT "$AF_B1_OLD" 'if false && err != nil {' "$AF_B2_OLD" "${AF_B2_OLD/if !validState/if false \&\& !validState}"
run AF_B3_brokerid_clauses lifecycle.go '^TestA100T1ApplyFillB3' CAUGHT 'fill.BrokerOrderID == "" || fill.BrokerOrderID != position.Observed.BrokerOrderID {' 'false {'
run AF_B3_mismatch_only lifecycle.go '^TestA100T1ApplyFillB3' CAUGHT 'fill.BrokerOrderID != position.Observed.BrokerOrderID {' 'false {'
run AF_B6 lifecycle.go '^TestA100T1ApplyFillB6' CAUGHT 'if fill.Quantity == 0 || fill.Quantity > position.Observed.Quantity || fill.Quantity > position.Holdings {' 'if false {'
run AF_B6_on_B7 lifecycle.go '^TestA100T1ApplyFillB7' CAUGHT 'if fill.Quantity == 0 || fill.Quantity > position.Observed.Quantity || fill.Quantity > position.Holdings {' 'if false {'
run AF_B7 lifecycle.go '^TestA100T1ApplyFillB7' CAUGHT 'if position.Observed.Quantity == 0 {' 'if false {'
run PR_B2 lifecycle.go '^TestA100T1PrepareRegisterB2' CAUGHT "$PR_B2_OLD" 'if false {'
run PR_B2_on_B3 lifecycle.go '^TestA100T1PrepareRegisterB3' CAUGHT "$PR_B2_OLD" 'if false {'
run PR_B2_on_B4 lifecycle.go '^TestA100T1PrepareRegisterB4' CAUGHT "$PR_B2_OLD" 'if false {'
# A1 P1-2: 포지션 latch 절만 제거 — UNPROTECTED+EntryLatch 단독 사례가 잡아야 함.
run PR_B2_no_poslatch lifecycle.go '^TestA100T1PrepareRegisterB2' CAUGHT ' || position.EntryLatch != "" || (position.Phase' ' || (position.Phase'
# A1 P1-1: 거절 본문이 호출자 상태(공유 맵)를 제자리 변경 — 호출 전 스냅숏 비교가 잡아야 함.
run ALIAS_PR_B2 lifecycle.go '^TestA100T1PrepareRegisterB2' CAUGHT "$PR_B2_RET" "p := state.positions[key]; p.Holdings++; state.positions[key] = p
		$PR_B2_RET"
run PR_B3 lifecycle.go '.' SHADOWED 'if position.HasPending {' 'if false && position.HasPending {'
run PR_B4 lifecycle.go '.' SHADOWED 'if position.Observed.Status == BrokerActive {
		return state, BrokerCommand{}, refuse(RefusalOperationPending, "protection already active")' 'if false {
		return state, BrokerCommand{}, refuse(RefusalOperationPending, "protection already active")'
run PR_B5 lifecycle.go '^TestA100T1PrepareRegisterB5' CAUGHT 'if !capability.exactOperationLookup {
		return state, BrokerCommand{}, refuse(RefusalInvalidObservation, "exact operation lookup unavailable")' 'if false {
		return state, BrokerCommand{}, refuse(RefusalInvalidObservation, "exact operation lookup unavailable")'
# A1 P2-2(M8): US 포지션만 진리표를 건너뜀 — 두 시장 구조 시험이 잡아야 함.
run TRUTH_SKIP_US state.go '^TestA100T1UnreachablePrepareRegisterB3B4' CAUGHT 'if !validPositionTruth(state, key, position) {' 'if key.Market != MarketUS && !validPositionTruth(state, key, position) {'

# RED 증인: B3·B4 고유 문구를 단언하는 시험(저장소 밖 .go.txt)을 overlay 로 넣어 현 코드에서 RED 임을 보임.
RED="$HERE/a100_t1_red_unreachable_test.go.txt"
if [ -f "$RED" ]; then
  REDPATH="$ROOT/internal/protectionlifecycle/a100_t1_red_unreachable_test.go"
  cp "$SRC" "$WORK/plain.go"
  printf '{"Replace":{"%s":"%s","%s":"%s"}}' "$SRC" "$WORK/plain.go" "$REDPATH" "$RED" > "$WORK/red.json"
  if out=$(go test -overlay "$WORK/red.json" -count=1 -run '^TestA100T1RED' -v ./internal/protectionlifecycle 2>&1); then got=PASS; else got=FAIL; fi
  echo "[$( [ $got = FAIL ] && echo OK || echo UNEXPECTED )] RED_witness_current_code expect=FAIL got=$got"
  printf '%s\n' "$out" | { grep -E '^(--- |    --- )|want=' || true; } | sed 's/^/    /'
  # 양성 대조: B2 의 포지션 절(EntryLatch·phase)을 빼면(가상의 core 수정) RED 시험이 GREEN 이 되는지 — 시험이 B3·B4 본문에 닿을 수 있음을 보임.
  python3 - "$SRC" "$WORK/nophase.go" <<'PY'
import sys
src, dst = sys.argv[1:]
t = open(src).read()
old = ' || position.EntryLatch != "" || (position.Phase != Unprotected && position.Phase != Terminal) {'
assert t.count(old) == 1
open(dst, "w").write(t.replace(old, ' {', 1))
PY
  printf '{"Replace":{"%s":"%s","%s":"%s"}}' "$SRC" "$WORK/nophase.go" "$REDPATH" "$RED" > "$WORK/red2.json"
  if out=$(go test -overlay "$WORK/red2.json" -count=1 -run '^TestA100T1RED' -v ./internal/protectionlifecycle 2>&1); then got=PASS; else got=FAIL; fi
  echo "[$( [ $got = PASS ] && echo OK || echo UNEXPECTED )] RED_witness_positive_control(B2 포지션 절 제거) expect=PASS got=$got"
  printf '%s\n' "$out" | { grep -E '^(--- |    --- )|want=' || true; } | sed 's/^/    /'
fi
# 부수 관찰(D3): REPLACE_PENDING·REPLACE_UNKNOWN 중 부분 체결이 진리표를 깨는 상태를 재봉인해 돌려주고 영구 정지시키는지 — 로그만 출력.
PROBE="$HERE/t1_probe_fill_during_replace_pending_test.go.txt"
if [ -f "$PROBE" ]; then
  printf '{"Replace":{"%s":"%s"}}' "$ROOT/internal/protectionlifecycle/zz_t1_probe_test.go" "$PROBE" > "$WORK/probe.json"
  go test -overlay "$WORK/probe.json" -count=1 -run '^TestProbeFillDuringPending$' -v ./internal/protectionlifecycle 2>&1 | { grep -E 'zz_t1_probe_test.go:[0-9]+:' || true; } | sed 's/^/    /'
fi
END_SHA=$(sha256sum "$SRC" | cut -d' ' -f1)
[ "$END_SHA" = "$START_SHA" ] || { echo "lifecycle.go 가 하네스 중 바뀜"; exit 3; }
[ "$(sha256sum "$STATE" | cut -d' ' -f1)" = "$STATE_SHA" ] || { echo "state.go 가 하네스 중 바뀜"; exit 3; }
echo "lifecycle.go sha 불변: $END_SHA"

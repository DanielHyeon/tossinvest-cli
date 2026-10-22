#!/usr/bin/env bash
# a087 tasks 3.8 — exit_slippage.sql 변이 하네스. 질의는 시험이 실행 시점에 디스크에서 읽으므로 go -overlay 가 닿지 않음 —
# 그래서 사본을 떠 두고 제자리에서 바꾼 뒤 매번 사본으로 되돌리고, 끝에 sha256 로 원복을 단언함.
# 사용(저장소 루트에서): bash openspec/changes/a087-a-protective-exit-is-a-market-order/analysis/exit-slippage-3.8/query_mutants.sh
set -euo pipefail
ROOT=$(git rev-parse --show-toplevel)
cd "$ROOT"
Q=openspec/changes/a087-a-protective-exit-is-a-market-order/analysis/exit-slippage-3.8/exit_slippage.sql
TMP=$(mktemp -d)
trap 'cp "$TMP/orig.sql" "$Q"; rm -rf "$TMP"' EXIT
cp "$Q" "$TMP/orig.sql"
before=$(sha256sum < "$Q")
run() { # 이름 원문 변이문
	cp "$TMP/orig.sql" "$Q"
	python3 -I - "$Q" "$2" "$3" <<'PY'
import sys
p, o, n = sys.argv[1:]
s = open(p).read()
assert s.count(o) == 1, o
open(p, "w").write(s.replace(o, n))
PY
	if go test ./internal/journal/ -run '^TestA087ExitSlippageQueryBindsTheDecisionPriceToItsOwnFill$' -count=1 >"$TMP/out" 2>&1; then
		echo "$1 SURVIVED"
	else
		echo "$1 CAUGHT — $(grep -m1 '_test.go:' "$TMP/out" | sed 's/^ *//')"
	fi
}
cp "$TMP/orig.sql" "$Q"
if go test ./internal/journal/ -run '^TestA087ExitSlippageQueryBindsTheDecisionPriceToItsOwnFill$' -count=1 >/dev/null 2>&1; then
	echo "control PASS"
else
	echo "control FAIL — 변이 결과를 믿을 수 없음"; exit 3
fi
run "Q1 day-scope join removed" $'   AND s.trading_day = TRIM(i.trading_day)\n   AND s.symbol' '   AND s.symbol'
run "Q2 NULLIF guard removed" "CAST(NULLIF(TRIM(s.average_price), '') AS REAL)" "CAST(coalesce(s.average_price, '') AS REAL)"
run "Q3 priceless protective rows kept" "AND TRIM(coalesce(e.observed_price, '')) <> ''" "AND 1 = 1"
cp "$TMP/orig.sql" "$Q"
[ "$(sha256sum < "$Q")" = "$before" ] && echo "restored (sha256 equal)"

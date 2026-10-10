#!/usr/bin/env bash
# mutants.sh — 가드·단일 전송 변이 점검(제자리 치환 → 시험 → 복원 sha 대조).
#
# 무변이 대조군이 GREEN 이어야 시작함. 변이마다 (1) 파일 sha 가 바뀌었는지(닿았는지),
# (2) 빌드 실패가 아닌 시험 실패인지, (3) 복원 후 sha 가 원본과 같은지를 확인함.
# 실 API 호출 없음 — 시험은 httptest mock 만 침.
#
#   bash tools/a087-market-sell-probe/mutants.sh
set -euo pipefail

ROOT="$(git -C "$(dirname "$0")" rev-parse --show-toplevel)"
DIR="$ROOT/tools/a087-market-sell-probe"
PKG="./tools/a087-market-sell-probe/"
cd "$ROOT"
[ "$(basename "$DIR")" = "a087-market-sell-probe" ] || { echo "wrong dir"; exit 1; }

runtests() { go test -count=1 "$PKG" >"$1" 2>&1; }

LOG="$(mktemp)"
CUR_TARGET=""
CUR_BACKUP=""
# 중단돼도 변이를 남기지 않게 현재 대상 파일을 되돌림.
cleanup() {
  if [ -n "$CUR_BACKUP" ] && [ -f "$CUR_BACKUP" ]; then cp "$CUR_BACKUP" "$CUR_TARGET"; rm -f "$CUR_BACKUP"; fi
  rm -f "$LOG"
}
trap cleanup EXIT
SOURCES_BEFORE="$(cd "$DIR" && sha256sum ./*.go)"
if ! runtests "$LOG"; then echo "NULL CONTROL RED — refusing to run mutants"; cat "$LOG"; exit 1; fi
echo "null control: GREEN"

# id|file|perl substitution
MUTANTS=(
  'M1-sell-guard-deleted|order.go|s/if spec\.Side != fixedSide \{/if false \&\& spec.Side != fixedSide {/'
  'M2-qty-cap-relaxed|order.go|s/spec\.Quantity > maxQuantity \{/spec.Quantity > maxQuantity+1 {/'
  'M2b-qty-const-raised|order.go|s/maxQuantity    = 2/maxQuantity    = 3/'
  'M3-retry-on-5xx|send.go|s/(\tresp, err := client\.Do\(req\)\n)/$1\tif err == nil \&\& resp.StatusCode >= 500 {\n\t\tresp.Body.Close()\n\t\tretry := req.Clone(req.Context())\n\t\tretry.Body, _ = req.GetBody()\n\t\tresp, err = client.Do(retry)\n\t}\n/'
  'M4-confirm-check-skipped|main.go|s/if err := verifyConfirm\(order, o\.confirm, d\.now\(\)\); err != nil \{/if err := verifyConfirm(order, o.confirm, d.now()); false \&\& err != nil {/'
  'M4b-token-compare-bypassed|order.go|s/\[\]byte\(token\)\) != 1 \{/[]byte(token)) != 1 \&\& false {/'
  'M5-market-guard-deleted|order.go|s/if spec\.Market != fixedMarket \{/if false \&\& spec.Market != fixedMarket {/'
  'M6-limit-guard-deleted|order.go|s/if spec\.OrderType != fixedOrderType \{/if false \&\& spec.OrderType != fixedOrderType {/'
  'M7-redirect-followed|send.go|s/\t\tCheckRedirect: func\(\*http\.Request, \[\]\*http\.Request\) error \{ return http\.ErrUseLastResponse \},\n//'
  'M8-exclusive-create-dropped|receipt.go|s/os\.O_WRONLY\|os\.O_CREATE\|os\.O_EXCL/os.O_WRONLY|os.O_CREATE|os.O_TRUNC/'
  'M9-expiry-skipped|order.go|s/if age > confirmWindow \{/if false \&\& age > confirmWindow {/'
  'M10-terminal-check-skipped|main.go|s/if !d\.stdinIsTerminal\(\) \{/if false \&\& !d.stdinIsTerminal() {/'
  'M11-symbol-guard-widened|order.go|s/\^\[0-9\]\{6\}\$/^[0-9A-Z]{4,7}\$/'
)

status=0
for entry in "${MUTANTS[@]}"; do
  IFS='|' read -r id file expr <<<"$entry"
  target="$DIR/$file"
  backup="$(mktemp)"
  cp "$target" "$backup"
  CUR_TARGET="$target"
  CUR_BACKUP="$backup"
  before="$(sha256sum "$target" | cut -d' ' -f1)"
  perl -0pi -e "$expr" "$target"
  after="$(sha256sum "$target" | cut -d' ' -f1)"
  if [ "$before" = "$after" ]; then
    verdict="NOT-REACHED"
  elif runtests "$LOG"; then
    verdict="SURVIVED"
  elif grep -q -e '\[build failed\]' -e '\[setup failed\]' "$LOG"; then
    verdict="INVALID(build)"
  else
    verdict="CAUGHT by $(grep -o -E -- '--- FAIL: [A-Za-z0-9_/]+' "$LOG" | sed 's/--- FAIL: //' | sort -u | tr '\n' ' ')"
  fi
  cp "$backup" "$target"
  rm -f "$backup"
  CUR_BACKUP=""
  restored="$(sha256sum "$target" | cut -d' ' -f1)"
  [ "$restored" = "$before" ] || { echo "RESTORE FAILED for $file"; exit 2; }
  echo "$id: $verdict (restored sha ${restored:0:12})"
  case "$verdict" in CAUGHT*) ;; *) status=1 ;; esac
done

[ "$(cd "$DIR" && sha256sum ./*.go)" = "$SOURCES_BEFORE" ] || { echo "SOURCE SHA DRIFT after the run"; exit 2; }
if ! runtests "$LOG"; then echo "POST-RUN RED — sources not restored?"; exit 2; fi
echo "post-run: GREEN"
exit "$status"

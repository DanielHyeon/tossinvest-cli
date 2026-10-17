#!/usr/bin/env bash
# mutants.sh — 가드·단일 전송 변이 점검. 변이는 **저장소 밖 사본**에서만 일어남.
#
# 왜 사본인가(a087 review.md 「3차 proposal-freeze 재리뷰」 B-P1-1): 이 도구는 실돈 주문을 냄.
# 공유 작업 트리의 소스를 제자리로 바꾸면 변이가 살아 있는 동안(또는 SIGKILL·OOM 으로 복원이
# 안 돈 뒤) 같은 트리에서 `go run ./tools/a087-market-sell-probe --execute …` 가 변이체를 컴파일해
# 실주문을 낼 수 있음(M3 재전송·M7 리다이렉트 추종·M4/M9 확인 우회가 정확히 그 변이임).
# 그래서 공유 트리는 읽기만 하고, 커밋 하나를 `git worktree add --detach` 로 저장소 밖에 꺼내
# 그 사본만 고침. 중단돼도 남는 것은 저장소 밖 사본뿐임.
#
# 순서: 대상 커밋 해석 → 사본 생성(경로에 pid) → 사본 HEAD == 대상 커밋 단언 → 무변이 대조군 GREEN
#       → 변이마다 (닿음 = 파일 sha 변화, 빌드 실패가 아닌 시험 실패, 복원 sha 대조) → 사후 GREEN
#       → 공유 트리 sha 불변 단언 → 사본·전용 GOCACHE 제거.
# go 캐시는 전용 GOCACHE + GOFLAGS=-trimpath (임시 경로마다 공유 캐시가 불어나는 것 방지).
# 실 API 호출 없음 — 시험은 httptest mock 만 침.
#
#   bash tools/a087-market-sell-probe/mutants.sh            # HEAD (도구 디렉터리가 깨끗해야 함)
#   bash tools/a087-market-sell-probe/mutants.sh <commit>   # 특정 커밋
set -euo pipefail

ROOT="$(git -C "$(dirname "$0")" rev-parse --show-toplevel)"
REL="tools/a087-market-sell-probe"
PKG="./$REL/"
[ -d "$ROOT/$REL" ] || { echo "wrong repo: $ROOT/$REL missing"; exit 1; }

# 대상 커밋 해석. 인자가 없으면 HEAD 이고, 그때 도구 디렉터리가 더러우면 거절함 —
# 사본은 커밋 내용이라 미커밋 편집은 재지 않으므로, 보고가 보이는 소스와 어긋나게 됨.
if [ "$#" -gt 1 ]; then echo "usage: $0 [commit]"; exit 1; fi
if [ "$#" -eq 0 ]; then
  if [ -n "$(git -C "$ROOT" status --porcelain -- "$REL")" ]; then
    echo "REFUSED: $REL has uncommitted changes; commit them or pass a commit explicitly"
    git -C "$ROOT" status --porcelain -- "$REL"
    exit 1
  fi
  REV="HEAD"
else
  REV="$1"
fi
SHA="$(git -C "$ROOT" rev-parse --verify --quiet "$REV^{commit}")" || { echo "not a commit: $REV"; exit 1; }

# 공유 트리의 도구 소스 지문 — 끝에서 그대로인지 봄(이 하네스는 공유 트리를 쓰지 않음).
shared_sums() { (cd "$ROOT/$REL" && sha256sum ./*.go ./*.sh); }
SHARED_BEFORE="$(shared_sums)"

BASE="$(mktemp -d "${TMPDIR:-/tmp}/a087-mutants-$$-XXXXXX")"
BASE="$(cd "$BASE" && pwd -P)"
ROOT_REAL="$(cd "$ROOT" && pwd -P)"
case "$BASE/" in "$ROOT_REAL"/*) echo "REFUSED: scratch dir $BASE is inside the repository"; rm -rf "$BASE"; exit 1 ;; esac
COPY="$BASE/tree"
LOG="$BASE/test.log"

cleanup() {
  # 사본은 저장소 밖이라 남아도 공유 트리의 빌드에 섞이지 않음 — 그래도 지움.
  cd /
  if [ -d "$COPY" ]; then git -C "$ROOT" worktree remove --force "$COPY" >/dev/null 2>&1 || true; fi
  git -C "$ROOT" worktree prune >/dev/null 2>&1 || true
  chmod -R u+w "$BASE" 2>/dev/null || true
  rm -rf "$BASE"
}
trap cleanup EXIT
trap 'exit 130' INT TERM

# 훅은 끔(core.hooksPath=/dev/null) — 사본 생성이 저장소 훅을 돌리지 않게 함.
git -C "$ROOT" -c core.hooksPath=/dev/null worktree add --detach --quiet "$COPY" "$SHA"

# 시작 sha 단언 — 사본 HEAD 가 의도한 커밋이고 사본이 깨끗해야 측정이 그 커밋에 대한 진술이 됨.
COPY_HEAD="$(git -C "$COPY" rev-parse HEAD)"
[ "$COPY_HEAD" = "$SHA" ] || { echo "COPY HEAD $COPY_HEAD != intended $SHA"; exit 2; }
[ -z "$(git -C "$COPY" status --porcelain)" ] || { echo "COPY not clean"; exit 2; }

export GOCACHE="$BASE/gocache"
export GOFLAGS="-trimpath"
cd "$COPY"
# go 가 사본의 모듈을 보는지 단언 — 공유 트리의 go.mod 로 빌드되면 변이가 닿지 않음.
[ "$(go env GOMOD)" = "$COPY/go.mod" ] || { echo "go resolves $(go env GOMOD), not the copy"; exit 2; }

DIR="$COPY/$REL"
runtests() { go test -count=1 "$PKG" >"$1" 2>&1; }

echo "measuring commit $SHA in $COPY"
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
  # 변이 대상은 반드시 사본 안이어야 함 — 공유 트리 파일이면 즉시 중단.
  case "$target" in "$COPY"/*) ;; *) echo "REFUSED: mutant target $target outside the copy"; exit 2 ;; esac
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
  git -C "$COPY" checkout --quiet -- "$REL/$file"
  restored="$(sha256sum "$target" | cut -d' ' -f1)"
  [ "$restored" = "$before" ] || { echo "RESTORE FAILED for $file"; exit 2; }
  echo "$id: $verdict (reached: $file ${before:0:12} -> ${after:0:12}, restored ${restored:0:12})"
  case "$verdict" in CAUGHT*) ;; *) status=1 ;; esac
done

[ -z "$(git -C "$COPY" status --porcelain)" ] || { echo "COPY not clean after the run"; exit 2; }
if ! runtests "$LOG"; then echo "POST-RUN RED — copy not restored?"; exit 2; fi
echo "post-run: GREEN"
[ "$(shared_sums)" = "$SHARED_BEFORE" ] || { echo "SHARED TREE CHANGED during the run"; exit 2; }
echo "shared tree: untouched"
exit "$status"

#!/usr/bin/env bash
# a115 뮤테이션 하네스 — 저장소가 아니라 **커밋된 트리의 사본**(git archive <rev>)에서만 변이한다.
#   사용: run.sh <name> <mutation.py|none> [rev=HEAD]
# - 사본 이름에 pid 를 단다(두 판이 한 사본을 쓰면 서로의 변이가 기준이 된다). 한 번에 한 판만 돌린다.
# - 사본은 워킹트리가 아니라 커밋에서 푼다 — 병행 세션의 미커밋 편집이 판정에 섞이지 않는다.
# - 판정은 rc 가 아니라 출력이다(a114 post-review P2-3): FAIL 줄 → CAUGHT, 빌드 실패 → BROKEN,
#   네 패키지 모두 ok → SURVIVED, 그 밖 → UNKNOWN.
# - 끝에서 실제 저장소의 대상 파일이 그 커밋과 같은지 단언한다(변이가 저장소로 새지 않았다).
set -euo pipefail
H=$(cd "$(dirname "$0")" && pwd)
REPO=$(git -C "$H" rev-parse --show-toplevel)
name=$1; mut=$2; rev=${3:-HEAD}
sha=$(git -C "$REPO" rev-parse "$rev")
W=${TMPDIR:-/tmp}/a115-mut-$name-$$
L=${TMPDIR:-/tmp}/a115-mut-$name.log
rm -rf "$W"; mkdir -p "$W"
git -C "$REPO" archive "$sha" go.mod go.sum cmd internal docs/api | tar -x -C "$W"
if [ "$mut" != none ]; then
  python3 "$H/$mut" "$W"
fi
cd "$W"
set +e
GOFLAGS=-trimpath go test -count=1 -run "$(tr -d '\n' < "$H/tests.txt")" \
  ./cmd/tossctl/ ./internal/console/ ./internal/httpapi/ ./internal/strategyprojection/ > "$L" 2>&1
rc=$?
set -e
cd /
rm -rf "$W"
if grep -qE '\[build failed\]|\[setup failed\]' "$L"; then verdict=BROKEN
elif grep -qE '^(--- FAIL|    --- FAIL|panic:)' "$L"; then verdict=CAUGHT
elif [ "$(grep -cE '^ok[[:space:]]' "$L")" = 4 ]; then verdict=SURVIVED
else verdict=UNKNOWN; fi
# 저장소로 샌 변이가 없는지 — 대상 파일이 시작 커밋과 같아야 한다.
leak=$(git -C "$REPO" diff --name-only "$sha" -- cmd/tossctl/console.go cmd/tossctl/console_strategy_attach.go \
  cmd/tossctl/httpapi_strategy_attach.go internal/console/strategy_runtime_multimarket.go \
  internal/console/settings_tabs.go internal/httpapi/strategy_runtime.go internal/strategyprojection/presence.go)
echo "$name sha=${sha:0:8} rc=$rc verdict=$verdict leak=[${leak}]"
grep -E '^(--- FAIL|    --- FAIL|panic:)|build failed' "$L" | head -6 || true

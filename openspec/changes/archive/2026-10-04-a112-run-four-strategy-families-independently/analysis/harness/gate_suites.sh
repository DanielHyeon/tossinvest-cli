#!/usr/bin/env bash
# a112 8.1 · 8.3 스위트 러너 — gate-8.1-8.3-2026-10-04/suites-summary.log 의 명령 집합을 그대로 재현하고,
# 7.3.1 로트가 더한 패키지(strategyshadow · strategyprojection · httpapi · tools/a112-family-shadow)를 8.1 집합에 더함.
#
# 쓰기: gate_suites.sh <고정 커밋의 연결 워크트리> <출력 디렉터리>
# 측정은 연결 워크트리에서 함 — 병행 커밋이 측정 중 HEAD 를 옮기지 않게(기억: 긴 측정은 고정 워크트리에서).
# 각 명령의 전체 출력은 <출력>/<이름>.log, 요약은 <출력>/suites-summary.log(exit · 초 · ok 패키지 수 · FAIL 줄 수).
set -uo pipefail
tree=$1
out=$2
mkdir -p "$out"
summary="$out/suites-summary.log"
cd "$tree" || exit 2
echo "== worktree $(git rev-parse HEAD) $(date -Iseconds)" >"$summary"
pkgs81="./internal/breakoutlane/... ./internal/strategyflow/... ./internal/strategyrouter/... ./internal/strategyproposal/... ./internal/scheduler/... ./internal/strategyworker/... ./internal/strategycoordinator/... ./internal/strategyarbiter/... ./internal/strategyhandoff/... ./internal/continuationlane/... ./internal/reversallane/... ./internal/weeklyvaluelane/... ./internal/strategyshadow/... ./internal/strategyprojection/... ./internal/httpapi/... ./tools/a112-family-shadow/..."
run() {
	local name=$1
	shift
	local start=$SECONDS
	"$@" >"$out/$name.log" 2>&1
	local code=$?
	echo "$name exit=$code ($((SECONDS - start))s): $*" >>"$summary"
}
# shellcheck disable=SC2086
run 81-untagged go test -count=1 -timeout 30m $pkgs81 ./internal/app/engine/...
# shellcheck disable=SC2086
run 81-tagged go test -count=1 -timeout 30m -tags tossos_testseams $pkgs81 ./internal/app/engine/...
# shellcheck disable=SC2086
run 81-race-pkgs go test -race -count=1 -timeout 30m -tags tossos_testseams $pkgs81
run 83-make-test-race make test-race
run 83-make-test make test
run 83-make-test-seams make test-seams
run 83-make-vet make vet
run 83-make-lint make lint
run 83-make-validate make validate
run 83-openspec-validate openspec validate a112-run-four-strategy-families-independently --strict --no-interactive
run 83-pm-check python3 tools/pm/generate_master_tracker.py --check
echo "== done $(date -Iseconds)" >>"$summary"
for name in 83-make-test 83-make-test-seams 83-make-test-race 81-untagged 81-tagged 81-race-pkgs; do
	echo "$name ok-packages=$(grep -c '^ok ' "$out/$name.log") fail-lines=$(grep -c -E '^(FAIL|--- FAIL)' "$out/$name.log")" >>"$summary"
done

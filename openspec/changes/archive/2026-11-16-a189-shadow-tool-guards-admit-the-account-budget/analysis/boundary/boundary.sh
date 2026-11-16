#!/usr/bin/env bash
# a189 — 세 가드의 PASS→FAIL 경계를 원인 커밋과 그 부모에서 **실제 시험으로** 잼(읽기 전용).
# 커밋마다 go.mod·go.sum·cmd·internal 만 git archive 로 꺼내 $HOME/tossos-work 아래 사본에서 시험 이름을 지정해 돌림.
# 쓰는 법: boundary.sh <저장소 절대경로>
set -euo pipefail
REPO=${1:?repo}
ROOT=$(mktemp -d "$HOME/tossos-work/a189-boundary.$$.XXXX")
trap 'rm -rf "$ROOT"' EXIT
export GOFLAGS=-trimpath GOCACHE="$HOME/tossos-work/.a189-gocache" HOME_ORIG="$HOME"
# 시험이 실제 홈의 설정을 읽지 않게 격리함
export XDG_CONFIG_HOME="$ROOT/xdg" XDG_CACHE_HOME="$ROOT/xdgcache"
run() { # run <commit> <pkg> <test regex>
	local c=$1 pkg=$2 re=$3 dir
	dir="$ROOT/$(git -C "$REPO" rev-parse --short "$c")"
	if [ ! -d "$dir" ]; then
		mkdir -p "$dir"
		git -C "$REPO" archive "$c" -- go.mod go.sum cmd internal | tar -x -C "$dir"
	fi
	echo "### $c ($(git -C "$REPO" log -1 --format='%h %s' "$c")) — $pkg -run '$re'"
	set +e
	(cd "$dir" && go test "$pkg" -run "$re" -count=1 2>&1 | grep -E '^(--- |ok|FAIL|PASS|panic|    [a-z_0-9]+\.go:[0-9]+)' | head -20)
	set -e
}
run 3dc0c776b^ ./cmd/tossctl/ '^TestTheOrdersSeamResolvesTheAccountOnceAndBuildsNoSecondClient$'
run 3dc0c776b  ./cmd/tossctl/ '^TestTheOrdersSeamResolvesTheAccountOnceAndBuildsNoSecondClient$'
run 9c4bdf4de^ ./internal/app/engine/ '^TestA163TheJudgementFunctionsRenderExactlyAsBeforeTheChange$'
run 9c4bdf4de  ./internal/app/engine/ '^TestA163TheJudgementFunctionsRenderExactlyAsBeforeTheChange$'
run 5c0aeb2b8^ ./internal/app/engine/ '^TestExactlyOneProductionSiteAdmitsIntoTheSeam$'
run 5c0aeb2b8  ./internal/app/engine/ '^TestExactlyOneProductionSiteAdmitsIntoTheSeam$'

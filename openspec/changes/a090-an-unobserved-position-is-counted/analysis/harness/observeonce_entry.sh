#!/usr/bin/env bash
# ObserveOnce 분기 진입 실측 — 깨끗한 detached worktree 에서 engine 패키지 전체를 covermode=set 으로 돌리고
# ObserveOnce 줄 범위(ast.json 의 start..end)의 블록만 뽑음. 공유 트리의 이웃 미커밋 Go 가 섞이지 않게 함.
# 사용: observeonce_entry.sh <commit> <출력디렉터리>
set -euo pipefail
commit=${1:?commit}; out=${2:?outdir}
repo=$(git -C "$(dirname "$0")" rev-parse --show-toplevel)
wt=$(mktemp -d "${TMPDIR:-/tmp}/a090-entry-XXXXXX")
trap 'git -C "$repo" worktree remove --force "$wt" >/dev/null 2>&1 || true' EXIT
git -C "$repo" worktree add --detach "$wt" "$commit" >/dev/null
test "$(git -C "$wt" rev-parse HEAD)" = "$(git -C "$repo" rev-parse "$commit^{commit}")"
mkdir -p "$out"
(cd "$wt" && go test ./internal/app/engine/ -count=1 -covermode=set -coverprofile="$out/engine.cov" >"$out/test.log" 2>&1)
python3 - "$repo/openspec/changes/a090-an-unobserved-position-is-counted/analysis/function-logic/internal-app-engine--exitobserver.observeonce/ast.json" "$out/engine.cov" >"$out/observeonce.blocks" <<'PY'
import json, re, sys
ast = json.load(open(sys.argv[1]))
lo, hi = ast["start"]["line"], ast["end"]["line"]
for line in open(sys.argv[2]):
    m = re.match(r".*/internal/app/engine/exitloop\.go:(\d+)\.\d+,(\d+)\.\d+ \d+ (\d+)$", line.strip())
    if m and lo <= int(m.group(1)) <= hi:
        print(line.strip().split("/engine/", 1)[1])
PY
cat "$out/observeonce.blocks"

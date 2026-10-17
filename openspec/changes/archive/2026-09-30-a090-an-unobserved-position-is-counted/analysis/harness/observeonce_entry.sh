#!/usr/bin/env bash
# a090 번들 분기 진입 실측 — 깨끗한 detached worktree 에서 engine 패키지 전체를 covermode=set 으로 돌리고,
# 이 change 의 모든 function-logic 번들(ast.json 의 start..end 줄 범위)에 든 블록만 뽑음.
# 공유 트리의 이웃 미커밋 Go 가 섞이지 않게 함. ast.json 은 **측정 커밋에서** 읽음(워킹트리 아님).
# 사용: observeonce_entry.sh <commit> <출력디렉터리>
# 산출: <출력디렉터리>/engine.cov · test.log · observeonce.blocks(머리줄 = 커밋·명령·go test 요약, 하네스가 씀)
set -euo pipefail
commit=${1:?commit}; out=${2:?outdir}
repo=$(git -C "$(dirname "$0")" rev-parse --show-toplevel)
full=$(git -C "$repo" rev-parse "$commit^{commit}")
change=openspec/changes/a090-an-unobserved-position-is-counted
wt=$(mktemp -d "${TMPDIR:-/tmp}/a090-entry-XXXXXX")
trap 'git -C "$repo" worktree remove --force "$wt" >/dev/null 2>&1 || true' EXIT
git -C "$repo" worktree add --detach "$wt" "$full" >/dev/null
test "$(git -C "$wt" rev-parse HEAD)" = "$full"
mkdir -p "$out"
cmd="go test ./internal/app/engine/ -count=1 -covermode=set"
(cd "$wt" && $cmd -coverprofile="$out/engine.cov" >"$out/test.log" 2>&1)
{
  printf '# commit %s · %s · %s\n' "$full" "$cmd" "$(tail -1 "$out/test.log" | tr '\t' ' ')"
  for bundle in $(git -C "$wt" ls-tree -d --name-only "$full" "$change/analysis/function-logic/"); do
    python3 - "$wt/$bundle/ast.json" "$out/engine.cov" "$(basename "$bundle")" <<'PY'
import json, re, sys
ast = json.load(open(sys.argv[1]))
lo, hi = ast["start"]["line"], ast["end"]["line"]
src = ast["file"].split("/")[-1]
print(f"## {sys.argv[3]} ({src}:{lo}-{hi})")
for line in open(sys.argv[2]):
    m = re.match(r".*/internal/app/engine/(\S+):(\d+)\.\d+,(\d+)\.\d+ \d+ (\d+)$", line.strip())
    if m and m.group(1) == src and lo <= int(m.group(2)) <= hi:
        print(line.strip().split("/engine/", 1)[1])
PY
  done
} >"$out/observeonce.blocks"
cat "$out/observeonce.blocks"

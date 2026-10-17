#!/usr/bin/env bash
# a066 — 공유 워크트리 대신 사본에서 go test 를 돌림.
#
# 왜: 병행 로트가 미추적 RED 시험(컴파일 안 되는 파일)을 같은 패키지에 두면 그 패키지 시험 전체가 빌드 실패함
# (2026-09-27 a124 의 internal/journal RED 파일). 사본에는 git 추적 파일(작업 트리 내용) + 이 로트의 미추적 파일만
# 넣음 — 남의 미추적 파일은 들어오지 않음.
#
# 사용: test_in_copy.sh <copy-dir> <own-untracked-file>... -- <go test args...>
set -euo pipefail
copy=${1:?copy dir}
shift
own=()
while [ "$#" -gt 0 ] && [ "$1" != "--" ]; do own+=("$1"); shift; done
[ "${1:-}" = "--" ] && shift
root=$(git rev-parse --show-toplevel)
mkdir -p "$copy"
list=$(mktemp)
trap 'rm -f "$list"' EXIT
git -C "$root" ls-files -- go.mod go.sum internal cmd tools > "$list"
for f in "${own[@]}"; do echo "$f" >> "$list"; done
rsync -a --delete --files-from="$list" "$root/" "$copy/"
# --delete 는 --files-from 과 함께 목록 밖 파일을 지우지 않으므로 사본의 미추적 잔재를 따로 지움.
( cd "$copy" && comm -23 <(find go.mod go.sum internal cmd tools -type f 2>/dev/null | sort) <(sort "$list") | xargs -r rm -f )
echo "copy HEAD $(git -C "$root" rev-parse --short HEAD) files $(wc -l < "$list")" >&2
cd "$copy"
exec go test "$@"

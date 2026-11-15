#!/usr/bin/env bash
# a189 #2 — ExitObserver.record 렌더 해시를 커밋마다 잼(읽기 전용: git show 로 blob 을 꺼내 임시 디렉터리에서 렌더).
# 범위: 9c4bdf4de 의 부모부터 base 7d1e3163a 까지 **모든** 커밋(경로 필터·병합 단순화 없음 — rev-list 전수).
# 커밋마다 exitloop.go 의 blob id 를 읽고, 서로 다른 blob 만 렌더해서 커밋→해시 표를 만듦.
# 쓰는 법: render_at_commits.sh <저장소 절대경로>
set -euo pipefail
REPO=${1:?repo}
HERE=$(cd "$(dirname "$0")" && pwd)
WORK=$(mktemp -d "$HOME/tossos-work/a189-render.$$.XXXX")
trap 'rm -rf "$WORK"' EXIT
mkdir -p "$WORK/r"
cp "$HERE/render.go.txt" "$WORK/r/main.go"
export GOFLAGS=-trimpath
(cd "$WORK/r" && go mod init a189render >/dev/null 2>&1 && go build -o "$WORK/render" .)
FILE=internal/app/engine/exitloop.go
CAUSE=9c4bdf4de
BASE=7d1e3163a

echo "# 대조: 워킹트리 $FILE (시험 실패 문구의 현재 값과 같아야 함)"
"$WORK/render" "$REPO/$FILE"

echo "# 커밋 수: $(git -C "$REPO" rev-list --count "$CAUSE^..$BASE") (rev-list $CAUSE^..$BASE, + 경계 $CAUSE^ 자신)"
declare -A seen
{ git -C "$REPO" rev-parse "$CAUSE^"; git -C "$REPO" rev-list "$CAUSE^..$BASE"; } | while read -r c; do
	blob=$(git -C "$REPO" rev-parse --verify --quiet "$c:$FILE" || echo MISSING)
	echo "$c $blob"
done > "$WORK/commit-blob.txt"
cut -d' ' -f2 "$WORK/commit-blob.txt" | sort -u > "$WORK/blobs.txt"
echo "# 서로 다른 blob 수: $(wc -l < "$WORK/blobs.txt")"
while read -r blob; do
	if [ "$blob" = MISSING ]; then echo "MISSING-FILE"; continue; fi
	git -C "$REPO" cat-file blob "$blob" > "$WORK/f.go"
	rec=$("$WORK/render" "$WORK/f.go" | awk '$1=="record"{print $2}')
	n=$(git -C "$REPO" cat-file blob "$blob" | grep -c '^func (o \*ExitObserver) record(' || true)
	users=$(awk -v b="$blob" '$2==b' "$WORK/commit-blob.txt" | wc -l)
	echo "blob $blob record=${rec:-NOT-FOUND} decl_lines=$n commits=$users"
done < "$WORK/blobs.txt"

echo "# 경계 커밋 — 엔진 생산 파일 전수(시험과 같은 범위: internal/app/engine/*.go 중 _test.go 아님)를 렌더해 고정 열 함수를 모음"
for c in "$CAUSE^" "$CAUSE" "$BASE"; do
	rm -rf "$WORK/eng"; mkdir -p "$WORK/eng"
	git -C "$REPO" ls-tree --name-only "$c" internal/app/engine/ | grep -E '\.go$' | grep -v '_test\.go$' | while read -r f; do
		git -C "$REPO" show "$c:$f" > "$WORK/eng/$(basename "$f")"
	done
	echo "## $c ($(git -C "$REPO" rev-parse --short "$c")): 파일 $(ls "$WORK/eng" | wc -l)개"
	"$WORK/render" "$WORK"/eng/*.go | sort
done

echo "# record 렌더 원문 diff: $CAUSE^ → $CAUSE (주석 없는 go/printer 렌더 — 조건식·인자·대입식 전부 포함)"
git -C "$REPO" show "$CAUSE^:$FILE" > "$WORK/p.go"; git -C "$REPO" show "$CAUSE:$FILE" > "$WORK/c.go"
"$WORK/render" -src record "$WORK/p.go" > "$WORK/p.txt"; "$WORK/render" -src record "$WORK/c.go" > "$WORK/c.txt"
diff -u "$WORK/p.txt" "$WORK/c.txt" | sed '1,2d' || true
echo "# record 렌더 원문 diff: $CAUSE → $BASE"
git -C "$REPO" show "$BASE:$FILE" > "$WORK/b.go"; "$WORK/render" -src record "$WORK/b.go" > "$WORK/b.txt"
if cmp -s "$WORK/c.txt" "$WORK/b.txt"; then echo "동일(0 바이트 차이)"; else diff -u "$WORK/c.txt" "$WORK/b.txt" | sed '1,2d'; fi

echo "# record 해시가 바뀐 커밋(부모 대비) — rev-list 전수, 병합은 첫 부모 대비"
git -C "$REPO" rev-list "$CAUSE^..$BASE" | while read -r c; do
	p=$(git -C "$REPO" rev-parse "$c^1")
	bc=$(git -C "$REPO" rev-parse --verify --quiet "$c:$FILE" || echo MISSING)
	bp=$(git -C "$REPO" rev-parse --verify --quiet "$p:$FILE" || echo MISSING)
	[ "$bc" = "$bp" ] && continue
	git -C "$REPO" cat-file blob "$bc" > "$WORK/c.go"; git -C "$REPO" cat-file blob "$bp" > "$WORK/p.go"
	hc=$("$WORK/render" "$WORK/c.go" | awk '$1=="record"{print $2}')
	hp=$("$WORK/render" "$WORK/p.go" | awk '$1=="record"{print $2}')
	if [ "$hc" != "$hp" ]; then
		echo "CHANGED $(git -C "$REPO" log -1 --format='%h %s' "$c") : $hp -> $hc"
	else
		echo "file-only $(git -C "$REPO" log -1 --format='%h %s' "$c") : record 불변 $hc"
	fi
done

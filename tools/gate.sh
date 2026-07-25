#!/usr/bin/env sh
#
# gate.sh — openspec change 완료 게이트
#
# 하나의 change 를 "완료" 로 선언해도 되는지 기계적으로 검사한다. 사람이 눈으로
# 훑는 대신 아래 6개 조건을 전부 통과해야만 exit 0 이 된다.
#
#   1. tasks.md 존재
#   2. 미완료 체크박스 0개
#   3. review.md 존재 (gstack 리뷰 기록)
#   4. make test 통과
#   5. make vet 통과
#   6. make validate 통과
#
# 사용법:
#   bash tools/gate.sh <change-id>
#   make gate CHANGE=<change-id>
#
# 이 저장소는 NTFS 마운트라 실행 비트가 보존되지 않는다(docs/baseline.md 참고).
# 따라서 ./tools/gate.sh 로 직접 실행하지 말고 항상 `bash tools/gate.sh` 형태로
# 인터프리터를 명시해서 호출한다.
#
# 종료 코드: 0 = PASS, 1 = 게이트 실패, 2 = 사용법 오류
set -eu

# 전체 게이트 단계 수 (tasks / 미완료 / review / make test / make vet / make validate)
TOTAL_STEPS=6

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
REPO_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)

usage() {
	cat >&2 <<'USAGE'
usage: bash tools/gate.sh <change-id>
       make gate CHANGE=<change-id>

<change-id> 는 openspec/changes/ 아래의 디렉터리 이름입니다.
USAGE
	if [ -d "$REPO_ROOT/openspec/changes" ]; then
		echo "" >&2
		echo "사용 가능한 change-id:" >&2
		for d in "$REPO_ROOT"/openspec/changes/*/; do
			[ -d "$d" ] || continue
			name=$(basename "$d")
			# archive/ 는 openspec 이 완료된 change 를 옮겨 두는 곳이라 change 가 아니다.
			[ "$name" = "archive" ] && continue
			echo "  $name" >&2
		done
	fi
}

step() {
	echo ""
	echo "==> $1"
}

fail() {
	echo ""
	echo "GATE FAIL: $CHANGE_ID — $1" >&2
	exit 1
}

# ---- 인자 검사 ------------------------------------------------------------

if [ "$#" -ne 1 ] || [ -z "$1" ]; then
	usage
	exit 2
fi

CHANGE_ID="$1"

case "$CHANGE_ID" in
-h | --help | help)
	usage
	exit 2
	;;
esac

cd "$REPO_ROOT"

CHANGE_DIR="openspec/changes/$CHANGE_ID"
TASKS_FILE="$CHANGE_DIR/tasks.md"
REVIEW_FILE="$CHANGE_DIR/review.md"

echo "GATE: $CHANGE_ID"
echo "repo: $REPO_ROOT"

# ---- 1. tasks.md 존재 ------------------------------------------------------

step "1/$TOTAL_STEPS tasks.md 확인"
if [ ! -d "$CHANGE_DIR" ]; then
	echo "change 디렉터리가 없습니다: $CHANGE_DIR" >&2
	usage
	fail "존재하지 않는 change-id"
fi
if [ ! -f "$TASKS_FILE" ]; then
	fail "$TASKS_FILE 가 없습니다"
fi
echo "OK: $TASKS_FILE"

# ---- 2. 미완료 체크박스 ----------------------------------------------------

step "2/$TOTAL_STEPS 미완료 태스크 확인"
# grep -c 는 매치가 0건이면 exit 1 이므로 set -e 에 걸리지 않게 || true 로 받는다.
OPEN_COUNT=$(grep -c '^- \[ \]' "$TASKS_FILE" || true)
[ -n "$OPEN_COUNT" ] || OPEN_COUNT=0

if [ "$OPEN_COUNT" -ne 0 ]; then
	echo "미완료 태스크 $OPEN_COUNT 건:" >&2
	grep -n '^- \[ \]' "$TASKS_FILE" >&2 || true
	fail "미완료 태스크 $OPEN_COUNT 건이 남아 있습니다"
fi
echo "OK: 미완료 태스크 0건"

# ---- 3. review.md 존재 -----------------------------------------------------

step "3/$TOTAL_STEPS gstack 리뷰 기록 확인"
if [ ! -f "$REVIEW_FILE" ]; then
	echo "$REVIEW_FILE 가 없습니다." >&2
	echo "gstack 리뷰 기록 필요 — 리뷰 결과와 결정 사항을 위 경로에 남기세요." >&2
	fail "gstack 리뷰 기록 필요 ($REVIEW_FILE)"
fi
echo "OK: $REVIEW_FILE"

# ---- 4~6. make test / vet / validate --------------------------------------

STEP_NO=4
for target in test vet validate; do
	step "$STEP_NO/$TOTAL_STEPS make $target"
	if ! make "$target"; then
		fail "make $target 실패"
	fi
	echo "OK: make $target"
	STEP_NO=$((STEP_NO + 1))
done

echo ""
echo "GATE PASS: $CHANGE_ID"
exit 0

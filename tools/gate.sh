#!/usr/bin/env sh
#
# gate.sh — openspec change 완료 게이트
#
# 하나의 change 를 "완료" 로 선언해도 되는지 기계적으로 검사한다. 사람이 눈으로
# 훑는 대신 아래 9개 조건을 전부 통과해야만 exit 0 이 된다.
#
#   1. tasks.md 존재
#   2. 미완료 체크박스 0개
#   3. 짝 change (deploy-pair.txt) 도 완료 — 선언이 있을 때만
#   4. review.md 존재 (gstack 리뷰 기록)
#   5. Function Logic Map 산출물 완성 또는 명시적 면제
#   6. make sdd-check 통과
#   7. make test 통과
#   8. make vet 통과
#   9. make validate 통과
#
# 3단계(짝 change)에 대하여
# ------------------------
# 어떤 change 는 혼자 배포하면 배포 전보다 나쁘다. 예: 원장에 임차를 넣는 change 와
# 그 임차를 다시 집어 줄 배달 루프를 만드는 change. 둘 중 하나만 나가면 죽은 발송자의
# 알림이 아무도 안 집는 상태로 남는다. 그런 짝은 문장으로 적어 두는 것으로는 안 지켜져서
# (사람이 잊는다) 여기서 기계가 검사한다.
#
#   openspec/changes/<id>/deploy-pair.txt
#     한 줄에 짝 change id 하나. '#' 로 시작하는 줄과 빈 줄은 무시한다.
#
# 검사 셋:
#   (a) 짝 디렉터리와 tasks.md 가 실제로 있다
#   (b) 짝의 미완료 task 가 0건이다
#   (c) 짝이 보는 배포 단위 구성원이 내가 보는 것과 같다 (상호 선언 + 집합 일치)
#       — 한쪽만 선언하거나 A 는 B·C 를 보는데 B 는 A 만 보면 묶음이 조각난다
#
# 자기 자신을 짝으로 적는 것은 거절한다. 그대로 두면 (b)·(c) 를 자기 자신으로 통과한다.
#
# (b) 의 예외 하나: 짝의 미완료 중 'make gate CHANGE=<pair-id>' 를 포함한 줄 **하나**는
# 세지 않는다. 그 줄까지 세면 A 의 gate 가 B 의 gate task 를 기다리고 B 의 gate 도 A 를
# 기다려서 두 gate 가 영원히 서로를 기다린다. 이 예외는 침묵한 생략이 아니라 여기 적힌
# 예외다. 면제 줄이 둘 이상이면 실패시킨다 — 면제를 늘려 미완료를 숨길 수 없어야 한다.
#
# 선언 파일이 없으면 이 단계는 그냥 통과한다 — 기존 change 전부에 영향이 없다.
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

# 전체 게이트 단계 수
TOTAL_STEPS=9

# 미완료 task 줄의 정규식. CommonMark 는 목록 marker 앞 0~3칸 들여쓰기를 허용하므로
# 그것도 미완료로 센다 — 들여썼다고 안 세면 미완료를 숨길 수 있다(더 세는 쪽이 안전하다).
OPEN_TASK_RE='^ \{0,3\}- \[ \]'

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
OPEN_COUNT=$(grep -c "$OPEN_TASK_RE" "$TASKS_FILE" || true)
[ -n "$OPEN_COUNT" ] || OPEN_COUNT=0

if [ "$OPEN_COUNT" -ne 0 ]; then
	echo "미완료 태스크 $OPEN_COUNT 건:" >&2
	grep -n "$OPEN_TASK_RE" "$TASKS_FILE" >&2 || true
	fail "미완료 태스크 $OPEN_COUNT 건이 남아 있습니다"
fi
echo "OK: 미완료 태스크 0건"

# ---- 3. 짝 change (공동 배포 단위) -----------------------------------------

step "3/$TOTAL_STEPS 짝 change 확인 (deploy-pair.txt)"
PAIR_FILE="$CHANGE_DIR/deploy-pair.txt"

if [ ! -f "$PAIR_FILE" ]; then
	echo "OK: 짝 선언 없음 — 이 change 는 혼자 배포한다"
else
	# 주석과 빈 줄을 걷어낸 짝 목록. 앞뒤 공백과 CR 만 지운다 — 줄 안쪽 공백까지
	# 지우면 "a 099" 가 조용히 "a099" 가 되어 오타가 유효한 id 로 둔갑한다.
	PAIRS=$(sed -e 's/#.*//' -e 's/\r$//' -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//' \
		"$PAIR_FILE" | grep -v '^$' || true)
	if [ -z "$PAIRS" ]; then
		fail "$PAIR_FILE 에 짝 change id 가 하나도 없습니다 (빈 선언은 선언이 아닙니다)"
	fi

	# 이 배포 단위의 구성원 전체 = 나 + 내가 선언한 짝들. 정렬해서 비교용 문자열로 만든다.
	UNIT_SELF=$(printf '%s\n%s\n' "$CHANGE_ID" "$PAIRS" | sort -u)

	for pair in $PAIRS; do
		# id 형태 검사. 디렉터리 이름이므로 경로 조작 문자를 받지 않는다.
		case "$pair" in
		*/* | *' '* | .* | *..*)
			fail "짝 change id 가 올바르지 않습니다: '$pair'"
			;;
		esac

		# 자기 자신을 짝으로 선언하는 것은 묶음이 아니다 — 그대로 두면 완료·상호
		# 선언 검사를 자기 자신으로 통과한다.
		if [ "$pair" = "$CHANGE_ID" ]; then
			fail "$PAIR_FILE 이 자기 자신을 짝으로 선언했습니다 — 묶음은 둘 이상이어야 합니다"
		fi

		echo "짝: $pair"
		PAIR_DIR="openspec/changes/$pair"
		PAIR_TASKS="$PAIR_DIR/tasks.md"

		# (a) 짝이 실재하는가
		if [ ! -f "$PAIR_TASKS" ]; then
			fail "짝 change 의 tasks.md 가 없습니다: $PAIR_TASKS"
		fi

		# (b) 짝의 미완료 0건.
		#
		# 교착을 막는 면제는 짝 자신의 gate 줄 **하나**뿐이고, 그 줄은
		# 'make gate CHANGE=<pair-id>' 를 포함해야 한다. 문자열 'make gate' 만으로
		# 거르면 "make gate 예외 회귀 테스트를 작성한다" 같은 평범한 미완료 task 도
		# 함께 면제되어 짝이 안 끝났는데 통과한다.
		PAIR_EXEMPT_PAT="make gate CHANGE=$pair"
		PAIR_EXEMPT=$(grep "$OPEN_TASK_RE" "$PAIR_TASKS" | grep -c "$PAIR_EXEMPT_PAT" || true)
		[ -n "$PAIR_EXEMPT" ] || PAIR_EXEMPT=0
		if [ "$PAIR_EXEMPT" -gt 1 ]; then
			grep -n "$OPEN_TASK_RE" "$PAIR_TASKS" | grep "$PAIR_EXEMPT_PAT" >&2 || true
			fail "짝 $pair 의 gate 면제 줄이 $PAIR_EXEMPT 건입니다 — 면제는 하나뿐이어야 합니다"
		fi

		PAIR_OPEN=$(grep "$OPEN_TASK_RE" "$PAIR_TASKS" | grep -vc "$PAIR_EXEMPT_PAT" || true)
		[ -n "$PAIR_OPEN" ] || PAIR_OPEN=0
		if [ "$PAIR_OPEN" -ne 0 ]; then
			echo "짝 $pair 의 미완료 태스크 $PAIR_OPEN 건:" >&2
			grep -n "$OPEN_TASK_RE" "$PAIR_TASKS" | grep -v "$PAIR_EXEMPT_PAT" >&2 || true
			fail "짝 $pair 이(가) 아직 완료가 아닙니다 — 같은 배포 단위이므로 함께 끝나야 합니다"
		fi

		# (c) 상호 선언 + 구성원 일치.
		#
		# 상호 선언만으로는 부족하다. A 가 B·C 를 선언하고 B 는 A 만 선언하면 B 의
		# gate 는 C 의 미완료를 안 본다 — 묶음이 조각난다. 그래서 짝이 보는 구성원
		# 집합이 내가 보는 것과 **같아야** 한다.
		PAIR_BACK="$PAIR_DIR/deploy-pair.txt"
		if [ ! -f "$PAIR_BACK" ]; then
			fail "짝 $pair 에 deploy-pair.txt 가 없습니다 (상호 선언 필요)"
		fi
		PAIR_DECL=$(sed -e 's/#.*//' -e 's/\r$//' -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//' \
			"$PAIR_BACK" | grep -v '^$' || true)
		UNIT_PAIR=$(printf '%s\n%s\n' "$pair" "$PAIR_DECL" | sort -u)
		if [ "$UNIT_SELF" != "$UNIT_PAIR" ]; then
			echo "내가 보는 배포 단위:" >&2
			printf '%s\n' "$UNIT_SELF" | sed 's/^/  /' >&2
			echo "$pair 이(가) 보는 배포 단위:" >&2
			printf '%s\n' "$UNIT_PAIR" | sed 's/^/  /' >&2
			fail "짝 $pair 과(와) 배포 단위 구성원이 다릅니다 — 모든 구성원이 같은 집합을 선언해야 합니다"
		fi
		echo "OK: $pair — 완료 + 구성원 일치"
	done
fi

# ---- 4. review.md 존재 -----------------------------------------------------

step "4/$TOTAL_STEPS gstack 리뷰 기록 확인"
if [ ! -f "$REVIEW_FILE" ]; then
	echo "$REVIEW_FILE 가 없습니다." >&2
	echo "gstack 리뷰 기록 필요 — 리뷰 결과와 결정 사항을 위 경로에 남기세요." >&2
	fail "gstack 리뷰 기록 필요 ($REVIEW_FILE)"
fi
echo "OK: $REVIEW_FILE"

# ---- 5. Function Logic Map -------------------------------------------------

STEP_NO=5
step "$STEP_NO/$TOTAL_STEPS Function Logic Map 증거 확인"
if ! python3 tools/logic-map/check_analysis.py --change "$CHANGE_ID"; then
	fail "Function Logic Map 산출물 미완료"
fi
echo "OK: Function Logic Map"

# ---- 6~9. Full SDD / test / vet / validate --------------------------------

STEP_NO=6
for target in sdd-check test vet validate; do
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

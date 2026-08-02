#!/usr/bin/env sh
#
# engine-autostart.sh — 부팅 시 TossOS 엔진을 띄우는 스크립트 (openspec change
# add-engine-runtime, task 2.2).
#
# ============================================================================
#  이 스크립트는 저장소에 **준비만** 되어 있다. 설치·활성화하지 마라.
# ============================================================================
#
# 조회 전용 soak과 달리 엔진은 주문 능력을 가진 프로세스다. "부팅마다 주문 능력
# 프로세스를 자동 기동한다"는 것은 도구 편의가 아니라 **운영 구성 변경**이므로,
# 설치·활성화는 게이트 ON 승인 절차(docs/WORKFLOW.md §0.7)의 항목이다 — 사람이
# 게이트 flip과 함께 명시적으로 승인한다. 리뷰 라운드 1 P2 처분.
#
# 참고로, 지금 설치해도 아무 일도 일어나지 않는다: 이 빌드는 기동 인터록 조항 6
# (ProtectionReady)을 만족할 수 없으므로 `tossctl engine run`은 거부하고 종료한다.
# 그 상태로 설치하면 이 스크립트는 재시작 상한까지 거부를 반복하고 멈춘다.
#
# ---------------------------------------------------------------------------
#  이 스크립트와 Go 상수는 한 기제의 두 반쪽이다
# ---------------------------------------------------------------------------
#
# pgrep 패턴과 재시작 상한은 cmd/tossctl/engineproc.go 의 상수와 같아야 한다.
# 한쪽만 바뀌면 콘솔이 보는 엔진과 이 스크립트가 보는 엔진이 달라지고, 그런 버그는
# 새벽 세 시에만 나타난다. cmd/tossctl/engineproc_test.go 의 drift 테스트가 두 값을
# 대조한다 — soak-autostart.sh 와 soakProcessPattern 이 이미 쓰는 선례 그대로다.
#
#   engineProcessPattern  = "tossctl( .*)? engine run"  → ENGINE_PATTERN
#   engineRestartCap      = 5                           → RESTART_CAP
#   enginelock.StaleAfter = 5m                          → MARKER_STALE_MINUTES
#
# ---------------------------------------------------------------------------
#  설치 방법 (게이트 ON 승인 시에만)
# ---------------------------------------------------------------------------
#
#   cp tools/engine-autostart.sh ~/.local/share/tossos/bin/
#   chmod +x ~/.local/share/tossos/bin/engine-autostart.sh
#   # 그리고 사용자 systemd 유닛이든 로그인 항목이든, 승인된 방식으로 등록한다.
#
# 이 저장소는 NTFS 마운트라 실행 비트가 보존되지 않는다(docs/baseline.md). 저장소
# 안에서 직접 돌릴 일이 있으면 `sh tools/engine-autostart.sh` 형태로 호출한다.
#
set -eu

# --- Go 상수와 대조되는 값 (drift 테스트가 지킨다) ---------------------------

# ENGINE_PATTERN: pgrep -f 로 실행 중인 엔진 **후보**를 찾는 확장 정규식.
#
# 후보다. 소유 판정은 Go 쪽에만 있다 (change a059, design D4). 콘솔은 명령줄에서
# --config-dir 를 되뽑아 journal 디렉터리가 자기 것과 같은 프로세스만 자기 엔진으로
# 인정하지만, 이 스크립트는 그러지 않는다 — 셸에서 그 판정을 재구현하면 Go와 어긋날 수
# 있고, 어긋난 두 반쪽이야말로 a059이 고친 종류의 버그다.
#
# 이 스크립트는 플래그 없이 기본 프로필만 띄우므로, 다른 프로필의 엔진을 보고 물러서는
# 오탐은 "기동하지 않는다" 하나로 끝난다 — 보수적인 방향이다.
ENGINE_PATTERN="tossctl( .*)? engine run"

# RESTART_CAP: 크래시 재시작 상한. 넘으면 멈추고 로그에 사유를 남긴다 —
# 무한 재시작은 아무도 모르는 채로 디스크를 로그로 채우는 방법이다.
RESTART_CAP=5

# MARKER_STALE_MINUTES: 활성 마커를 믿어주는 시간. enginelock.StaleAfter.
MARKER_STALE_MINUTES=5

# --- 경로 --------------------------------------------------------------------

TOSSCTL="${TOSSCTL:-$HOME/.local/bin/tossctl}"
DATA_DIR="${TOSSOS_DATA_DIR:-${XDG_DATA_HOME:-$HOME/.local/share}/tossos}"
LOG="$DATA_DIR/engine.log"
MARKER="$DATA_DIR/engine-run.json"

mkdir -p "$DATA_DIR"

log() {
	echo "[engine-autostart $(date -u +%Y-%m-%dT%H:%M:%SZ)] $*" >>"$LOG"
}

# --- 사전 확인 ----------------------------------------------------------------
#
# 두 가지를 본다: 프로세스가 이미 있는가(pgrep), 그리고 활성 마커가 신선한가.
# 어느 쪽도 배타 기제가 아니다 — 배타는 엔진이 journal 디렉터리에 거는 flock이고,
# 여기서 두 번 확인하는 이유는 "이미 돌고 있다"를 로그에 남기고 조용히 끝내기
# 위해서다. 확인을 건너뛰어도 엔진은 두 개가 되지 않는다. 그냥 두 번째가 거부되고
# 로그가 지저분해질 뿐이다.

if pgrep -f "$ENGINE_PATTERN" >/dev/null 2>&1; then
	log "이미 실행 중이다 (pgrep '$ENGINE_PATTERN'). 아무것도 하지 않는다."
	exit 0
fi

if [ -f "$MARKER" ] && find "$MARKER" -mmin "-$MARKER_STALE_MINUTES" 2>/dev/null | grep -q .; then
	log "활성 마커가 신선하다 ($MARKER, ${MARKER_STALE_MINUTES}분 이내). 아무것도 하지 않는다."
	exit 0
fi

if [ ! -x "$TOSSCTL" ]; then
	log "tossctl을 찾을 수 없다 ($TOSSCTL). TOSSCTL 환경변수로 경로를 지정하라."
	exit 1
fi

# --- 감독 루프 ----------------------------------------------------------------
#
# 엔진이 종료하면 재시작한다. 상한까지만 — 반복해서 죽는 엔진은 일시적 상태가
# 아니다(거부되는 기동이거나 손상된 journal이고, 둘 다 사람이 봐야 한다).
# 상한에 도달하면 멈추고, 활성 마커가 사라지므로 콘솔 대시보드에 "정지"로 뜬다.

attempt=0
while [ "$attempt" -le "$RESTART_CAP" ]; do
	if [ "$attempt" -eq 0 ]; then
		log "엔진을 시작한다: $TOSSCTL engine run"
	else
		log "엔진이 종료했다. 재시작 $attempt/$RESTART_CAP"
	fi

	set +e
	"$TOSSCTL" engine run >>"$LOG" 2>&1
	code=$?
	set -e

	if [ "$code" -eq 0 ]; then
		log "엔진이 정상 종료했다 (exit 0). 재시작하지 않는다."
		exit 0
	fi

	log "엔진이 exit $code 로 종료했다."
	attempt=$((attempt + 1))

	# 즉시 다시 붙지 않는다: 거부로 죽는 엔진은 1초에 다섯 번 거부한다.
	sleep 5
done

log "재시작 상한 $RESTART_CAP 회에 도달했다. 더 시작하지 않는다 — 로그를 확인하라 ($LOG)."
exit 1

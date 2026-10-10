#!/usr/bin/env bash
# a128 측정 세션 기동 래퍼 — TypeSafe 키를 .env 에서 읽어 환경 변수로만 전달
# (키를 저장소·로그·명령줄 인자에 남기지 않음. .env 는 gitignore 확인됨 — 계약 문서 7절).
#
# 사용:
#   ./run-probe.sh KR  [로그파일] [추가 플래그...]   # 당일 KR 세션 (09:00~15:30 KST)
#   ./run-probe.sh US  [로그파일] [추가 플래그...]   # 당일 US 세션 (22:30~05:00 KST, DST)
#   ./run-probe.sh BOTH [로그파일] [추가 플래그...]
#
# 측정일 기동 예 (개장 전 미리 켜 두기 — --wait-open 이 개장까지 대기):
#   ./run-probe.sh KR /tmp/a128-kr.log --wait-open &
#
# 주의: Claude 하네스 background task 는 세션 종료와 함께 죽음("context canceled",
# 2026-10-07 실증 — 1일차 무산). 측정일에는 세션을 장 마감 + horizon 까지 유지하거나
# setsid nohup 으로 사람이 직접 띄울 것. 세션 밖 자가 지속 예약은 금지(분류기 차단 outcome).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
MARKET="${1:?usage: run-probe.sh KR|US|BOTH [logfile] [extra flags...]}"
LOG="${2:-/dev/stdout}"
shift $(( $# > 1 ? 2 : 1 ))

# 키 매핑: 프로브는 TYPESAFE_API_KEY 를 읽고, 사람은 .env 의 JEV_AI_API_KEY 에 보관함
KEY="$(grep '^JEV_AI_API_KEY=' "$ROOT/.env" | cut -d= -f2- || true)"
[ -n "$KEY" ] || { echo "run-probe.sh: JEV_AI_API_KEY 가 $ROOT/.env 에 없음" >&2; exit 1; }

cd "$ROOT"
TYPESAFE_API_KEY="$KEY" exec go run ./tools/a128-jev-shadow-probe run --market "$MARKET" "$@" >>"$LOG" 2>&1

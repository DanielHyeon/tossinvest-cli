#!/usr/bin/env bash
# a115 tasks 1.6 — 콘솔 실측(규칙 13). 엔진 없이, 격리 config 에서 콘솔을 띄워 전략 화면 본문을 잰다.
#   사용: measure.sh <tossctl 바이너리> <clean|dead> <port>
# - clean: 빈 config(0700) — descriptor 부재 → dormant 여야 한다.
# - dead : 같은 config 에 잔재 descriptor(endpoint.json 0600) + 죽은 socket(runtime.sock 0600, 주인 없음) → 도달 불가.
# - GET 만 한다(세션 링크 · /strategy-runtime · /settings/strategy). 버튼·폼·POST 없음(안전 불변식 1·7).
# - 엔진 autostart 는 빈 config 에서 꺼져 있다(runConfiguredEngineAutostart 가 설정을 읽는다) — 그래도 끝에서
#   엔진 프로세스·descriptor 가 새로 생기지 않았음을 확인한다.
set -euo pipefail
BIN=$1; MODE=$2; PORT=$3
D=$(mktemp -d /tmp/a115-console-XXXX); chmod 700 "$D"
OUT=$(mktemp -d /tmp/a115-console-out-XXXX)
cleanup() { [ -n "${PID:-}" ] && kill -INT "$PID" 2>/dev/null && wait "$PID" 2>/dev/null; rm -rf "$D"; }
trap cleanup EXIT
if [ "$MODE" = dead ]; then
  C="$D/.strategy-runtime-read"; mkdir -m 700 "$C"
  python3 - "$C/runtime.sock" <<'PY'
import socket, sys, os
s = socket.socket(socket.AF_UNIX); s.bind(sys.argv[1]); s.close()   # 파일은 남고 listener 는 없다
os.chmod(sys.argv[1], 0o600)
PY
  printf '{"schemaVersion":"tossos.strategy-runtime-unix/v1","socket":"runtime.sock","token":"%s","pid":16}' \
    "$(printf 'a%.0s' $(seq 64))" > "$C/endpoint.json"; chmod 600 "$C/endpoint.json"
fi
"$BIN" --config-dir "$D" console --port "$PORT" > "$OUT/stdout" 2> "$OUT/stderr" &
PID=$!
for _ in $(seq 100); do grep -q 'session=' "$OUT/stdout" "$OUT/stderr" 2>/dev/null && break; sleep 0.1; done
URL=$(grep -ohE "http://127\.0\.0\.1:$PORT/\?session=[A-Za-z0-9_-]+" "$OUT/stdout" "$OUT/stderr" | head -1)
[ -n "$URL" ] || { echo "세션 URL 없음"; cat "$OUT/stderr"; exit 1; }
J="$OUT/jar"
curl -s -o /dev/null -w 'login=%{http_code}\n' -c "$J" -b "$J" "$URL"
curl -s -o "$OUT/strategy.html" -w 'strategy-runtime=%{http_code}\n' -c "$J" -b "$J" "http://127.0.0.1:$PORT/strategy-runtime"
curl -s -o "$OUT/settings.html" -w 'settings/strategy=%{http_code}\n' -c "$J" -b "$J" "http://127.0.0.1:$PORT/settings/strategy"
echo "--- 전략 화면 안내 줄"
grep -oE 'runtime endpoint 미기동[^<]*|runtime projection을 읽지 못했다[^<]*' "$OUT/strategy.html" || echo "(안내 줄 없음)"
echo "--- 시장 상태"
grep -oE 'data-market="(KR|US)" data-market-status="[A-Z_]+"' "$OUT/strategy.html" | sort -u
grep -oE '\b(NOT_CONFIGURED|RUNTIME_UNAVAILABLE)\b' "$OUT/strategy.html" | sort | uniq -c
echo "--- 설정 요약"
grep -oE 'KR OFF/UNKNOWN · US OFF/UNKNOWN — dormant 미배선|읽지 못함 — 전략 lane 판독이 유효하지 않다' "$OUT/settings.html" | sort -u || echo "(요약 없음)"
echo "--- 콘솔 stderr(전략 줄)"
grep -E '전략' "$OUT/stderr" || echo "(전략 경고 없음)"
echo "--- 부작용 확인"
pgrep -f -- "--config-dir $D engine" >/dev/null && echo "엔진 프로세스 있음(!)" || echo "엔진 프로세스 없음"
ls -la "$D" | tail -n +2 | awk '{print $1, $NF}'
rm -rf "$OUT"

#!/usr/bin/env python3
"""a130 4판 — 편입 묶음 1분봉 실측 하네스(데이터 없음 · 경로는 인자).

사용: python3 a130_intake_candles.py <저널 사본(읽기 전용)> <엔진 config 디렉터리>
- 사본의 backfill 모집단(CLOSED · 편입 · COMPLETED · 행 없음)에서 편입 뒤 양의 조정 묶음을 뽑고(created_at > 편입 observed_at),
  묶음마다 봉 이름(floor_minute(created_at) + 1분 — 봉 종료 시각)을 `tossctl ops call candles --backend openapi --config-dir <config>` 로
  읽기 전용 GET(1.2초 간격, adjusted=false, count=1) 해 그 이름의 봉이 있는지 셈.
- 출력은 수뿐(정규장 안 · 밖, 봉 있음 · 없음). 종목 · 가격은 출력하지 않음(F-6). --config-dir 는 엔진과 같은 토큰 캐시(a082).
"""
import datetime, json, sqlite3, subprocess, sys, time
from zoneinfo import ZoneInfo

def main() -> int:
    copy, config = sys.argv[1], sys.argv[2]
    c = sqlite3.connect(f"file:{copy}?mode=ro", uri=True)
    P = lambda s: datetime.datetime.fromisoformat(s.replace("Z", "+00:00"))
    rows = c.execute("""SELECT p.id, p.market, p.symbol, a.observed_at FROM positions p
        JOIN position_adoptions a ON a.id = p.adoption_id
        WHERE p.state = 'CLOSED' AND NOT EXISTS (SELECT 1 FROM trade_outcomes o WHERE o.position_id = p.id)
          AND EXISTS (SELECT 1 FROM exit_events e WHERE e.position_id = p.id AND e.action = 'COMPLETED')""").fetchall()
    counts = {"lots": 0, "regular": 0, "off": 0, "bar": 0, "nobar": 0, "positions": len(rows)}
    for pid, market, symbol, observed in rows:
        tz = ZoneInfo("Asia/Seoul" if market == "kr" else "America/New_York")
        for created, prev, new in c.execute(
                "SELECT created_at, prev_quantity, new_quantity FROM position_adjustments WHERE position_id = ?", (pid,)):
            if float(new) <= float(prev) or P(created) <= P(observed):
                continue
            counts["lots"] += 1
            local = P(created).astimezone(tz)
            minutes = local.hour * 60 + local.minute
            regular = (9 * 60 <= minutes < 15 * 60 + 30) if market == "kr" else (9 * 60 + 30 <= minutes < 16 * 60)
            counts["regular" if regular else "off"] += 1
            label = P(created).replace(second=0, microsecond=0) + datetime.timedelta(minutes=1)
            params = {"symbol": symbol, "interval": "1m", "count": 1,
                      "before": label.strftime("%Y-%m-%dT%H:%M:%SZ"), "adjusted": False}
            out = subprocess.run(["tossctl", "ops", "call", "candles", "--backend", "openapi", "--config-dir", config,
                                  "--params", json.dumps(params)], capture_output=True, text=True)
            got = None
            try:
                candles = json.loads(out.stdout).get("candles") or []
                if candles:
                    got = datetime.datetime.fromisoformat(candles[0]["time"]).astimezone(datetime.timezone.utc)
            except Exception:
                got = None
            counts["bar" if got == label else "nobar"] += 1
            time.sleep(1.2)
    print(json.dumps(counts))
    return 0

if __name__ == "__main__":
    sys.exit(main())

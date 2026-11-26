"""a204 C1 재생 CLI — 파일만 읽고 보고서 JSON 하나를 새로 씀(브로커 · 네트워크 0).

    python3 replay.py --source <dir> [--config <json>] --output <새 파일>

종료 코드: 0 = 보고서 씀, 2 = 설정 · 입력 · 출력 경로 거절(필드 이름이 든 오류를 stderr 에).
"""

from __future__ import annotations

import argparse
import hashlib
import json
import re
import sys
from collections import Counter
from datetime import date, datetime, time
from pathlib import Path

import aee_exits
import c1_entry
import sim_config
import verdict
from c1_entry import KST, Bar

TOOL_VERSION = "a204-c1-replay/v1"
FIXTURE_FORMAT = "a204-fixture/v1"
HISTORY_SESSIONS = 2  # 30분 EMA 워밍업(15 봉) · 5분 ATR(15 봉)에 충분한 전 세션 수
PORT_MANIFEST = Path(__file__).resolve().parent / "stockos_port" / "MANIFEST.json"
LIMITATIONS = [
    "과거 호가 없음 — 진입은 그 분 시가(P0 ask), 매도는 1틱 아래 bid. 체결 가능성 · 잔량 · 장 마감 뒤 보유는 모델링하지 않음.",
    "종목군은 a203 U-A 목록 그대로 — 생존 편향 · 모집단 미래 정보(a203 design §5).",
    "봉 안 가격 순서는 시가 → 저가 → 고가 → 종가 가정(손절 먼저), 시각 해상도 1분.",
    "진입 ATR% 는 StockOS 정의를 재구현한 값이며 StockOS 실제 스냅샷과 대조하지 않음.",
    "낙관 쪽 가정: 진입을 되돌림 봉 직후 분의 시가로 즉시 체결, 청산도 조건 틱에서 즉시 체결(라이브는 시세 주기 · 주문 대기만큼 늦음).",
    "남은 포지션은 마지막 정규장 봉에서 EOD 청산 — 라이브 기본(eod_flatten 꺼짐)의 다음 날 보유 · 갭 위험은 판정에 없음(EOD 건수는 지표에 표시).",
    "R 은 진입마다 R% 로 나눈 값 — R% 가 작은 거래가 평균 R 을 크게 움직일 수 있어 net% 평균을 함께 표시(판정은 R).",
]


class InputError(ValueError):
    pass


SYMBOL_RE = re.compile(r"^[0-9A-Za-z]{1,12}$")  # 경로 이탈 차단 — 종목 코드는 파일 이름 한 조각


class FixtureSource:
    """세션 자료 어댑터(design §5) — 픽스처 형식 a204-fixture/v1. a203 캐시 어댑터는 열린 항목 O1.

    입력 검증은 여기 경계 한 곳에서 함: 형식이 틀린 자료는 InputError(종료 코드 2)로 거절하고 조용히 넘기지 않음.
    """

    def __init__(self, root: Path):
        self.root = Path(root)
        self._read: dict[str, str] = {}  # 읽은 파일 → sha256(입력 지문용)
        m = self._json(self.root / "manifest.json")
        if not isinstance(m, dict) or m.get("format") != FIXTURE_FORMAT:
            raise InputError(f"manifest format 이 {FIXTURE_FORMAT} 가 아님")
        try:
            self._sessions = [date.fromisoformat(d) for d in m["sessions"]]
            self._sample = [date.fromisoformat(d) for d in m["sample"]]
            self._universe = {date.fromisoformat(d): s for d, s in m["universe"].items()}
        except (AttributeError, KeyError, TypeError, ValueError) as exc:
            raise InputError(f"manifest 형식 오류: {exc}") from exc
        if self._sessions != sorted(set(self._sessions)):
            raise InputError("sessions 는 중복 없는 오름차순이어야 함")
        if len(set(self._sample)) != len(self._sample) or not set(self._sample) <= set(self._sessions):
            raise InputError("sample 은 중복 없는 sessions 의 부분집합이어야 함")
        for d in self._sample:
            syms = self._universe.get(d)
            if not isinstance(syms, list) or not syms or len(set(syms)) != len(syms) \
                    or not all(isinstance(x, str) and SYMBOL_RE.match(x) for x in syms):
                raise InputError(f"universe[{d}] 는 중복 없는 종목 코드 목록이어야 함")

    def _json(self, path: Path):
        try:
            raw = path.read_bytes()
            # NaN · Infinity 는 거절(틱 걷기가 끝나지 않음)
            value = json.loads(raw, parse_constant=lambda c: (_ for _ in ()).throw(ValueError(c)))
        except (OSError, ValueError) as exc:
            raise InputError(f"{path}: 읽을 수 없음 {exc}") from exc
        self._read[str(path.relative_to(self.root))] = hashlib.sha256(raw).hexdigest()
        return value

    def fingerprint(self) -> str:
        return hashlib.sha256(json.dumps(sorted(self._read.items())).encode()).hexdigest()

    def sessions(self) -> list[date]:
        return self._sessions

    def sample(self) -> list[date]:
        return sorted(self._sample)

    def universe(self, d: date) -> list[str]:
        return self._universe[d]

    def _load(self, symbol: str, d: date) -> dict | None:
        path = self.root / "bars" / symbol / f"{d.isoformat()}.json"
        if not path.is_file():
            return None
        data = self._json(path)
        if not isinstance(data, dict):
            raise InputError(f"{path}: JSON 객체여야 함")
        return data

    def minute_bars(self, symbol: str, d: date) -> list[Bar] | None:
        data = self._load(symbol, d)
        if not data or not data.get("bars"):
            return None
        out = []
        try:
            for row in data["bars"]:
                hh, mm = (int(x) for x in row[0].split(":"))
                start = datetime.combine(d, time(hh, mm), tzinfo=KST)
                o, h, l, c, v = (float(x) for x in row[1:6])
                if len(row) != 6:
                    raise ValueError("행은 [시각, 시, 고, 저, 종, 거래량] 6칸")
                if not (min(o, h, l, c) > 0 and v >= 0 and h >= max(o, c) and l <= min(o, c)):
                    raise ValueError(f"OHLCV 불일치 {row}")
                if out and start <= out[-1].start:
                    raise ValueError(f"분 시각이 오름차순 · 중복 없음이 아님 {row[0]}")
                out.append(Bar(start, o, h, l, c, v))
        except (TypeError, ValueError, AttributeError) as exc:
            raise InputError(f"bars/{symbol}/{d}: {exc}") from exc
        # 정규장 연속매매 봉 + 15:30 종가 단일가 봉만(15:20~15:29 동시호가 구간은 체결 불가 — design §2)
        out = [b for b in out if time(9, 0) <= b.start.time() < time(15, 20) or b.start.time() == time(15, 30)]
        return out or None

    def prior_close(self, symbol: str, d: date) -> float | None:
        data = self._load(symbol, d)
        value = data.get("prior_close") if data else None
        if value is None:
            return None
        if isinstance(value, bool) or not isinstance(value, (int, float)) or value <= 0:
            raise InputError(f"bars/{symbol}/{d}: prior_close 는 양수여야 함")
        return float(value)


def _replay_symbol(source, symbol: str, d: date, cfg: dict, funnels: dict, trades: dict) -> bool:
    """한 종목-세션 재생. 결손(당일 봉 · 전 세션 종가 · 직전 세션 봉 없음)이면 False."""

    today = source.minute_bars(symbol, d)
    prior = source.prior_close(symbol, d)
    idx = source.sessions().index(d)
    previous = source.minute_bars(symbol, source.sessions()[idx - 1]) if idx > 0 else None
    if today is None or prior is None or previous is None:
        # 직전 세션 봉이 없으면 30분 EMA 워밍업이 불가능 — 거시 거절이 아니라 결손으로 셈(design §5)
        return False
    history = []
    for h in source.sessions()[max(0, idx - HISTORY_SESSIONS):idx - 1]:
        history += source.minute_bars(symbol, h) or []
    history += previous
    entry, funnel = c1_entry.find_entry(today, history, prior)
    for mode in funnels:
        funnels[mode].update(funnel)
        if entry is not None:
            t = aee_exits.simulate(entry, today, cfg, "live" if mode == "live-effective" else "designed")
            trades[mode].append((d, symbol, t))
    return True


def _trade_json(d, symbol, t) -> dict:
    return {
        "session": d.isoformat(), "symbol": symbol, "entry_time": t.entry.time.isoformat(),
        "entry_price": t.entry.price, "stop": t.entry.stop, "level": t.entry.level, "entry_atr_pct": t.entry.atr_pct,
        "risk_pct": t.risk_pct, "net_pct": t.net_pct, "r": t.r, "exit": t.exit_reason,
        "fills": [{"qty": f.qty, "price": f.price, "action": f.action, "reason": f.reason, "time": f.time.isoformat()}
                  for f in t.fills],
    }


def run(source, cfg: dict) -> dict:
    sample = source.sample()
    modes = ("live-effective", "as-designed")
    funnels = {m: Counter() for m in modes}
    trades = {m: [] for m in modes}
    gaps: dict[date, int] = {}
    for d in sample:
        for symbol in source.universe(d):
            if not _replay_symbol(source, symbol, d, cfg, funnels, trades):
                gaps[d] = gaps.get(d, 0) + 1
    universe_sizes = {d: len(source.universe(d)) for d in sample}
    results = {}
    for m in modes:
        rows = [verdict.Row(d, t.entry.time, s, t.r, t.net_pct, t.exit_reason) for d, s, t in trades[m]]
        results[m] = {
            "funnel": dict(funnels[m]),
            "trades": [_trade_json(*x) for x in sorted(trades[m], key=lambda x: (x[2].entry.time, x[1]))],
            "verdict_inputs": verdict.evaluate(sample, rows, universe_sizes, gaps),
        }
    manifest = json.loads(PORT_MANIFEST.read_text(encoding="utf-8"))
    return {
        "tool": TOOL_VERSION,
        "limitations": LIMITATIONS,
        "config_digest": sim_config.digest(cfg),
        "config": cfg,
        "stockos_commit": manifest["source_commit"],
        "sample_fingerprint": hashlib.sha256("\n".join(d.isoformat() for d in sample).encode()).hexdigest(),
        "input_fingerprint": source.fingerprint(),  # 읽은 모든 파일(manifest · 봉)의 sha256 목록의 sha256
        "sample_sessions": len(sample),
        "gaps": {d.isoformat(): n for d, n in sorted(gaps.items())},
        "verdict_mode": "live-effective",
        "verdict": results["live-effective"]["verdict_inputs"],
        "modes": results,
    }


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--source", required=True, type=Path)
    parser.add_argument("--config", type=Path)
    parser.add_argument("--output", required=True, type=Path)
    args = parser.parse_args(argv)
    try:
        cfg = sim_config.load(args.config) if args.config else sim_config.validate({})
        report = run(FixtureSource(args.source), cfg)
        with open(args.output, "x", encoding="utf-8") as fh:  # 덮어쓰기 · 심링크 경합 거절
            json.dump(report, fh, ensure_ascii=False, indent=2, sort_keys=True, allow_nan=False)
    except sim_config.ConfigError as exc:
        print(f"설정 거절: {exc}", file=sys.stderr)
        return 2
    except (InputError, OSError, ValueError) as exc:
        print(f"입력 · 출력 거절: {exc}", file=sys.stderr)
        return 2
    return 0


if __name__ == "__main__":
    sys.exit(main())

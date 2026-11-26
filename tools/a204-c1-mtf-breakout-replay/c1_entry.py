"""C1 진입 규칙 — 30분 추세 필터 + 3분 돌파 · 첫 되돌림 (design §2 사전등록 숫자 그대로).

모든 집계는 「닫힌 봉만」 씀: N분 칸의 끝 시각이 평가 시각 이하일 때만 그 봉을 봄(미래 정보 차단).
"""

from __future__ import annotations

from collections import Counter
from dataclasses import dataclass
from datetime import datetime, time, timedelta, timezone
from statistics import mean, pstdev

KST = timezone(timedelta(hours=9))
SESSION_OPEN = time(9, 0)
SESSION_MINUTES = 390  # 09:00~15:30

# 사전등록 숫자(design §2) — 바꾸면 새 change
MACRO_BARS = 15
EMA_FAST, EMA_SLOW = 5, 15
LOOKBACK_3M = 20
Z_MIN = 2.0
RETEST_BAND_PERMILLE = 3  # ±0.3% — 정수 곱셈으로 비교해 부동소수 경계 오차를 없앰
RETEST_WINDOW = 10
SURGE_PCT = 8
ENTRY_CUTOFF = time(14, 45)
ATR_BARS = 14  # StockOS ATR_SCALED_PULLBACK_LOOKBACK_BARS


@dataclass(frozen=True)
class Bar:
    start: datetime
    open: float
    high: float
    low: float
    close: float
    volume: float
    minutes: int = 1

    @property
    def end(self) -> datetime:
        return self.start + timedelta(minutes=self.minutes)


@dataclass(frozen=True)
class Entry:
    time: datetime  # 진입 1분봉 시작 = 되돌림 3분 봉 끝
    price: float  # 그 1분봉 시가(P0 ask)
    stop: float  # 구조 손절
    level: float  # 돌파 기준선 L
    atr_pct: float | None  # StockOS 정의 진입 ATR%


def kr_tick(price: float) -> int:
    """KRX 호가 단위 — a140 sizing.go:17-34 `krTickMinor` 와 같은 표."""

    for bound, tick in ((2000, 1), (5000, 5), (20000, 10), (50000, 50), (200000, 100), (500000, 500)):
        if price < bound:
            return tick
    return 1000


def aggregate(bars: list[Bar], minutes: int) -> list[Bar]:
    """1분봉을 09:00 기준 N분 칸으로 묶음. 마지막 칸 번호로 잘라 15:30 마감 봉은 마지막 칸에 들어감."""

    last_index = SESSION_MINUTES // minutes - 1
    buckets: dict[tuple, list[Bar]] = {}
    for b in sorted(bars, key=lambda x: x.start):
        open_at = datetime.combine(b.start.date(), SESSION_OPEN, tzinfo=b.start.tzinfo)
        idx = min(last_index, int((b.start - open_at).total_seconds() // 60) // minutes)
        buckets.setdefault((b.start.date(), idx), []).append(b)
    out = []
    for (day, idx), group in sorted(buckets.items()):
        start = datetime.combine(day, SESSION_OPEN, tzinfo=group[0].start.tzinfo) + timedelta(minutes=idx * minutes)
        out.append(Bar(start, group[0].open, max(g.high for g in group), min(g.low for g in group),
                       group[-1].close, sum(g.volume for g in group), minutes))
    return out


def closed(bars: list[Bar], as_of: datetime) -> list[Bar]:
    return [b for b in bars if b.end <= as_of]


def ema(values: list[float], n: int) -> float:
    """α = 2/(n+1), 첫 값으로 시작하는 재귀 EMA(pandas adjust=False)."""

    alpha = 2 / (n + 1)
    value = values[0]
    for v in values[1:]:
        value += alpha * (v - value)
    return value


def macro_ok(closed_30m: list[Bar]) -> bool:
    """가장 최근 닫힌 30분 봉 15개로 EMA5 > EMA15. 15개 미만이면 거래 없음."""

    if len(closed_30m) < MACRO_BARS:
        return False
    closes = [b.close for b in closed_30m[-MACRO_BARS:]]
    return ema(closes, EMA_FAST) > ema(closes, EMA_SLOW)


def volume_z(volume: float, previous: list[float]) -> float | None:
    sd = pstdev(previous)
    if sd == 0:
        return None
    return (volume - mean(previous)) / sd


def breakout_level(bar: Bar, previous: list[Bar]) -> float | None:
    """돌파 봉이면 기준선 L(앞 20봉 최고가), 아니면 None."""

    if len(previous) < LOOKBACK_3M:
        return None
    prev = previous[-LOOKBACK_3M:]
    level = max(b.high for b in prev)
    body = bar.close - bar.open
    z = volume_z(bar.volume, [b.volume for b in prev])
    if bar.close > level and body > 0 and bar.high - bar.close <= body and z is not None and z >= Z_MIN:
        return level
    return None


def surge_excluded(close: float, prior_close: float) -> bool:
    # 정수 가격에서 부동소수 경계 오차를 피하려 곱셈으로 비교함
    return (close - prior_close) * 100 >= SURGE_PCT * prior_close


def atr_pct_5m(bars_1m: list[Bar], as_of: datetime) -> float | None:
    """StockOS 시그널 `atr_pct` 재구현: 닫힌 5분 봉 마지막 15개의 단순 평균 TR(14) / 마지막 종가 × 100.

    근거: strategy_signal_generator.py:443-488(`_atr_pct_from_bars`) · 6320-6345(전 세션 포함, 닫힌 봉만).
    """

    b5 = closed(aggregate(bars_1m, 5), as_of)
    if len(b5) < ATR_BARS + 1 or b5[-1].close <= 0:
        return None
    trs = [max(b.high - b.low, abs(b.high - p.close), abs(b.low - p.close)) for p, b in zip(b5[-ATR_BARS - 1:-1], b5[-ATR_BARS:])]
    return sum(trs) / ATR_BARS / b5[-1].close * 100


def find_entry(today: list[Bar], history: list[Bar], prior_close: float) -> tuple[Entry | None, Counter]:
    """한 종목 · 한 세션의 C1 진입 탐색. 진입은 세션당 최대 1회.

    setup 하나(돌파 → 첫 되돌림)를 끝까지 본 뒤, 그 끝 봉 다음부터 새 돌파를 찾음(재무장).
    """

    funnel: Counter = Counter()
    b3 = aggregate(today, 3)
    b30 = aggregate(history + today, 30)
    minute_by_start = {b.start: b for b in today}
    j = LOOKBACK_3M
    while j < len(b3):
        level = breakout_level(b3[j], b3[:j])
        if level is None:
            j += 1
            continue
        funnel["breakout"] += 1
        window = b3[j + 1:j + 1 + RETEST_WINDOW]
        first = next((k for k, b in enumerate(window) if b.low * 1000 <= level * (1000 + RETEST_BAND_PERMILLE)), None)
        if first is None:
            # 창이 세션 끝에서 잘렸어도 10봉 안에 되돌림이 없으면 시한 초과로 셈
            funnel["timeout"] += 1
            j += 1 + RETEST_WINDOW
            continue
        rb = window[first]
        j_end = j + 1 + first
        if not (rb.low * 1000 >= level * (1000 - RETEST_BAND_PERMILLE) and rb.close > level and rb.close > rb.open):
            funnel["pullback_failed"] += 1
            j = j_end + 1
            continue
        signal_at = rb.end
        if signal_at.time() > ENTRY_CUTOFF:
            funnel["cutoff"] += 1
            return None, funnel
        if not macro_ok(closed(b30, signal_at)):
            funnel["macro_rejected"] += 1
        elif surge_excluded(rb.close, prior_close):
            funnel["surge_rejected"] += 1
        elif signal_at not in minute_by_start:
            funnel["no_entry_bar"] += 1
        else:
            price = minute_by_start[signal_at].open
            stop = rb.low - kr_tick(rb.low - 1)  # 한 틱 아래 — 가격대 경계에서는 아래 가격대의 틱
            if stop >= price:
                stop = price * 0.97  # AEE:2391-2404 `_initial_stop_price` 폴백(진입가 − 3%)
            funnel["entry"] += 1
            return Entry(signal_at, price, stop, level, atr_pct_5m(history + today, signal_at)), funnel
        j = j_end + 1
    return None, funnel

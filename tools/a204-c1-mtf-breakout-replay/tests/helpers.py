"""시험용 봉 생성기 — 3분 봉 명세를 1분 봉 셋으로 풀어 집계가 같은 OHLCV 를 되돌리게 만듦."""

from __future__ import annotations

from datetime import date, datetime, timedelta

from c1_entry import KST, Bar

DAY = date(2026, 6, 10)
PREV = date(2026, 6, 9)


def at(d: date, hh: int, mm: int) -> datetime:
    return datetime(d.year, d.month, d.day, hh, mm, tzinfo=KST)


def expand3(d: date, index: int, o, h, l, c, v) -> list[Bar]:
    """09:00 기준 index 번째 3분 칸을 1분 봉 셋으로 풂(집계하면 o,h,l,c,v 가 됨)."""

    start = at(d, 9, 0) + timedelta(minutes=3 * index)
    v1, v2 = v // 3, v // 3
    v3 = v - v1 - v2
    return [
        Bar(start, o, h, l, o, v1),
        Bar(start + timedelta(minutes=1), o, o, o, o, v2),
        Bar(start + timedelta(minutes=2), o, max(o, c), min(o, c), c, v3),
    ]


def flat_session(d: date, start_price: int, end_price: int, n3: int = 130) -> list[Bar]:
    """시가 = 종가인(돌파 불가) 3분 봉 n3 개 — 값이 start→end 로 직선 이동, 거래량 3000/3600 교대."""

    bars: list[Bar] = []
    for k in range(n3):
        p = start_price + (end_price - start_price) * k // max(1, n3 - 1)
        p -= p % 10
        bars += expand3(d, k, p, p + 50, p - 50, p, 3000 if k % 2 == 0 else 3600)
    return bars


def setup_day(
    rise_bars: int = 0,
    breakout=(10000, 10120, 9990, 10100, 4500),
    pullbacks=((10070, 10110, 10060, 10100, 3300),),
    tail_minutes: int = 120,
    tail_price: int = 10100,
) -> list[Bar]:
    """당일: 오르는 채움 봉 rise_bars 개 → 평평한 20 봉(10000, 고 10050) → 돌파 봉 → 되돌림 봉들 → 1분 꼬리."""

    bars: list[Bar] = []
    k = 0
    for i in range(rise_bars):
        p = 9000 + 1000 * i // max(1, rise_bars)
        p -= p % 10
        bars += expand3(DAY, k, p, p + 50, p - 50, p, 3000 if k % 2 == 0 else 3600)
        k += 1
    for _ in range(20):
        bars += expand3(DAY, k, 10000, 10050, 9950, 10000, 3000 if k % 2 == 0 else 3600)
        k += 1
    bars += expand3(DAY, k, *breakout)
    k += 1
    for spec in pullbacks:
        bars += expand3(DAY, k, *spec)
        k += 1
    start = at(DAY, 9, 0) + timedelta(minutes=3 * k)
    end = at(DAY, 15, 30)
    while start <= end and tail_minutes > 0:
        bars.append(Bar(start, tail_price, tail_price, tail_price, tail_price, 1000))
        start += timedelta(minutes=1)
        tail_minutes -= 1
    return bars


def rising_prev() -> list[Bar]:
    return flat_session(PREV, 9000, 9900)

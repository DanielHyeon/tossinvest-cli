# stockos-port: source=packages/trading/stockos_trading/indicators/__init__.py commit=fcad869ecc531a6de80777f59b6fb842a2891101 — verbatim copy, do not edit
from __future__ import annotations

from math import sqrt
from statistics import mean

from ..models import BollingerBand, Candle, IchimokuSnapshot


def require_window(values: list[float], period: int, name: str) -> list[float]:
    if period <= 0:
        raise ValueError(f"{name} period must be positive")
    if len(values) < period:
        raise ValueError(f"{name} requires at least {period} values")
    return values[-period:]


def simple_moving_average(values: list[float], period: int) -> float:
    return mean(require_window(values, period, "sma"))


def bollinger_band(closes: list[float], period: int = 20, multiplier: float = 2.0) -> BollingerBand:
    window = require_window(closes, period, "bollinger")
    middle = mean(window)
    variance = sum((value - middle) ** 2 for value in window) / period
    stddev = sqrt(variance)
    return BollingerBand(
        middle=middle,
        upper=middle + multiplier * stddev,
        lower=middle - multiplier * stddev,
        stddev=stddev,
    )


def cci(candles: list[Candle], period: int = 20) -> float:
    window = require_window([c.typical_price for c in candles], period, "cci")
    average = mean(window)
    mean_deviation = mean(abs(price - average) for price in window)
    if mean_deviation == 0:
        return 0.0
    return (window[-1] - average) / (0.015 * mean_deviation)


def heikin_ashi(candles: list[Candle]) -> list[Candle]:
    if not candles:
        return []

    result: list[Candle] = []
    for index, candle in enumerate(candles):
        ha_close = (candle.open + candle.high + candle.low + candle.close) / 4
        if index == 0:
            ha_open = (candle.open + candle.close) / 2
        else:
            previous = result[-1]
            ha_open = (previous.open + previous.close) / 2
        ha_high = max(candle.high, ha_open, ha_close)
        ha_low = min(candle.low, ha_open, ha_close)
        result.append(
            Candle(
                symbol=candle.symbol,
                timestamp=candle.timestamp,
                open=ha_open,
                high=ha_high,
                low=ha_low,
                close=ha_close,
                volume=candle.volume,
            )
        )
    return result


def ichimoku(candles: list[Candle]) -> IchimokuSnapshot:
    highs = [c.high for c in candles]
    lows = [c.low for c in candles]
    closes = [c.close for c in candles]

    def midpoint(period: int) -> float | None:
        if len(candles) < period:
            return None
        return (max(highs[-period:]) + min(lows[-period:])) / 2

    conversion = midpoint(9)
    base = midpoint(26)
    span_a = (conversion + base) / 2 if conversion is not None and base is not None else None
    span_b = midpoint(52)
    lagging = closes[-26] if len(closes) >= 26 else None
    return IchimokuSnapshot(
        conversion_line=conversion,
        base_line=base,
        leading_span_a=span_a,
        leading_span_b=span_b,
        lagging_span=lagging,
    )


def crossed_above(previous_value: float, current_value: float, previous_level: float, current_level: float) -> bool:
    return previous_value <= previous_level and current_value > current_level


def crossed_below(previous_value: float, current_value: float, previous_level: float, current_level: float) -> bool:
    return previous_value >= previous_level and current_value < current_level


def fractal_rebound_target(previous_drop: float, previous_rebound: float, current_drop: float) -> float:
    if previous_drop <= 0:
        raise ValueError("previous_drop must be positive")
    if previous_rebound < 0 or current_drop < 0:
        raise ValueError("rebound and current_drop must be non-negative")
    return current_drop * (previous_rebound / previous_drop)


def vwap(candles: list[Candle]) -> float | None:
    # Caller-defined session: the slice itself is treated as the window.
    # Returning None on zero cumulative volume avoids ZeroDivisionError
    # while preserving the "no silent partial result" invariant downstream
    # gates rely on (a None vwap means BLOCK, not 0).
    if not candles:
        return None
    total_volume = 0.0
    weighted_sum = 0.0
    for c in candles:
        total_volume += c.volume
        weighted_sum += c.typical_price * c.volume
    if total_volume == 0:
        return None
    return weighted_sum / total_volume


def vwap_series(candles: list[Candle]) -> list[float | None]:
    # Streaming form of vwap() with the same look-ahead invariant:
    # series[k] only depends on candles[:k+1]. Replay harness uses this to
    # assert vwap_series(c)[k] == vwap(c[:k+1]) for every k.
    result: list[float | None] = []
    total_volume = 0.0
    weighted_sum = 0.0
    for c in candles:
        total_volume += c.volume
        weighted_sum += c.typical_price * c.volume
        result.append(weighted_sum / total_volume if total_volume > 0 else None)
    return result


def ema(values: list[float], period: int) -> float | None:
    # Partial-window EMA is silently look-ahead biased (early values get
    # over-weighted), so we refuse to return one and force the caller to
    # treat the indicator as "not ready" until it has full seed data.
    if period < 1:
        raise ValueError("ema period must be >= 1")
    if len(values) < period:
        return None
    alpha = 2 / (period + 1)
    current = sum(values[:period]) / period
    for value in values[period:]:
        current = alpha * value + (1 - alpha) * current
    return current


def ema_series(values: list[float], period: int) -> list[float | None]:
    # Mirror of ema() but materialises every intermediate value so strategy
    # code can compare EMA crossovers without recomputing the seed each bar.
    if period < 1:
        raise ValueError("ema period must be >= 1")
    result: list[float | None] = [None] * len(values)
    if len(values) < period:
        return result
    alpha = 2 / (period + 1)
    current = sum(values[:period]) / period
    # WHY: SMA seed lands at index ``period - 1`` (the first index where a
    # full window exists); the recursion then updates from index ``period``
    # onward. The off-by-one between seed placement and loop start is
    # intentional and matches the scalar ``ema()`` semantics.
    result[period - 1] = current
    for index in range(period, len(values)):
        current = alpha * values[index] + (1 - alpha) * current
        result[index] = current
    return result


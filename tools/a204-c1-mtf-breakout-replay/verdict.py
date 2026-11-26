"""판정 계산 — design §7 사전등록 규칙 그대로(숫자를 바꾸면 새 change)."""

from __future__ import annotations

from datetime import date, datetime
from statistics import mean
from typing import NamedTuple

HOLDOUT_SESSIONS = 50
MIN_TRADES = 40
MIN_MEAN_R = 0.10
MIN_POSITIVE_SESSION_RATIO = 0.55
MAX_CONSECUTIVE_LOSSES = 6
MAX_DRAWDOWN_R = 8.0
MAX_GAP_RATIO = 0.20
SCORE = {"c1_mean_r": 16, "c2c3": 10, "c4_losses_and_drawdown": 8, "c5_dev_same_sign": 6}


class Row(NamedTuple):
    session: date
    time: datetime
    symbol: str
    r: float
    net_pct: float = 0.0  # 보고용(판정은 r)
    exit: str = ""


def metrics(rows: list[Row]) -> dict:
    ordered = sorted(rows, key=lambda x: (x.time, x.symbol))
    by_session: dict[date, list[float]] = {}
    for x in ordered:
        by_session.setdefault(x.session, []).append(x.r)
    streak = worst_streak = 0
    equity = peak = drawdown = 0.0
    for x in ordered:
        streak = streak + 1 if x.r < 0 else 0
        worst_streak = max(worst_streak, streak)
        equity += x.r
        peak = max(peak, equity)
        drawdown = max(drawdown, peak - equity)
    daily = [mean(v) for v in by_session.values()]
    return {
        "trades": len(ordered),
        "mean_r": mean(x.r for x in ordered) if ordered else None,
        "sessions_with_trades": len(daily),
        "positive_session_ratio": sum(1 for d in daily if d > 0) / len(daily) if daily else None,
        "max_consecutive_losses": worst_streak,
        "max_drawdown_r": drawdown,
        "mean_net_pct": mean(x.net_pct for x in ordered) if ordered else None,  # 보고용 — R% 가 작은 거래의 꼬리 확인
        "eod_exits": sum(1 for x in ordered if x.exit == "EOD"),
    }


def _same_sign(a: float | None, b: float | None) -> bool:
    return a is not None and b is not None and ((a > 0 and b > 0) or (a < 0 and b < 0))


def evaluate(sample: list[date], rows: list[Row], universe_sizes: dict[date, int], gaps: dict[date, int]) -> dict:
    sessions = sorted(set(sample))  # 같은 날짜를 두 번 세면 hold-out 이 줄고 거래가 겹침
    holdout = set(sessions[-HOLDOUT_SESSIONS:])
    hold = metrics([x for x in rows if x.session in holdout])
    dev = metrics([x for x in rows if x.session not in holdout and x.session in set(sessions)])
    slots = sum(universe_sizes.get(d, 0) for d in holdout)
    gap_ratio = sum(gaps.get(d, 0) for d in holdout) / slots if slots else 1.0
    checks = {
        "c1_mean_r": hold["mean_r"] is not None and hold["mean_r"] >= MIN_MEAN_R,
        "c2_trades": hold["trades"] >= MIN_TRADES,
        "c3_positive_sessions": hold["positive_session_ratio"] is not None
        and hold["positive_session_ratio"] >= MIN_POSITIVE_SESSION_RATIO,
        "c4_losses_and_drawdown": hold["max_consecutive_losses"] <= MAX_CONSECUTIVE_LOSSES
        and hold["max_drawdown_r"] <= MAX_DRAWDOWN_R,
        "c5_dev_same_sign": _same_sign(dev["mean_r"], hold["mean_r"]),
    }
    insufficient = []
    if hold["trades"] < MIN_TRADES:
        insufficient.append(f"hold-out 거래 {hold['trades']} < {MIN_TRADES}")
    if gap_ratio > MAX_GAP_RATIO:
        insufficient.append(f"hold-out 결손 {gap_ratio:.4f} > {MAX_GAP_RATIO}")
    if len(sessions) <= HOLDOUT_SESSIONS:
        insufficient.append(f"표본 세션 {len(sessions)} ≤ {HOLDOUT_SESSIONS} — 개발 절반 없음")
    if insufficient:
        result, score = "INSUFFICIENT", 0
    else:
        result = "PASS" if all(checks.values()) else "FAIL"
        score = (SCORE["c1_mean_r"] * checks["c1_mean_r"]
                 + SCORE["c2c3"] * (checks["c2_trades"] and checks["c3_positive_sessions"])
                 + SCORE["c4_losses_and_drawdown"] * checks["c4_losses_and_drawdown"]
                 + SCORE["c5_dev_same_sign"] * checks["c5_dev_same_sign"])
    return {
        "verdict": result,
        "insufficient_reasons": insufficient,
        "checks": checks,
        "score_e": score,
        "holdout": hold,
        "development": dev,
        "holdout_gap_ratio": gap_ratio,
        "split": {
            "holdout_first": sessions[-HOLDOUT_SESSIONS].isoformat() if len(sessions) >= HOLDOUT_SESSIONS else None,
            "holdout_sessions": len(holdout),
            "development_sessions": len(sessions) - len(holdout),
        },
    }

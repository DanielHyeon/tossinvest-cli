"""StockOS AEE(apps/api/stockos_api/auto_exit_execution.py @ fcad869e) 청산 오케스트레이션 재구현.

청산 판단 자체는 원문 복사본 `stockos_port.exit_strategy.evaluate_exit` 가 함. 이 파일은 AEE 가 그 함수 둘레에서
하는 일(고점 추적 · floor ratchet · 문맥 조립 · 세션 조임 · TP2 해제 · 보호 floor 이탈 · floor 올림 · 부분익절 표시)을
같은 순서로 다시 씀(design §4.2). 줄 번호는 모두 AEE 원본 기준.
"""

from __future__ import annotations

from dataclasses import dataclass, field, replace
from datetime import datetime, time

from c1_entry import KST, Bar, Entry, ema, kr_tick
from stockos_port.costs import Market, net_pnl
from stockos_port.exit_strategy import (
    ExitAction,
    ExitContext,
    ExitPolicy,
    adapt_exit_policy,
    compute_effective_hard_stop_pct,
    evaluate_exit,
    is_orderable_exit_action,
)
from stockos_port.ladder import profile
from stockos_port.profit_ladder import LadderState

# shortcut: 명목 수량 — StockOS 의 int(수량×비율) 반올림을 무시할 만큼 크게 둔 값, 결과는 %·R 이라 수량과 무관
Q0 = 10_000
MARKET = Market.KOSPI  # KOSDAQ 와 수수료 · 세율 같음(costs.py:73-90)
RUNNER_DISABLED_SECOND_TP_PCT = 10_000.0  # AEE:286
EXIT_FALLBACK_ATR_PCT = 0.8  # AEE:1956
LADDERS = {  # AEE:6766-6806 → ladder/registry.py DEFAULT_REGISTRY
    "COMMON_LADDER_BALANCED": profile.BALANCED.policy,
    "COMMON_LADDER_CONSERVATIVE": profile.CONSERVATIVE.policy,
    "COMMON_LADDER_DEFENSIVE": profile.DEFENSIVE.policy,
    "COMMON_LADDER_TREND": profile.TREND.policy,
    "COMMON_LADDER_RUNNER": profile.RUNNER.policy,
}
LATE, CLOSE = "LATE_SESSION_PROTECT", "CLOSE_MANAGEMENT"


@dataclass(frozen=True)
class Fill:
    qty: int
    price: float
    action: str
    reason: str
    time: datetime


@dataclass
class PositionState:
    entry: Entry
    high: float
    qty: int = Q0
    current_floor: float | None = None
    protected_floor_enabled: bool = False
    partial_tp_done: bool = False
    breach_started_at: datetime | None = None
    breach_at_price: float | None = None
    ladder_state: LadderState | None = None
    fills: list[Fill] = field(default_factory=list)

    @classmethod
    def open(cls, entry: Entry, ladder_policy=None) -> "PositionState":
        state = LadderState(policy_id=ladder_policy.policy_id) if ladder_policy is not None else None
        return cls(entry=entry, high=entry.price, ladder_state=state)


@dataclass(frozen=True)
class Trade:
    entry: Entry
    fills: list[Fill]
    net_pct: float
    risk_pct: float

    @property
    def r(self) -> float:
        return self.net_pct / self.risk_pct

    @property
    def exit_reason(self) -> str:
        return self.fills[-1].action


def session_bucket(moment: datetime) -> str | None:
    """StockOS gates.py:232-235 — 14:50 ≤ t < 15:20 마감 보호, 15:20 ≤ t ≤ 15:30 마감 관리."""

    t = moment.astimezone(KST).time()
    if time(14, 50) <= t < time(15, 20):
        return LATE
    if time(15, 20) <= t <= time(15, 30):
        return CLOSE
    return None


def runtime_policy(cfg: dict) -> ExitPolicy:
    """설정 → ExitPolicy. AEE:1993-2053 + KRX 익절 AEE:2055-2091. 명시 안 한 필드는 ExitPolicy 기본값."""

    return ExitPolicy(
        hard_stop_grace_minutes=cfg["exit_hard_stop_grace_minutes"],
        breakeven_atr_multiplier=cfg["exit_breakeven_atr_multiplier"],
        breakeven_atr_min_gap_pct=0.0,  # hardening_runtime_config.py:27 기본
        stop_confirm_seconds=cfg["exit_stop_confirm_seconds"],
        stop_atr_multiplier=cfg["exit_stop_atr_multiplier"],
        use_entry_atr_pct=False,  # AEE:1950 라이브 기본
        consume_signal_tp=False,  # AEE:1953 라이브 기본
        hard_stop_grace_catastrophe_pct=cfg["exit_hard_stop_grace_catastrophe_pct"],
        hard_stop_loss_pct=cfg["exit_hard_stop_loss_pct"],
        max_hold_minutes=cfg["max_hold_minutes"],
        weak_trend_stop_loss_pct=cfg["exit_weak_trend_stop_loss_pct"],
        trailing_stop_gap_pct=cfg["exit_trailing_stop_gap_pct"],
        first_take_profit_pct=cfg["krx_first_take_profit_pct"],
        second_take_profit_pct=cfg["krx_second_take_profit_pct"],
        partial_exit_ratio=cfg["krx_partial_exit_ratio"],
        rapid_profit_lock_base_capture_ratio=cfg["rapid_profit_lock_base_capture_ratio"],
        rapid_profit_lock_strong_capture_ratio=cfg["rapid_profit_lock_strong_capture_ratio"],
        rapid_profit_lock_tight_capture_ratio=cfg["rapid_profit_lock_tight_capture_ratio"],
    )


def session_adjusted(base: ExitPolicy, bucket: str | None) -> ExitPolicy:
    """AEE:2105-2176 `_session_adjusted_exit_policy` 값 그대로."""

    if bucket == LATE:
        return replace(
            base,
            breakeven_after_pct=min(base.breakeven_after_pct, 0.3),
            breakeven_atr_multiplier=0.0,
            stop_confirm_seconds=0.0,
            stop_atr_multiplier=0.0,
            first_take_profit_pct=min(base.first_take_profit_pct, 0.8),
            second_take_profit_pct=min(base.second_take_profit_pct, 1.5),
            trailing_stop_gap_pct=min(base.trailing_stop_gap_pct, 0.45),
            max_hold_minutes=min(base.max_hold_minutes, 30),
            partial_exit_ratio=max(base.partial_exit_ratio, 0.7),
            rapid_profit_lock_capture_ratio=max(base.rapid_profit_lock_capture_ratio, 0.7),
            rapid_profit_lock_base_capture_ratio=max(base.rapid_profit_lock_base_capture_ratio, 0.75),
            rapid_profit_lock_strong_capture_ratio=max(base.rapid_profit_lock_strong_capture_ratio, 0.8),
            rapid_profit_lock_tight_capture_ratio=max(base.rapid_profit_lock_tight_capture_ratio, 0.85),
        )
    if bucket == CLOSE:
        return replace(
            base,
            breakeven_after_pct=min(base.breakeven_after_pct, 0.15),
            breakeven_atr_multiplier=0.0,
            stop_confirm_seconds=0.0,
            stop_atr_multiplier=0.0,
            first_take_profit_pct=min(base.first_take_profit_pct, 0.4),
            second_take_profit_pct=min(base.second_take_profit_pct, 0.8),
            trailing_stop_gap_pct=min(base.trailing_stop_gap_pct, 0.3),
            max_hold_minutes=1,
            partial_exit_ratio=1.0,
            rapid_profit_lock_capture_ratio=max(base.rapid_profit_lock_capture_ratio, 0.8),
            rapid_profit_lock_base_capture_ratio=max(base.rapid_profit_lock_base_capture_ratio, 0.85),
            rapid_profit_lock_strong_capture_ratio=max(base.rapid_profit_lock_strong_capture_ratio, 0.9),
            rapid_profit_lock_tight_capture_ratio=max(base.rapid_profit_lock_tight_capture_ratio, 0.95),
        )
    return base


def effective_policy(cfg: dict, moment: datetime) -> ExitPolicy:
    # 순서가 중요함: 세션 조임 → TP2 해제(AEE:1487-1495, 6876-6884 — StockOS 소유 포지션)
    adjusted = session_adjusted(runtime_policy(cfg), session_bucket(moment))
    return replace(adjusted, second_take_profit_pct=RUNNER_DISABLED_SECOND_TP_PCT)


def raise_trailing_floor(st: PositionState, high: float, trail_stop_pct: float) -> None:
    """AEE:2518-2542 — high×(1−trail%) 가 현재 floor · 초기 손절 중 큰 값을 넘을 때만 floor 를 올림(내리지 않음)."""

    if trail_stop_pct <= 0:
        return
    trail_floor = high * (1 - trail_stop_pct / 100)
    floor_to_beat = max(v for v in (st.current_floor, st.entry.stop) if v is not None)
    if trail_floor <= floor_to_beat:
        return
    st.current_floor = trail_floor
    st.protected_floor_enabled = True


def _support_breach_elapsed(st: PositionState, support: float, price: float, moment: datetime) -> float | None:
    """AEE:2585-2628 — support 하회 지속 초. support 가 바뀌면(래칫) 타이머 재시작."""

    if price >= support:
        st.breach_started_at = st.breach_at_price = None
        return None
    if st.breach_started_at is None or st.breach_at_price != support:
        st.breach_started_at, st.breach_at_price = moment, support
        return 0.0
    return max(0.0, (moment - st.breach_started_at).total_seconds())


def _breach_confirmed(ctx: ExitContext, policy: ExitPolicy) -> bool:
    """AEE:2630-2645 `_stop_breach_confirmed`."""

    if ctx.day_change_pct <= policy.hard_stop_grace_catastrophe_pct:
        return True
    return ctx.support_breach_elapsed_seconds is None or ctx.support_breach_elapsed_seconds >= policy.stop_confirm_seconds


def _apply_stop_raise(st: PositionState, decision) -> None:
    """AEE:3222-3256 — breakeven · profit lock · early-surge 부분익절의 stop_price 로 floor 를 올림(높을 때만)."""

    if decision.stop_price is None:
        return
    early_surge_partial = (
        decision.action is ExitAction.TAKE_PARTIAL_PROFIT and decision.reason == "early_surge_fast_partial_profit"
    )
    if not early_surge_partial and decision.action not in (ExitAction.MOVE_STOP_TO_BREAKEVEN, ExitAction.RAISE_PROFIT_LOCK):
        return
    if st.current_floor is not None and st.current_floor >= decision.stop_price:
        return
    st.current_floor = decision.stop_price
    st.protected_floor_enabled = True


def _sell(st: PositionState, qty: int, price: float, action: str, reason: str, moment: datetime) -> Fill:
    # P0 1틱 스프레드: 매도는 그 가격의 1틱 아래 bid 로 체결(a203 design §4)
    fill = Fill(qty, price - kr_tick(price - 1), action, reason, moment)  # 가격대 경계에서는 아래 가격대의 틱
    st.qty -= qty
    st.fills.append(fill)
    return fill


def on_tick(st, price, moment, cfg, policy, ladder_policy, extra=None) -> Fill | None:
    """시세 한 번에 대한 AEE 평가(AEE:1459-1680 순서). 결정은 틱당 하나."""

    e = st.entry.price
    st.high = max(e, st.high, price, (extra or {}).get("bar_high") or 0)  # AEE:2464-2516
    raise_trailing_floor(st, st.high, cfg["exit_trail_stop_pct"])  # AEE:1459-1469
    support = max(v for v in (st.entry.stop, st.current_floor) if v is not None)  # AEE:2669-2673
    ctx = ExitContext(  # AEE:2717-2781
        entry_price=e,
        current_price=price,
        high_since_entry=max(st.high, price, e),
        quantity=st.qty,
        market=MARKET,
        minutes_held=max(0, int((moment - st.entry.time).total_seconds() // 60)),
        max_hold_exit_enabled=False,
        support_price=support,
        support_breach_elapsed_seconds=_support_breach_elapsed(st, support, price, moment),
        day_change_pct=(price - e) / e * 100,
        atr_pct=EXIT_FALLBACK_ATR_PCT,
        breakeven_atr_pct=st.entry.atr_pct,
        moment=moment,
        partial_tp_done=st.partial_tp_done,
        **(extra or {}),
    )
    if cfg["eod_flatten_enabled"] and session_bucket(moment) == CLOSE:  # AEE:1553-1564 · 1707-1755
        return _sell(st, st.qty, price, "EOD_FLATTEN", "eod_flatten", moment)
    if st.protected_floor_enabled and st.current_floor is not None and price <= st.current_floor and _breach_confirmed(ctx, policy):
        # AEE:2811-2866 — 전량, 라벨은 net 손익 부호(AEE:2843-2847)
        action = "TAKE_FULL_PROFIT" if net_pnl(e, price, st.qty, MARKET) > 0 else "STOP_LOSS"
        return _sell(st, st.qty, price, action, "protected_floor_breached", moment)
    decision = evaluate_exit(
        ctx, policy, ladder_policy=ladder_policy, ladder_state=st.ladder_state if ladder_policy is not None else None
    )
    if ladder_policy is not None and decision.next_ladder_state is not None:
        st.ladder_state = decision.next_ladder_state  # AEE:1622-1629
    _apply_stop_raise(st, decision)  # AEE:1648
    if not is_orderable_exit_action(decision.action) or decision.quantity_to_sell <= 0:
        return None
    if decision.action is ExitAction.TAKE_PARTIAL_PROFIT:
        st.partial_tp_done = True  # AEE:6735-6744
    return _sell(st, min(decision.quantity_to_sell, st.qty), price, decision.action.value, decision.reason, moment)


def walk(a: float, b: float) -> list[float]:
    """a 에서 b 까지 호가 단위로 한 틱씩(a 제외, b 포함). 내려갈 때는 바로 아래 가격대의 틱을 씀."""

    out, p = [], a
    while p != b:
        p = max(b, p - kr_tick(p - 1)) if b < p else min(b, p + kr_tick(p))
        out.append(p)
    return out


def intrabar_path(o: float, h: float, l: float, c: float) -> list[float]:
    """보수 경로 시가 → 저가 → 고가 → 종가: 한 봉이 손절과 익절을 다 건드리면 손절이 먼저(STOP_FIRST)."""

    return [o] + walk(o, l) + walk(l, h) + walk(h, c)


def risk_pct(entry: Entry, cfg: dict) -> float:
    """1R(design §6) = 진입가에서 가장 가까운 보호선(구조 손절 · 유효 하드 손절)까지의 %."""

    ctx = ExitContext(entry_price=entry.price, current_price=entry.price, high_since_entry=entry.price, quantity=Q0,
                      market=MARKET, minutes_held=0, atr_pct=EXIT_FALLBACK_ATR_PCT, breakeven_atr_pct=entry.atr_pct)
    active = adapt_exit_policy(ctx, effective_policy(cfg, entry.time))
    hard = compute_effective_hard_stop_pct(active.hard_stop_loss_pct, entry.atr_pct, active.stop_atr_multiplier)
    protective = max(entry.stop, entry.price * (1 + hard / 100))
    return (entry.price - protective) / entry.price * 100


def simulate(entry: Entry, session_bars: list[Bar], cfg: dict, mode: str) -> Trade:
    """진입 1분봉부터 세션 끝까지 청산을 돌림. mode = "live"(판정, 틱 경로) | "designed"(민감도, 종가 + 봉 고저)."""

    ladder_policy = LADDERS.get(cfg["common_exit_policy"])
    st = PositionState.open(entry, ladder_policy)
    bars = sorted(session_bars, key=lambda b: b.start)
    closes, vwap_num, vwap_den = [], 0.0, 0.0
    prev_extra = None
    held = [b for b in bars if b.start >= entry.time]
    for b in bars:
        closes.append(b.close)
        vwap_num += (b.high + b.low + b.close) / 3 * b.volume
        vwap_den += b.volume
        ema9 = ema(closes, 9)
        vwap = vwap_num / vwap_den if vwap_den > 0 else None
        if b.start < entry.time:
            prev_extra = {"close": b.close, "ema9": ema9, "vwap": vwap}
            continue
        policy = effective_policy(cfg, b.start)
        if mode == "live":
            for p in intrabar_path(b.open, b.high, b.low, b.close):
                on_tick(st, p, b.start, cfg, policy, ladder_policy)
                if st.qty == 0:
                    break
        else:
            extra = {"bar_high": b.high, "bar_low": b.low}
            if prev_extra is not None and vwap is not None and prev_extra["vwap"] is not None:
                extra.update(vwap_value=vwap, prev_vwap_value=prev_extra["vwap"], prev_close=prev_extra["close"],
                             ema9_value=ema9, prev_ema9_value=prev_extra["ema9"])
            on_tick(st, b.close, b.start, cfg, policy, ladder_policy, extra)
        prev_extra = {"close": b.close, "ema9": ema9, "vwap": vwap}
        if st.qty == 0:
            break
    if st.qty > 0 and held:
        # 장 마감 뒤 보유 모델링 없음 — 마지막 정규장 봉 종가 bid 로 청산(design §3)
        _sell(st, st.qty, held[-1].close, "EOD", "eod_last_regular_bar", held[-1].start)
    net = sum(net_pnl(entry.price, f.price, f.qty, MARKET) for f in st.fills)
    return Trade(entry, st.fills, net / (entry.price * Q0) * 100, risk_pct(entry, cfg))

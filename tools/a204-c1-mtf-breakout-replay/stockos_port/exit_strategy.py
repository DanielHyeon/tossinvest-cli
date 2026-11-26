# stockos-port: source=packages/trading/stockos_trading/exit_strategy.py commit=fcad869ecc531a6de80777f59b6fb842a2891101 — verbatim copy, do not edit
"""Exit-strategy orchestrator.

Plan reference: ``docs/implementation/006-plan-exit-ladder-and-session-hours.md``
§4 (PR-L1 ladder), §5 (PR-L2 limit-up close hold), §8 (decision priority).

Public surface kept stable for legacy callers — ``evaluate_exit(context,
policy)`` still returns the same ``ExitDecision`` shape and signs of the
historical 2-tier TP path. PR-L1 adds opt-in ladder evaluation when
``ladder_state`` + ``ladder_policy`` are both supplied; PR-L2 adds the
limit-up close defer post-filter that converts a profit-side decision
into ``HOLD`` when the symbol is locked at the upper limit on the
closing auction. The pipeline below is broken into helpers (each ≤50
lines) and ``evaluate_exit`` itself is a 5-line orchestrator.
"""

from __future__ import annotations

from dataclasses import dataclass
from datetime import date, datetime, timezone
from enum import Enum
from typing import Callable

from .costs import Market, break_even_sell_price, net_pnl
from .exit.stop_controls import HARD_STOP_LOSS_PCT_MIN
from .indicators import crossed_below
from .profit_ladder import (
    FillModel,
    LadderActionKind,
    LadderContext,
    LadderState,
    LimitUpHoldState,
    ProfitLadderPolicy,
    evaluate_ladder,
)

EXIT_REASON_TRENDLINE_ZONE_BODY_CLOSE_AGAINST = "trendline_zone_body_close_against"
EXIT_REASON_TRUE_CVD_DIVERGENCE_AGAINST = "true_cvd_divergence_against"


class ExitAction(str, Enum):
    HOLD = "HOLD"
    STOP_LOSS = "STOP_LOSS"
    MOVE_STOP_TO_BREAKEVEN = "MOVE_STOP_TO_BREAKEVEN"
    TAKE_PARTIAL_PROFIT = "TAKE_PARTIAL_PROFIT"
    TRAIL_PROFIT = "TRAIL_PROFIT"
    TAKE_FULL_PROFIT = "TAKE_FULL_PROFIT"
    TIME_EXIT = "TIME_EXIT"
    RAISE_PROFIT_LOCK = "RAISE_PROFIT_LOCK"
    TRAIL_EMA9_BREAK = "TRAIL_EMA9_BREAK"
    TRAIL_VWAP_CROSS = "TRAIL_VWAP_CROSS"
    # PR-L1 ladder actions (plan §4.4).
    LADDER_TAKE_PROFIT = "LADDER_TAKE_PROFIT"
    LADDER_PARTIAL = "LADDER_PARTIAL"
    LADDER_HOLD_STOP_PROMOTED = "LADDER_HOLD_STOP_PROMOTED"
    STOP_LOSS_LADDER = "STOP_LOSS_LADDER"
    # a029 — explicit end-of-day full flatten for intraday holdings.
    EOD_FLATTEN = "EOD_FLATTEN"
    # Plan 041 ① — 스카우트 time stop 전용. 일반 TIME_EXIT와 구분해
    # 15초 reprice 경로를 활성화한다 (price-drop gate 우회).
    TIME_STOP_LIMIT = "time_stop_limit"


# Set of actions that should never be routed as broker orders. Used by
# the pipeline to short-circuit before calling KIS.
_NON_ORDERABLE_ACTIONS: frozenset[ExitAction] = frozenset(
    {
        ExitAction.HOLD,
        ExitAction.LADDER_HOLD_STOP_PROMOTED,
        ExitAction.MOVE_STOP_TO_BREAKEVEN,
    }
)


def is_orderable_exit_action(action: ExitAction) -> bool:
    """True when the action represents a real broker order request.

    State-only / hold-only decisions (HOLD, LADDER_HOLD_STOP_PROMOTED,
    MOVE_STOP_TO_BREAKEVEN) emit audit events but never call KIS.
    Pipeline guards on this before invoking ``place_order``.
    """

    return action not in _NON_ORDERABLE_ACTIONS


# Set of profit-side actions that ARE eligible for limit-up close defer.
# STOP_LOSS, STOP_LOSS_LADDER, RISK_EXIT, MANUAL_EXIT, and loss-taking
# TIME_EXIT are NEVER deferred (plan §5.3 / CONFIRM-L2).
_LIMIT_UP_DEFERRABLE_ACTIONS: frozenset[ExitAction] = frozenset(
    {
        ExitAction.TAKE_PARTIAL_PROFIT,
        ExitAction.TAKE_FULL_PROFIT,
        ExitAction.TRAIL_PROFIT,
        ExitAction.RAISE_PROFIT_LOCK,
        ExitAction.LADDER_TAKE_PROFIT,
        ExitAction.LADDER_PARTIAL,
    }
)


@dataclass(frozen=True)
class ExitPolicy:
    hard_stop_loss_pct: float = -1.2
    weak_trend_stop_loss_pct: float = -0.8
    breakeven_after_pct: float = 0.7
    first_take_profit_pct: float = 1.2
    second_take_profit_pct: float = 2.2
    trailing_stop_gap_pct: float = 0.8
    max_hold_minutes: int = 90
    partial_exit_ratio: float = 0.5
    adaptive: bool = True
    rapid_profit_lock_floor_pct: float = 0.45
    rapid_profit_lock_capture_ratio: float = 0.55
    rapid_profit_lock_base_capture_ratio: float = 0.60
    rapid_profit_lock_strong_capture_ratio: float = 0.65
    rapid_profit_lock_tight_capture_ratio: float = 0.70
    early_surge_take_profit_pct: float = 0.8
    early_surge_partial_exit_ratio: float = 0.5
    # 진입 직후 손절 유예(grace period). 체결 직후 스프레드/노이즈로
    # 가격이 진입가 바로 아래로 잠깐 내려가도 손절을 미뤄 "매수 직후 즉시
    # 손절 매도" 손실 루프를 막는다. ``hard_stop_grace_minutes`` 동안에는
    # ``hard_stop_grace_catastrophe_pct`` 보다 완만한 하락의 손절을 유예한다.
    # 기본 grace_minutes 0 = 유예 없음(기존 동작 보존). 라이브 설정에서 켠다.
    # catastrophe 기본 -1.5%는 라이브/런타임/마이그레이션 default와 일치시킨다
    # (grace를 켜면서 catastrophe를 명시 안 해도 시스템 표준값이 적용되도록).
    hard_stop_grace_minutes: int = 0
    hard_stop_grace_catastrophe_pct: float = -1.5
    # 익절(breakeven) 보호 arm/floor의 ATR 스케일 배수. 0.0 = 끔(기존 고정
    # 임계 동작 100% 보존, 롤백 기본값). 켜면 breakeven_atr_pct 충전 시에만
    # MFE(high-water) >= k×ATR에서 floor를 high-water − k×ATR로 올린다.
    # 라이브 기본값은 _LIVE_DEFAULT_BREAKEVEN_ATR_MULTIPLIER(1.0)로 켠다.
    breakeven_atr_multiplier: float = 0.0
    # a041 b1 — breakeven 트레일 arm gap%의 하한(min-gap). 0.0 = 끔(비트동일, 롤백
    # 기본). >0이면 유효 arm gap = max(min_gap, k×min(진입ATR%, cap))로, 작은/부실
    # 진입ATR%가 만드는 hair-trigger breakeven(winner 조기컷)을 막는다. floor는 여전히
    # max(break_even, ·)라 진입가 아래로 안 감(단방향, 손실 보호 불변).
    breakeven_atr_min_gap_pct: float = 0.0
    # 손절(기술적/플로어) 발동 확인 계층(P1-1 RC-K3). support 하회가
    # stop_confirm_seconds 이상 지속될 때만 발동(노이즈 단일틱 청산 차단).
    # disaster(catastrophe)는 즉시. 0.0 = 끔(단일틱, 기존 동작 보존, 롤백 기본).
    stop_confirm_seconds: float = 0.0
    # 손절 거리 변동성 스케일 배수 k — 유효 손절 = max(hard_stop, k×ATR%).
    # 0.0 = 끔(고정값). C-① 구조손절 비활성일 때만 폴백 적용(design D7 H5).
    stop_atr_multiplier: float = 0.0
    # 058 P1-1 §5 RC-K5 — ExitContext.atr_pct 를 진입 시점 실제 ATR%(entry_atr_pct)
    # 로 배선할지 여부(게이트). False = 끔(0.8 고정, 라이브 데드코드 유지=비트동일).
    # 켜면 trailing/adapt(volatility_buffer)가 실데이터로 작동. breakeven 격리 불변.
    # 충전율 ~11%(US 0%)라 첫 배포 off, _build_exit_context 에서만 소비(H3 단일지점).
    use_entry_atr_pct: bool = False
    # 058 P1-1 §6 D4 — 시그널 take_profit_1 소비 게이트. exit이 시그널 stop_price는
    # 소비하면서 take_profit_1은 무시(더 먼 정책 익절 사용)하던 비대칭을 해소한다.
    # 켜면 시그널 첫 익절이 정책 first_take_profit_pct 보다 **가까울 때만** 그것을
    # 첫 부분익절 임계로 소비(단방향 min — 익절을 더 멀게 만들지 않음, design D6).
    # False = 끔(정책 익절 그대로, 비트동일, 롤백 기본). consume된 값은 adapt
    # 재구성을 거쳐 _first_tp_candidate 가 active 로 읽으므로 3분기 전파 필수(H4).
    consume_signal_tp: bool = False


RAPID_LOCK_PROFIT_TRIGGER_PCT = 3.0
RAPID_LOCK_STRONG_PROFIT_PCT = 5.0
RAPID_LOCK_SURGE_DAY_CHANGE_PCT = 5.0
RAPID_LOCK_BASE_CAPTURE_RATIO = 0.60
RAPID_LOCK_STRONG_CAPTURE_RATIO = 0.65
RAPID_LOCK_TIGHT_CAPTURE_RATIO = 0.70

# 익절 breakeven arm 임계의 변동성 스케일 상한(%). ATR 기여를 이 값으로 cap하여
# 비정상 ATR 스파이크가 breakeven 무장을 무한정 늦추지 못하게 한다.
# adapt_exit_policy 의 volatility_buffer cap(min(atr_pct, 3.0))과 일치시킨다.
BREAKEVEN_ATR_CAP = 3.0

# a013 — 런타임 hard_stop ATR 스케일(k×진입ATR%)의 진입 ATR% 상한(%). breakeven(위
# BREAKEVEN_ATR_CAP)·adapt volatility_buffer(min(atr_pct,3.0))과 동일한 3.0 으로 맞춰
# 비정상 ATR 스파이크(갭/상한가, inf)가 손절을 무한정 넓히지 못하게 한다. cap=3.0,
# k 최대 3.0 → 유효 hard 최대 −9%로 유계(단방향 안전, 형제 소비자와 컨벤션 일치).
STOP_ATR_CAP = 3.0


def compute_effective_hard_stop_pct(
    hard_stop_loss_pct: float,
    entry_atr_pct: float | None,
    stop_atr_multiplier: float,
    *,
    cap: float = STOP_ATR_CAP,
) -> float:
    """a013 — 유효 hard 손절% = −max(|hard|, k×min(진입ATR%, cap)).

    ``_stop_candidate``(라이브)와 a013 shadow 관측기가 **공유하는 단일 출처**(formula
    드리프트 0). ``min`` 이므로 **넓히기만** 한다(좁히지 않음 = 손실 보호 약화 금지,
    단방향 안전). k<=0(라이브 기본 0.0) 또는 진입 ATR None/<=0 이면 기존 hard 그대로
    (fail-safe + flag OFF 비트동일). cap 으로 갭/상한가/inf 비정상 ATR 의 무한 확대 방지.
    """

    if stop_atr_multiplier > 0 and entry_atr_pct is not None and entry_atr_pct > 0:
        capped_atr_pct = min(entry_atr_pct, cap)
        atr_scaled_stop_pct = -(stop_atr_multiplier * capped_atr_pct)
        return min(hard_stop_loss_pct, atr_scaled_stop_pct)
    return hard_stop_loss_pct


def compute_effective_breakeven_arm_pct(
    breakeven_atr_multiplier: float,
    breakeven_atr_pct: float | None,
    *,
    cap: float = BREAKEVEN_ATR_CAP,
    min_gap: float = 0.0,
) -> float | None:
    """a041 b1 — breakeven 트레일 arm gap% = max(min_gap, k×min(진입ATR%, cap)).

    ``_breakeven_candidate``(라이브 ATR 경로)의 arm_pct **단일 출처**. ``max`` 하한이라
    **넓히기만** 한다(단방향: arm 지연·floor 하향, 단 floor는 호출부 max(break_even, ·)로
    break_even 이상 보장 = 손실 보호 불변). ``min_gap=0`` → k×min(atr,cap) 그대로
    (**비트동일**, 롤백 기본). multiplier≤0 또는 진입ATR None/≤0 이면 None(ATR 경로 비활성 =
    기존 고정 임계 경로 보존, fail-closed). cap 으로 비정상 ATR 스파이크(갭/상한가/inf)의
    무한 확대 방지(형제 ``BREAKEVEN_ATR_CAP``·a013 ``compute_effective_hard_stop_pct`` 컨벤션 일치).
    """

    if (
        breakeven_atr_multiplier <= 0
        or breakeven_atr_pct is None
        or breakeven_atr_pct <= 0
    ):
        return None
    scaled = breakeven_atr_multiplier * min(breakeven_atr_pct, cap)
    return max(scaled, min_gap)


@dataclass(frozen=True)
class ExitContext:
    entry_price: float
    current_price: float
    high_since_entry: float
    quantity: int
    market: Market
    minutes_held: int
    # Caller-scoped control for the generic max-hold TIME_EXIT branch.
    # Default true preserves every existing direct/shared evaluator caller.
    # The quote-monitor context builder explicitly disables this branch while
    # leaving EOD/onset/protective exits outside this control.
    max_hold_exit_enabled: bool = True
    bollinger_middle: float | None = None
    bollinger_upper: float | None = None
    support_price: float | None = None
    # 기술적 손절(support/bollinger) 조건이 연속 유지된 시간(초). caller가
    # quote.timestamp 기준으로 추적해 주입(design D7 M3). None = 미제공
    # (기존 단일틱 동작 = 확인 계층 비활성과 동일).
    support_breach_elapsed_seconds: float | None = None
    cci: float | None = None
    trend_slope: float = 0.0
    resistance_slope: float = 0.0
    atr_pct: float = 0.8
    # breakeven arm 전용 실제 ATR%(진입 시점, position_state.entry_atr_pct 출처).
    # ``atr_pct``(trailing/adapt 공유, 라이브에서 미충전 0.8 고정)와 분리한 격리
    # 필드 — active breakeven ATR path는 이 값이 없으면 고정 0.7%로 폴백하지
    # 않고 비활성화한다. multiplier off path만 기존 고정 임계를 보존한다.
    breakeven_atr_pct: float | None = None
    # 058 P1-1 §6 D4 — 시그널 take_profit_1 의 진입가 대비 거리(%). caller가
    # feature_signal.take_profit_1 있으면 (tp1-entry)/entry*100 로 주입(없으면
    # None). ``ExitPolicy.consume_signal_tp`` 게이트 on일 때만 _first_tp_candidate
    # 가 정책 first_take_profit_pct 와 min 비교에 소비(breakeven_atr_pct 격리 선례).
    signal_take_profit_1_pct: float | None = None
    rvol: float = 1.0
    day_change_pct: float = 0.0
    vwap_value: float | None = None
    prev_vwap_value: float | None = None
    prev_close: float | None = None
    ema9_value: float | None = None
    prev_ema9_value: float | None = None
    # PR-L2 — limit-up close hold caller-opt-in fields.
    is_limit_up: bool = False
    limit_up_locked: bool = False
    is_session_closing_bar: bool = False
    close_auction_phase: bool = False
    session_date: date | None = None
    moment: datetime | None = None  # required by the ladder branch when used
    bar_high: float | None = None
    bar_low: float | None = None
    ladder_fill_model: FillModel = FillModel.STOP_FIRST
    early_surge_entry: bool = False
    partial_tp_done: bool = False
    position_direction: str = "long"
    zone_body_close_above_upper: bool = False
    zone_body_close_below_lower: bool = False
    cvd_exit_enforce_enabled: bool = False
    order_flow_status: str | None = None
    cvd_falling: bool = False
    cvd_rising: bool = False
    cvd_divergence_with_price: bool = False

    @property
    def gross_return_pct(self) -> float:
        return ((self.current_price - self.entry_price) / self.entry_price) * 100


@dataclass(frozen=True)
class ExitDecision:
    action: ExitAction
    quantity_to_sell: int
    stop_price: float | None
    reason: str
    net_pnl_krw: float
    break_even_price: float
    applied_policy: ExitPolicy
    # PR-L1: when the ladder branch fires, expose the rung index for
    # downstream audit/idempotency. ``None`` when the legacy 2-tier path
    # produced the decision.
    ladder_rung_index: int | None = None
    # PR-L1: post-decision ladder state (decision-time mutation only —
    # partial_taken_ratio_total / completed update on fill via
    # execution_sync).
    next_ladder_state: LadderState | None = None
    # PR-L2: when defer fires, the original action that was overridden.
    deferred_from_action: ExitAction | None = None


def adapt_exit_policy(
    context: ExitContext, policy: ExitPolicy | None = None
) -> ExitPolicy:
    active = policy or ExitPolicy()
    if not active.adaptive:
        return active

    strong_trend = (
        context.trend_slope > 0
        and context.resistance_slope >= 0
        and context.rvol >= 2.0
    )
    weak_trend = context.trend_slope < 0 or context.rvol < 1.0
    volatility_buffer = max(0.0, min(context.atr_pct, 3.0)) * 0.25

    if strong_trend:
        return ExitPolicy(
            hard_stop_loss_pct=max(
                HARD_STOP_LOSS_PCT_MIN,
                min(-0.6, active.hard_stop_loss_pct - volatility_buffer),
            ),
            weak_trend_stop_loss_pct=active.weak_trend_stop_loss_pct,
            breakeven_after_pct=active.breakeven_after_pct,
            first_take_profit_pct=active.first_take_profit_pct + 0.3,
            second_take_profit_pct=active.second_take_profit_pct + 0.8,
            trailing_stop_gap_pct=active.trailing_stop_gap_pct
            + 0.4
            + volatility_buffer,
            max_hold_minutes=active.max_hold_minutes + 30,
            partial_exit_ratio=max(0.3, active.partial_exit_ratio - 0.2),
            adaptive=active.adaptive,
            rapid_profit_lock_floor_pct=active.rapid_profit_lock_floor_pct,
            rapid_profit_lock_capture_ratio=active.rapid_profit_lock_capture_ratio,
            rapid_profit_lock_base_capture_ratio=(
                active.rapid_profit_lock_base_capture_ratio
            ),
            rapid_profit_lock_strong_capture_ratio=(
                active.rapid_profit_lock_strong_capture_ratio
            ),
            rapid_profit_lock_tight_capture_ratio=(
                active.rapid_profit_lock_tight_capture_ratio
            ),
            early_surge_take_profit_pct=active.early_surge_take_profit_pct,
            early_surge_partial_exit_ratio=active.early_surge_partial_exit_ratio,
            hard_stop_grace_minutes=active.hard_stop_grace_minutes,
            hard_stop_grace_catastrophe_pct=active.hard_stop_grace_catastrophe_pct,
            breakeven_atr_multiplier=active.breakeven_atr_multiplier,
            breakeven_atr_min_gap_pct=active.breakeven_atr_min_gap_pct,
            stop_confirm_seconds=active.stop_confirm_seconds,
            stop_atr_multiplier=active.stop_atr_multiplier,
            consume_signal_tp=active.consume_signal_tp,
        )

    if weak_trend:
        return ExitPolicy(
            hard_stop_loss_pct=max(
                active.weak_trend_stop_loss_pct,
                active.hard_stop_loss_pct + 0.4,
            ),
            weak_trend_stop_loss_pct=active.weak_trend_stop_loss_pct,
            breakeven_after_pct=max(0.4, active.breakeven_after_pct - 0.2),
            first_take_profit_pct=max(0.8, active.first_take_profit_pct - 0.2),
            second_take_profit_pct=max(1.5, active.second_take_profit_pct - 0.4),
            trailing_stop_gap_pct=max(0.4, active.trailing_stop_gap_pct - 0.2),
            max_hold_minutes=max(30, active.max_hold_minutes - 30),
            partial_exit_ratio=min(0.7, active.partial_exit_ratio + 0.2),
            adaptive=active.adaptive,
            rapid_profit_lock_floor_pct=active.rapid_profit_lock_floor_pct,
            rapid_profit_lock_capture_ratio=active.rapid_profit_lock_capture_ratio,
            rapid_profit_lock_base_capture_ratio=(
                active.rapid_profit_lock_base_capture_ratio
            ),
            rapid_profit_lock_strong_capture_ratio=(
                active.rapid_profit_lock_strong_capture_ratio
            ),
            rapid_profit_lock_tight_capture_ratio=(
                active.rapid_profit_lock_tight_capture_ratio
            ),
            early_surge_take_profit_pct=active.early_surge_take_profit_pct,
            early_surge_partial_exit_ratio=active.early_surge_partial_exit_ratio,
            hard_stop_grace_minutes=active.hard_stop_grace_minutes,
            hard_stop_grace_catastrophe_pct=active.hard_stop_grace_catastrophe_pct,
            breakeven_atr_multiplier=active.breakeven_atr_multiplier,
            breakeven_atr_min_gap_pct=active.breakeven_atr_min_gap_pct,
            stop_confirm_seconds=active.stop_confirm_seconds,
            stop_atr_multiplier=active.stop_atr_multiplier,
            consume_signal_tp=active.consume_signal_tp,
        )

    return ExitPolicy(
        hard_stop_loss_pct=max(
            HARD_STOP_LOSS_PCT_MIN,
            active.hard_stop_loss_pct - volatility_buffer,
        ),
        weak_trend_stop_loss_pct=active.weak_trend_stop_loss_pct,
        breakeven_after_pct=active.breakeven_after_pct,
        first_take_profit_pct=active.first_take_profit_pct,
        second_take_profit_pct=active.second_take_profit_pct,
        trailing_stop_gap_pct=active.trailing_stop_gap_pct + volatility_buffer,
        max_hold_minutes=active.max_hold_minutes,
        partial_exit_ratio=active.partial_exit_ratio,
        adaptive=active.adaptive,
        rapid_profit_lock_floor_pct=active.rapid_profit_lock_floor_pct,
        rapid_profit_lock_capture_ratio=active.rapid_profit_lock_capture_ratio,
        rapid_profit_lock_base_capture_ratio=(
            active.rapid_profit_lock_base_capture_ratio
        ),
        rapid_profit_lock_strong_capture_ratio=(
            active.rapid_profit_lock_strong_capture_ratio
        ),
        rapid_profit_lock_tight_capture_ratio=(
            active.rapid_profit_lock_tight_capture_ratio
        ),
        early_surge_take_profit_pct=active.early_surge_take_profit_pct,
        early_surge_partial_exit_ratio=active.early_surge_partial_exit_ratio,
        hard_stop_grace_minutes=active.hard_stop_grace_minutes,
        hard_stop_grace_catastrophe_pct=active.hard_stop_grace_catastrophe_pct,
        breakeven_atr_multiplier=active.breakeven_atr_multiplier,
        breakeven_atr_min_gap_pct=active.breakeven_atr_min_gap_pct,
        stop_confirm_seconds=active.stop_confirm_seconds,
        stop_atr_multiplier=active.stop_atr_multiplier,
        consume_signal_tp=active.consume_signal_tp,
    )


def rapid_profit_lock_price(context: ExitContext, policy: ExitPolicy) -> float | None:
    capture_ratio = _rapid_profit_lock_capture_ratio(context, policy)
    rapid_up = capture_ratio is not None
    if not rapid_up or context.high_since_entry <= context.entry_price:
        return None
    captured_gain = (context.high_since_entry - context.entry_price) * capture_ratio
    minimum_gain = context.entry_price * (policy.rapid_profit_lock_floor_pct / 100)
    return context.entry_price + max(captured_gain, minimum_gain)


def _rapid_profit_lock_capture_ratio(
    context: ExitContext, policy: ExitPolicy
) -> float | None:
    """Return the high-water gain ratio to protect for rapid winners."""

    gross_return = context.gross_return_pct
    if (
        gross_return < RAPID_LOCK_PROFIT_TRIGGER_PCT
        and context.day_change_pct < RAPID_LOCK_SURGE_DAY_CHANGE_PCT
    ):
        return None

    base_ratio = max(
        policy.rapid_profit_lock_capture_ratio,
        policy.rapid_profit_lock_base_capture_ratio,
        RAPID_LOCK_BASE_CAPTURE_RATIO,
    )
    strong_ratio = max(
        base_ratio,
        policy.rapid_profit_lock_strong_capture_ratio,
        RAPID_LOCK_STRONG_CAPTURE_RATIO,
    )
    tight_ratio = max(
        strong_ratio,
        policy.rapid_profit_lock_tight_capture_ratio,
        RAPID_LOCK_TIGHT_CAPTURE_RATIO,
    )
    ratio = base_ratio
    if (
        gross_return >= RAPID_LOCK_STRONG_PROFIT_PCT
        or context.day_change_pct >= RAPID_LOCK_SURGE_DAY_CHANGE_PCT
    ):
        ratio = strong_ratio

    strong_trend_or_volume = (
        context.trend_slope > 0
        and context.resistance_slope >= 0
        and context.rvol >= 2.0
    )
    weak_trend_or_volume = context.trend_slope < 0 or context.rvol < 1.0
    if weak_trend_or_volume:
        return tight_ratio
    if strong_trend_or_volume:
        return strong_ratio
    return ratio


def evaluate_exit(
    context: ExitContext,
    policy: ExitPolicy | None = None,
    *,
    ladder_policy: ProfitLadderPolicy | None = None,
    ladder_state: LadderState | None = None,
    limit_up_hold_state: LimitUpHoldState | None = None,
) -> ExitDecision:
    """Run the exit pipeline.

    Order (plan §4.5.1 collect → select → defer → persist):

    1. Validate inputs.
    2. Collect candidate actions (legacy 2-tier path + optional ladder
       path; the two are mutually exclusive — passing one of
       ladder_policy/ladder_state without the other is a ``ValueError``).
    3. Select the single final action via the 7-step priority
       (``_select_final_action``).
    4. Apply the limit-up close defer post-filter to the selected
       decision (``_apply_limit_up_defer_to_selected``) — the helper
       sees only the winner, never the candidate list.
    """

    if (ladder_policy is None) ^ (ladder_state is None):
        raise ValueError(
            "ladder_policy and ladder_state must both be provided or both omitted"
        )

    active = adapt_exit_policy(context, policy)
    if context.quantity <= 0:
        raise ValueError("quantity must be positive")
    if context.entry_price <= 0 or context.current_price <= 0:
        raise ValueError("prices must be positive")

    candidates = _collect_exit_candidates(
        context=context,
        active=active,
        ladder_policy=ladder_policy,
        ladder_state=ladder_state,
    )
    selected = _select_final_action(candidates)
    return _apply_limit_up_defer_to_selected(
        selected=selected,
        context=context,
        hold_state=limit_up_hold_state,
    )


def evaluate_safety_exit(
    context: ExitContext,
    policy: ExitPolicy | None = None,
) -> ExitDecision:
    """Evaluate protective exits only.

    This fail-closed path is used when a persisted policy is corrupt or has
    been removed from the LIVE allowlist. It deliberately preserves immediate
    technical/hard/disaster stops while suppressing time, profit, trailing,
    and ladder transitions until an operator repairs the policy state.
    """

    active = adapt_exit_policy(context, policy)
    if context.quantity <= 0:
        raise ValueError("quantity must be positive")
    if context.entry_price <= 0 or context.current_price <= 0:
        raise ValueError("prices must be positive")
    pnl, break_even, return_pct = _compute_pnl_breakdown(context)
    for branch in (
        _trendline_zone_body_close_candidate,
        _true_cvd_divergence_candidate,
        _stop_candidate,
    ):
        candidate = branch(
            context,
            active,
            pnl,
            break_even,
            return_pct,
            False,
        )
        if candidate is not None:
            return candidate
    return _make_decision(
        action=ExitAction.HOLD,
        quantity_to_sell=0,
        stop_price=None,
        reason="policy_safety_only_hold",
        context=context,
        active=active,
        pnl=pnl,
        break_even=break_even,
    )


# ---- helpers ---------------------------------------------------------------


def _make_decision(
    *,
    action: ExitAction,
    quantity_to_sell: int,
    stop_price: float | None,
    reason: str,
    context: ExitContext,
    active: ExitPolicy,
    pnl: float,
    break_even: float,
    ladder_rung_index: int | None = None,
    next_ladder_state: LadderState | None = None,
) -> ExitDecision:
    return ExitDecision(
        action=action,
        quantity_to_sell=quantity_to_sell,
        stop_price=stop_price,
        reason=reason,
        net_pnl_krw=pnl,
        break_even_price=break_even,
        applied_policy=active,
        ladder_rung_index=ladder_rung_index,
        next_ladder_state=next_ladder_state,
    )


def _collect_exit_candidates(
    *,
    context: ExitContext,
    active: ExitPolicy,
    ladder_policy: ProfitLadderPolicy | None,
    ladder_state: LadderState | None,
) -> list[ExitDecision]:
    """Return all candidate decisions for this bar.

    The two profit paths (legacy 2-tier vs ladder) are mutually
    exclusive at the call site, but stop-loss / time-exit / breakeven
    / Phase G triggers all run unconditionally. The ladder branch is
    appended last so the priority selector sees both kinds of profit
    candidates side-by-side.
    """

    pnl, break_even, return_pct = _compute_pnl_breakdown(context)
    candidates: list[ExitDecision] = []

    candidates.extend(
        _collect_legacy_candidates(
            context=context,
            active=active,
            pnl=pnl,
            break_even=break_even,
            return_pct=return_pct,
            include_legacy_tp=ladder_policy is None,
            include_trailing_profit=not (
                ladder_policy is not None
                and ladder_policy.policy_id == "hybrid_50"
                and ladder_state is not None
                and not ladder_state.completed
            ),
        )
    )
    if ladder_policy is not None and ladder_state is not None:
        ladder_candidate = _evaluate_ladder_branch(
            context=context,
            ladder_policy=ladder_policy,
            ladder_state=ladder_state,
            active=active,
            pnl=pnl,
            break_even=break_even,
        )
        if ladder_candidate is not None:
            hybrid_final_handoff = (
                ladder_policy.policy_id == "hybrid_50"
                and not ladder_state.completed
                and ladder_candidate.next_ladder_state is not None
                and ladder_candidate.next_ladder_state.completed
                and ladder_candidate.action is ExitAction.LADDER_HOLD_STOP_PROMOTED
            )
            if hybrid_final_handoff:
                candidates = [
                    candidate
                    for candidate in candidates
                    if candidate.action
                    in {
                        ExitAction.STOP_LOSS,
                        ExitAction.STOP_LOSS_LADDER,
                    }
                ]
            candidates.append(ladder_candidate)
    if not candidates:
        candidates.append(
            _make_decision(
                action=ExitAction.HOLD,
                quantity_to_sell=0,
                stop_price=None,
                reason="no_exit_signal",
                context=context,
                active=active,
                pnl=pnl,
                break_even=break_even,
            )
        )
    return candidates


def _compute_pnl_breakdown(context: ExitContext) -> tuple[float, float, float]:
    break_even = break_even_sell_price(
        context.entry_price, context.quantity, context.market
    )
    pnl = net_pnl(
        context.entry_price, context.current_price, context.quantity, context.market
    )
    return pnl, break_even, context.gross_return_pct


_BranchFn = Callable[
    [ExitContext, ExitPolicy, float, float, float, bool], "ExitDecision | None"
]


def _collect_legacy_candidates(
    *,
    context: ExitContext,
    active: ExitPolicy,
    pnl: float,
    break_even: float,
    return_pct: float,
    include_legacy_tp: bool,
    include_trailing_profit: bool,
) -> list[ExitDecision]:
    """All non-ladder candidates: stops, time-exit, Phase G, legacy TP, breakeven.

    Each branch is delegated to a small helper. The first one that produces
    a candidate short-circuits the rest — preserves the strict precedence
    of the pre-decomposition implementation.
    """

    branches: tuple[_BranchFn, ...] = (
        _trendline_zone_body_close_candidate,
        _true_cvd_divergence_candidate,
        _stop_candidate,
        _time_exit_candidate,
        _early_surge_partial_tp_candidate,
        _rapid_lock_candidate,
        _second_tp_candidate,
        _phase_g_candidate_branch,
        _trail_profit_candidate,
        _first_tp_candidate,
        _breakeven_candidate,
    )
    for branch in branches:
        if branch is _trail_profit_candidate and not include_trailing_profit:
            continue
        candidate = branch(
            context, active, pnl, break_even, return_pct, include_legacy_tp
        )
        if candidate is not None:
            return [candidate]
    return []


def _trendline_zone_body_close_candidate(
    context: ExitContext,
    active: ExitPolicy,
    pnl: float,
    break_even: float,
    _return_pct: float,
    _include_tp: bool,
) -> ExitDecision | None:
    if not _is_body_close_against_position(context):
        return None
    return _make_decision(
        action=ExitAction.STOP_LOSS,
        quantity_to_sell=context.quantity,
        stop_price=context.current_price,
        reason=EXIT_REASON_TRENDLINE_ZONE_BODY_CLOSE_AGAINST,
        context=context,
        active=active,
        pnl=pnl,
        break_even=break_even,
    )


def _is_body_close_against_position(context: ExitContext) -> bool:
    direction = context.position_direction.lower()
    if direction == "long":
        return context.zone_body_close_below_lower
    if direction == "short":
        return context.zone_body_close_above_upper
    return False


def _true_cvd_divergence_candidate(
    context: ExitContext,
    active: ExitPolicy,
    pnl: float,
    break_even: float,
    _return_pct: float,
    _include_tp: bool,
) -> ExitDecision | None:
    if not context.cvd_exit_enforce_enabled:
        return None
    if context.order_flow_status != "true_cvd":
        return None
    if not context.cvd_divergence_with_price:
        return None
    direction = context.position_direction.lower()
    against_long = direction == "long" and context.cvd_falling
    against_short = direction == "short" and context.cvd_rising
    if not (against_long or against_short):
        return None
    return _make_decision(
        action=ExitAction.STOP_LOSS,
        quantity_to_sell=context.quantity,
        stop_price=context.current_price,
        reason=EXIT_REASON_TRUE_CVD_DIVERGENCE_AGAINST,
        context=context,
        active=active,
        pnl=pnl,
        break_even=break_even,
    )


def _stop_candidate(
    context: ExitContext,
    active: ExitPolicy,
    pnl: float,
    break_even: float,
    return_pct: float,
    _include_tp: bool,
) -> ExitDecision | None:
    technical_stop = (
        context.support_price is not None
        and context.current_price < context.support_price
    ) or (
        context.bollinger_middle is not None
        and context.current_price < context.bollinger_middle
        and pnl < 0
    )
    # 진입 직후 손절 유예(grace period): 체결 직후 grace 윈도우 동안에는
    # 스프레드/노이즈성 소폭 하락의 손절을 미룬다. 단 catastrophe(급락)
    # 임계 이하로 떨어지면 유예 없이 즉시 손절한다.
    in_grace = context.minutes_held < active.hard_stop_grace_minutes
    catastrophic = return_pct <= active.hard_stop_grace_catastrophe_pct
    if in_grace and not catastrophic:
        return None
    # 손절 발동 확인 계층(P1-1 RC-K3): 기술적 손절(support/bollinger)은 지속
    # 하회가 stop_confirm_seconds 이상일 때만 발동해 시초 노이즈 단일틱 청산을
    # 막는다. elapsed 미제공(None)이거나 임계 0 이면 즉시(기존 동작 보존),
    # disaster(catastrophe)는 확인 무시 즉시 발동.
    breach_confirmed = (
        context.support_breach_elapsed_seconds is None
        or context.support_breach_elapsed_seconds >= active.stop_confirm_seconds
    )
    technical_stop_armed = technical_stop and (breach_confirmed or catastrophic)
    # a013: 런타임 hard_stop 을 진입 시점 실 ATR 기준으로 넓힌다(손절을 노이즈 밴드
    # 밖에 둔다). 유효 hard = −max(|hard_stop_loss_pct|, k×진입ATR%).
    #   - k = active.stop_atr_multiplier (그동안 어디서도 소비 안 되던 死필드를 소비).
    #   - 진입 ATR% = context.breakeven_atr_pct (= position_state.entry_atr_pct, 격리
    #     충전). trailing/adapt 가 공유하는 context.atr_pct(라이브 0.8 고정)는 쓰지
    #     않는다 — blast radius 격리(breakeven_atr_pct 격리 선례와 동형).
    # min 이므로 **넓히기만** 한다(좁히지 않음 = 손실 보호 약화 금지, 단방향 안전).
    # k<=0(라이브 기본 0.0) 또는 진입 ATR None/<=0 이면 기존 hard 그대로
    # (fail-safe + flag OFF 비트동일). formula 는 a013 shadow 관측기와 공유하는 순수
    # ``compute_effective_hard_stop_pct`` 단일 출처(드리프트 0, 동작 보존 리팩터).
    effective_hard_stop_pct = compute_effective_hard_stop_pct(
        active.hard_stop_loss_pct,
        context.breakeven_atr_pct,
        active.stop_atr_multiplier,
    )
    hard_stop_fires = return_pct <= effective_hard_stop_pct
    # grace 윈도우 내 catastrophe는 넓힌 hard가 지연시키지 못한다.
    # after-grace ATR floor 의미는 a013 계약대로 유지하고, 적응형 hard는
    # adapt_exit_policy에서 A170 절대 하한(-5%)보다 깊어지지 않도록 cap한다.
    if in_grace and catastrophic:
        hard_stop_fires = True
    if hard_stop_fires or technical_stop_armed:
        return _make_decision(
            action=ExitAction.STOP_LOSS,
            quantity_to_sell=context.quantity,
            stop_price=context.current_price,
            reason="hard_or_technical_stop",
            context=context,
            active=active,
            pnl=pnl,
            break_even=break_even,
        )
    return None


def _time_exit_candidate(
    context: ExitContext,
    active: ExitPolicy,
    pnl: float,
    break_even: float,
    return_pct: float,
    _include_tp: bool,
) -> ExitDecision | None:
    if not context.max_hold_exit_enabled:
        return None
    target_reached = return_pct >= active.first_take_profit_pct
    if context.minutes_held >= active.max_hold_minutes and not target_reached:
        return _make_decision(
            action=ExitAction.TIME_EXIT,
            quantity_to_sell=context.quantity,
            stop_price=context.current_price,
            reason="time_exit_max_hold",
            context=context,
            active=active,
            pnl=pnl,
            break_even=break_even,
        )
    return None


def _early_surge_partial_tp_candidate(
    context: ExitContext,
    active: ExitPolicy,
    pnl: float,
    break_even: float,
    return_pct: float,
    include_tp: bool,
) -> ExitDecision | None:
    if (
        not include_tp
        or not context.early_surge_entry
        or context.partial_tp_done
        or return_pct < active.early_surge_take_profit_pct
    ):
        return None
    qty = max(1, int(context.quantity * active.early_surge_partial_exit_ratio))
    qty = min(qty, context.quantity)
    rapid_lock = rapid_profit_lock_price(context, active)
    stop_price = max(value for value in (rapid_lock, break_even) if value is not None)
    return _make_decision(
        action=ExitAction.TAKE_PARTIAL_PROFIT,
        quantity_to_sell=qty,
        stop_price=stop_price,
        reason="early_surge_fast_partial_profit",
        context=context,
        active=active,
        pnl=pnl,
        break_even=break_even,
    )


def _rapid_lock_candidate(
    context: ExitContext,
    active: ExitPolicy,
    pnl: float,
    break_even: float,
    return_pct: float,
    _include_tp: bool,
) -> ExitDecision | None:
    # A breached high-water trailing stop is an actionable SELL. Do not let a
    # state-only rapid-lock raise hide that pullback, especially in ladder mode
    # where static legacy TP is intentionally disabled.
    if (
        context.current_price < _trailing_stop_price(context, active)
        and return_pct >= active.first_take_profit_pct
    ):
        return None
    rapid_lock = rapid_profit_lock_price(context, active)
    if rapid_lock is not None and context.current_price > rapid_lock:
        return _make_decision(
            action=ExitAction.RAISE_PROFIT_LOCK,
            quantity_to_sell=0,
            stop_price=max(rapid_lock, break_even),
            reason="rapid_up_profit_lock_raised",
            context=context,
            active=active,
            pnl=pnl,
            break_even=break_even,
        )
    return None


def _second_tp_candidate(
    context: ExitContext,
    active: ExitPolicy,
    pnl: float,
    break_even: float,
    return_pct: float,
    include_tp: bool,
) -> ExitDecision | None:
    if include_tp and return_pct >= active.second_take_profit_pct:
        return _make_decision(
            action=ExitAction.TAKE_FULL_PROFIT,
            quantity_to_sell=context.quantity,
            stop_price=None,
            reason="second_profit_target",
            context=context,
            active=active,
            pnl=pnl,
            break_even=break_even,
        )
    return None


def _phase_g_candidate_branch(
    context: ExitContext,
    active: ExitPolicy,
    pnl: float,
    break_even: float,
    _return_pct: float,
    _include_tp: bool,
) -> ExitDecision | None:
    return _phase_g_candidate(context, active, pnl, break_even)


def _trail_profit_candidate(
    context: ExitContext,
    active: ExitPolicy,
    pnl: float,
    break_even: float,
    return_pct: float,
    include_tp: bool,
) -> ExitDecision | None:
    trailing_stop = _trailing_stop_price(context, active)
    if (
        context.current_price < trailing_stop
        and return_pct >= active.first_take_profit_pct
    ):
        return _make_decision(
            action=ExitAction.TRAIL_PROFIT,
            quantity_to_sell=context.quantity,
            stop_price=trailing_stop,
            reason="trailing_stop_locked_profit",
            context=context,
            active=active,
            pnl=pnl,
            break_even=break_even,
        )
    return None


def _trailing_stop_price(context: ExitContext, active: ExitPolicy) -> float:
    # Plan 041 ③ — adaptive=False 이고 atr_pct > 0 일 때만 ATR 동적 갭 적용.
    # adaptive=True 인 경우 adapt_exit_policy() 가 이미 trailing_stop_gap_pct
    # 에 volatility_buffer(atr_pct × 0.25)를 반영했으므로 이중 적용 방지.
    if not active.adaptive and context.atr_pct > 0:
        gap_pct = max(active.trailing_stop_gap_pct, 1.5 * context.atr_pct)
    else:
        gap_pct = active.trailing_stop_gap_pct
    return context.high_since_entry * (1 - gap_pct / 100)


def _effective_first_tp_pct(context: ExitContext, active: ExitPolicy) -> float:
    """첫 익절 임계 — D4 시그널 take_profit_1 소비(단방향) 반영.

    게이트(consume_signal_tp) on이고 시그널 첫 익절 거리가 양수이며 정책
    first_take_profit_pct 보다 **가까우면** 그것을 쓴다(min — 익절을 더 멀게
    만들지 않음, design D6). off/미제공/더 멀면 정책값 그대로(비트동일).
    """
    base = active.first_take_profit_pct
    signal_pct = context.signal_take_profit_1_pct
    if active.consume_signal_tp and signal_pct is not None and signal_pct > 0:
        return min(base, signal_pct)
    return base


def _first_tp_candidate(
    context: ExitContext,
    active: ExitPolicy,
    pnl: float,
    break_even: float,
    return_pct: float,
    include_tp: bool,
) -> ExitDecision | None:
    # WHY: gate on partial_tp_done so the legacy 2-tier branch stops re-firing
    # once the first partial has been submitted. Without this guard, the same
    # price above first_take_profit_pct re-triggers TAKE_PARTIAL_PROFIT every
    # tick (042700 2026-05-21 — five sequential first_profit_target SELLs
    # liquidated an adopted 10-share position in five seconds).
    if (
        not include_tp
        or context.partial_tp_done
        or return_pct < _effective_first_tp_pct(context, active)
    ):
        return None
    qty = max(1, int(context.quantity * active.partial_exit_ratio))
    return _make_decision(
        action=ExitAction.TAKE_PARTIAL_PROFIT,
        quantity_to_sell=qty,
        stop_price=None,
        reason="first_profit_target",
        context=context,
        active=active,
        pnl=pnl,
        break_even=break_even,
    )


def _breakeven_candidate(
    context: ExitContext,
    active: ExitPolicy,
    pnl: float,
    break_even: float,
    return_pct: float,
    _include_tp: bool,
) -> ExitDecision | None:
    if active.breakeven_atr_multiplier <= 0:
        if (
            return_pct >= active.breakeven_after_pct
            and context.current_price > break_even
        ):
            return _make_decision(
                action=ExitAction.MOVE_STOP_TO_BREAKEVEN,
                quantity_to_sell=0,
                stop_price=break_even,
                reason="protect_from_round_trip_loss",
                context=context,
                active=active,
                pnl=pnl,
                break_even=break_even,
            )
        return None

    # a041 b1 — arm gap을 단일 출처 compute_effective_breakeven_arm_pct 로 계산(min_gap
    # 하한 적용). None = ATR 경로 비활성(multiplier≤0 또는 진입ATR None/≤0), 기존 guard 보존.
    arm_pct = compute_effective_breakeven_arm_pct(
        active.breakeven_atr_multiplier,
        context.breakeven_atr_pct,
        min_gap=active.breakeven_atr_min_gap_pct,
    )
    if arm_pct is None:
        return None

    high_water = max(context.high_since_entry, context.current_price)
    mfe_pct = ((high_water - context.entry_price) / context.entry_price) * 100
    if mfe_pct < arm_pct:
        return None

    atr_floor = high_water - (context.entry_price * arm_pct / 100)
    stop_floor = max(break_even, atr_floor)
    if context.current_price > stop_floor:
        return _make_decision(
            action=ExitAction.MOVE_STOP_TO_BREAKEVEN,
            quantity_to_sell=0,
            stop_price=stop_floor,
            reason="protect_from_round_trip_loss",
            context=context,
            active=active,
            pnl=pnl,
            break_even=break_even,
        )
    return None


def _phase_g_candidate(
    context: ExitContext, active: ExitPolicy, pnl: float, break_even: float
) -> ExitDecision | None:
    if not _has_phase_g_inputs(context):
        return None
    if crossed_below(
        context.prev_close,  # type: ignore[arg-type]
        context.current_price,
        context.prev_ema9_value,  # type: ignore[arg-type]
        context.ema9_value,  # type: ignore[arg-type]
    ):
        return _make_decision(
            action=ExitAction.TRAIL_EMA9_BREAK,
            quantity_to_sell=context.quantity,
            stop_price=context.current_price,
            reason="ema9_close_break_down",
            context=context,
            active=active,
            pnl=pnl,
            break_even=break_even,
        )
    if crossed_below(
        context.prev_close,  # type: ignore[arg-type]
        context.current_price,
        context.prev_vwap_value,  # type: ignore[arg-type]
        context.vwap_value,  # type: ignore[arg-type]
    ):
        return _make_decision(
            action=ExitAction.TRAIL_VWAP_CROSS,
            quantity_to_sell=context.quantity,
            stop_price=context.current_price,
            reason="vwap_close_cross_down",
            context=context,
            active=active,
            pnl=pnl,
            break_even=break_even,
        )
    return None


def _evaluate_ladder_branch(
    *,
    context: ExitContext,
    ladder_policy: ProfitLadderPolicy,
    ladder_state: LadderState,
    active: ExitPolicy,
    pnl: float,
    break_even: float,
) -> ExitDecision | None:
    """Wrap ``profit_ladder.evaluate_ladder`` into an ``ExitDecision`` candidate.

    Returns ``None`` when the ladder produces no transition. Pending
    order suppression (plan §4.5.2.1) lets rung promotion happen but
    swallows the order action so a duplicate broker submit cannot occur
    — the next bar (after the pending fill arrives) will re-evaluate.
    """

    moment = context.moment or datetime.now(timezone.utc)
    transition = evaluate_ladder(
        LadderContext(
            entry_price=context.entry_price,
            current_price=context.current_price,
            moment=moment,
            high=context.bar_high,
            low=context.bar_low,
            fill_model=context.ladder_fill_model,
        ),
        ladder_policy,
        ladder_state,
    )
    if transition.action is None:
        return None

    action_kind = transition.action
    suppress_for_pending = ladder_state.has_pending_order and action_kind in {
        LadderActionKind.LADDER_PARTIAL,
        LadderActionKind.LADDER_TAKE_PROFIT,
    }
    if suppress_for_pending:
        # Promote stop / state but emit no order; signal to caller via
        # LADDER_HOLD_STOP_PROMOTED so audit gets `exit.ladder_blocked_by_pending_order`.
        return _make_decision(
            action=ExitAction.LADDER_HOLD_STOP_PROMOTED,
            quantity_to_sell=0,
            stop_price=None,
            reason="ladder_blocked_by_pending_order",
            context=context,
            active=active,
            pnl=pnl,
            break_even=break_even,
            ladder_rung_index=transition.rung_index_promoted_to,
            next_ladder_state=transition.next_state,
        )

    qty_to_sell = (
        max(
            1,
            int(round(context.quantity * transition.suggested_partial_ratio)),
        )
        if transition.suggested_partial_ratio > 0
        else 0
    )
    exit_action = _ladder_kind_to_exit_action(action_kind)
    stop_price = (
        context.entry_price * (1 + transition.next_state.protected_stop_pct / 100)
        if exit_action == ExitAction.LADDER_HOLD_STOP_PROMOTED
        else None
    )
    return _make_decision(
        action=exit_action,
        quantity_to_sell=qty_to_sell,
        stop_price=stop_price,
        reason=transition.reason,
        context=context,
        active=active,
        pnl=pnl,
        break_even=break_even,
        ladder_rung_index=transition.rung_index_promoted_to,
        next_ladder_state=transition.next_state,
    )


def _ladder_kind_to_exit_action(kind: LadderActionKind) -> ExitAction:
    return {
        LadderActionKind.LADDER_PARTIAL: ExitAction.LADDER_PARTIAL,
        LadderActionKind.LADDER_TAKE_PROFIT: ExitAction.LADDER_TAKE_PROFIT,
        LadderActionKind.LADDER_HOLD_STOP_PROMOTED: ExitAction.LADDER_HOLD_STOP_PROMOTED,
        LadderActionKind.STOP_LOSS_LADDER: ExitAction.STOP_LOSS_LADDER,
    }[kind]


# 7-step priority (plan §8). Lower number = higher priority. Hard exits
# always win over ladder/legacy profit takes.
_PRIORITY_TABLE: dict[ExitAction, int] = {
    # 1 — risk / kill / manual exits (no instances generated by this
    # module; reserved for the pipeline-injected priority).
    ExitAction.STOP_LOSS: 2,
    ExitAction.STOP_LOSS_LADDER: 2,
    ExitAction.EOD_FLATTEN: 2,
    # 3 reserved for the limit-up defer (post-filter, not a candidate).
    ExitAction.LADDER_TAKE_PROFIT: 4,
    ExitAction.LADDER_PARTIAL: 4,
    ExitAction.TAKE_FULL_PROFIT: 5,
    ExitAction.TAKE_PARTIAL_PROFIT: 5,
    ExitAction.TRAIL_PROFIT: 5,
    ExitAction.RAISE_PROFIT_LOCK: 5,
    ExitAction.TRAIL_EMA9_BREAK: 5,
    ExitAction.TRAIL_VWAP_CROSS: 5,
    ExitAction.TIME_EXIT: 6,
    ExitAction.HOLD: 7,
    ExitAction.LADDER_HOLD_STOP_PROMOTED: 7,
    ExitAction.MOVE_STOP_TO_BREAKEVEN: 7,
}


def _select_final_action(candidates: list[ExitDecision]) -> ExitDecision:
    """Pick the single winning decision from the candidate list.

    Sorts by ``(priority, -ladder_rung_index_or_minus_one)`` so that
    when two ladder candidates tie on priority the higher rung wins
    (rare — typically only one ladder candidate is produced per bar).
    """

    if not candidates:
        raise ValueError("_select_final_action requires at least one candidate")

    def key(decision: ExitDecision) -> tuple[int, int]:
        priority = _PRIORITY_TABLE.get(decision.action, 99)
        rung_index = (
            decision.ladder_rung_index if decision.ladder_rung_index is not None else -1
        )
        return (priority, -rung_index)

    return sorted(candidates, key=key)[0]


def _apply_limit_up_defer_to_selected(
    *,
    selected: ExitDecision,
    context: ExitContext,
    hold_state: LimitUpHoldState | None,
) -> ExitDecision:
    """Convert a profit-side decision into HOLD when limit-up close lock fires.

    Plan §5.2 — applies when:
      - close_auction_phase OR is_session_closing_bar
      - is_limit_up AND limit_up_locked
      - not already_deferred_this_session
      - selected.action ∈ deferrable set (profit-side only)
    """

    if hold_state is None:
        return selected
    if not (context.close_auction_phase or context.is_session_closing_bar):
        return selected
    if not (context.is_limit_up and context.limit_up_locked):
        return selected
    if context.session_date is not None and hold_state.already_deferred(
        context.session_date
    ):
        return selected
    if selected.action not in _LIMIT_UP_DEFERRABLE_ACTIONS:
        return selected

    return ExitDecision(
        action=ExitAction.HOLD,
        quantity_to_sell=0,
        stop_price=None,
        reason="limit_up_close_defer_to_next_session",
        net_pnl_krw=selected.net_pnl_krw,
        break_even_price=selected.break_even_price,
        applied_policy=selected.applied_policy,
        ladder_rung_index=selected.ladder_rung_index,
        next_ladder_state=selected.next_ladder_state,
        deferred_from_action=selected.action,
    )


def _has_phase_g_inputs(context: ExitContext) -> bool:
    return (
        context.vwap_value is not None
        and context.prev_vwap_value is not None
        and context.prev_close is not None
        and context.ema9_value is not None
        and context.prev_ema9_value is not None
    )


__all__ = [
    "ExitAction",
    "ExitContext",
    "ExitDecision",
    "ExitPolicy",
    "EXIT_REASON_TRENDLINE_ZONE_BODY_CLOSE_AGAINST",
    "adapt_exit_policy",
    "evaluate_exit",
    "evaluate_safety_exit",
    "is_orderable_exit_action",
    "rapid_profit_lock_price",
]

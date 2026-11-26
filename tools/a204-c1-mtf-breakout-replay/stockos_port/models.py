# stockos-port: source=packages/trading/stockos_trading/models.py commit=fcad869ecc531a6de80777f59b6fb842a2891101 — verbatim copy, do not edit
from __future__ import annotations

import math
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from typing import Any


class Side(str, Enum):
    BUY = "BUY"
    SELL = "SELL"


class BrokerOrderType(str, Enum):
    LIMIT = "limit"
    MARKET = "market"
    MARKETABLE_LIMIT = "marketable_limit"
    STOP_LIMIT = "stop_limit"


class OrderExecutionRoute(str, Enum):
    """Broker-side routing intent for a stock order.

    ``AUTO`` lets the order boundary resolve the safest route from
    market/session/side/intent. Every non-AUTO value maps to a concrete KIS
    order family or wire ``ORD_DVSN``; it is intentionally narrower than a
    free-form string so live routes stay auditable.
    """

    AUTO = "auto"
    REGULAR = "regular"
    US_DAYTIME = "us_daytime"
    US_RESERVED = "us_reserved"
    US_OPEN_MARKET = "us_open_market"
    US_OPEN_LIMIT = "us_open_limit"
    US_CLOSE_MARKET = "us_close_market"
    US_CLOSE_LIMIT = "us_close_limit"
    KR_PRE_MARKET = "kr_pre_market"
    KR_AFTER_MARKET = "kr_after_market"
    KR_AFTER_HOURS_SINGLE = "kr_after_hours_single"


class OrderIntent(str, Enum):
    """Intent of an outgoing order or an in-pipe routing decision.

    The first three values (``ENTRY``, ``EXIT``, ``SCALE_OUT``) describe
    a candidate-side intent on :class:`OrderCandidate`. The remaining
    values (``PROFIT_EXIT``, ``RISK_EXIT``, ``MANUAL_EXIT``,
    ``POSITION_STATE_UPDATE``) are the routing-side intents read by
    ``gates.time_gate`` (PR-L3 plan §6.0.2). The two roles share an
    enum so consumers can pattern-match across both surfaces; existing
    ``OrderCandidate.intent = OrderIntent.ENTRY`` callers keep working
    unchanged.
    """

    ENTRY = "ENTRY"
    EXIT = "EXIT"
    SCALE_OUT = "SCALE_OUT"
    PROFIT_EXIT = "PROFIT_EXIT"
    RISK_EXIT = "RISK_EXIT"
    MANUAL_EXIT = "MANUAL_EXIT"
    POSITION_STATE_UPDATE = "POSITION_STATE_UPDATE"


class TradingMode(str, Enum):
    DRY_RUN = "dry_run"
    PAPER = "paper"
    LIVE = "live"


class SignalType(str, Enum):
    WATCH = "watch"
    ENTRY_CANDIDATE = "entry_candidate"
    SCALE_OUT = "scale_out"
    EXIT = "exit"
    BLOCKED = "blocked"


@dataclass(frozen=True)
class Candle:
    symbol: str
    timestamp: datetime
    open: float
    high: float
    low: float
    close: float
    volume: float

    @property
    def typical_price(self) -> float:
        return (self.high + self.low + self.close) / 3

    @property
    def turnover(self) -> float:
        return self.close * self.volume


@dataclass(frozen=True)
class BollingerBand:
    middle: float
    upper: float
    lower: float
    stddev: float


@dataclass(frozen=True)
class IchimokuSnapshot:
    conversion_line: float | None
    base_line: float | None
    leading_span_a: float | None
    leading_span_b: float | None
    lagging_span: float | None


@dataclass(frozen=True)
class VolumeSnapshot:
    rvol: float
    slope: float
    turnover: float
    spread_pct: float
    breakout_volume_ratio: float


@dataclass(frozen=True)
class OrderCandidate:
    symbol: str
    side: Side
    price: float
    quantity: int
    signal_type: SignalType
    timestamp: datetime
    target_price: float
    stop_price: float
    intent: OrderIntent = OrderIntent.ENTRY
    reasons: tuple[str, ...] = ()
    metadata: dict[str, Any] = field(default_factory=dict)
    # 058 StopPlan V2 Change A — additive 진단 메타(전부 기본 None).
    # Change A는 shadow 전용이라 이 필드를 **채우지 않는다**(항상 None) —
    # 권위 충전은 활성 변경(C/D) 소유. 스키마만 미리 두어 forward-compat 확보.
    # 기존 precommit 계약(target/stop 검증)은 불변.
    stop_plan_version: str | None = None
    stop_source: str | None = None
    structural_stop_class: str | None = None
    structural_stop_class_source: str | None = None

    def __post_init__(self) -> None:
        # Phase D — entry/target/stop precommit contract.
        # Every OrderCandidate must carry an explicit target_price and
        # stop_price ordered correctly against the entry. Validating at
        # construction time means no code path (strategy, replay, fixtures,
        # ad-hoc producers) can smuggle a bare candidate past Guardian.
        for label, value in (
            ("price", self.price),
            ("target_price", self.target_price),
            ("stop_price", self.stop_price),
        ):
            if not math.isfinite(value):
                raise ValueError(f"{label} must be finite, got {value!r}")
            if value <= 0:
                raise ValueError(f"{label} must be positive, got {value!r}")
        if self.quantity <= 0:
            raise ValueError(f"quantity must be positive, got {self.quantity!r}")
        if self.side == Side.BUY:
            if self.target_price <= self.price:
                raise ValueError(
                    f"target_price ({self.target_price}) must exceed entry "
                    f"price ({self.price}) for BUY"
                )
            if self.stop_price >= self.price:
                raise ValueError(
                    f"stop_price ({self.stop_price}) must be below entry "
                    f"price ({self.price}) for BUY"
                )
        # SELL: short entries are blocked elsewhere (execution_sync rejects
        # synthetic shorts). For SELL exits the target/stop ordering is the
        # mirror image, but we only enforce positivity here — the real
        # contract for exits is owned by exit_strategy.evaluate_exit, not
        # this entry-side dataclass.

    @property
    def notional(self) -> float:
        return self.price * self.quantity


@dataclass(frozen=True)
class PositionSnapshot:
    symbol: str
    quantity: int
    avg_price: float
    last_price: float
    market: str = "KRX"

    @property
    def exposure(self) -> float:
        return abs(self.quantity * self.last_price)

    @property
    def unrealized_pnl(self) -> float:
        return (self.last_price - self.avg_price) * self.quantity

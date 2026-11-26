# stockos-port: source=packages/trading/stockos_trading/profit_ladder.py commit=fcad869ecc531a6de80777f59b6fb842a2891101 — verbatim copy, do not edit
"""Profit-target ladder — multi-rung ratchet (PR-L1).

Plan reference: ``docs/implementation/006-plan-exit-ladder-and-session-hours.md``
§4 (PR-L1). Implements the user-requested ratchet:

    initial buy → target up → profit-basis adjust → take-profit adjust →
    next target up → ...

The ladder is a *decision module*. State is hydrated by the repository
layer from ``positions`` columns (see PR-L1 schema delta in plan §4.2)
and mutated through immutable ``LadderState`` snapshots. Order routing,
audit emission, and KIS submission live elsewhere — this module never
touches IO.

Two timing rules (plan §4.5.2) are enforced by separating fields:

* ``activated_rung_index`` and ``protected_stop_pct`` move at *decision*
  time (rung target observed in the current bar).
* ``partial_taken_ratio_total`` and ``completed`` move at *execution*
  time (KIS fill received) — ``execution_sync.apply_execution`` writes
  these via the repository.
"""

from __future__ import annotations

from dataclasses import dataclass, replace
from datetime import date, datetime
from enum import Enum


# WHY: kept here to avoid an import cycle with exit_strategy. The ladder
# branch returns a ``LadderTransition`` whose ``action`` is one of these;
# ``exit_strategy`` re-exports the matching ``ExitAction`` members so the
# pipeline never depends on this module directly. The string values must
# stay in sync with ``ExitAction``'s ladder members.
class LadderActionKind(str, Enum):
    LADDER_PARTIAL = "LADDER_PARTIAL"
    LADDER_TAKE_PROFIT = "LADDER_TAKE_PROFIT"
    LADDER_HOLD_STOP_PROMOTED = "LADDER_HOLD_STOP_PROMOTED"
    STOP_LOSS_LADDER = "STOP_LOSS_LADDER"


# WHY: backtest fill model is configurable per plan §4.5.4. STOP_FIRST is
# the conservative default (same-bar high=target + low=stop → STOP wins).
# Live tick streams supply real event order via two consecutive
# ``evaluate_ladder`` calls, so this enum is only consulted when the
# caller passes ``high`` + ``low`` (a same-bar OHLC backtest).
class FillModel(str, Enum):
    STOP_FIRST = "STOP_FIRST"
    TARGET_FIRST = "TARGET_FIRST"


@dataclass(frozen=True)
class LadderRung:
    """One rung of the ratchet.

    ``stop_pct`` is the protected stop level *relative to entry price*,
    expressed as a percent (positive = entry-and-then-some, 0 = breakeven).
    ``partial_ratio`` is taken from the *remaining* quantity at that rung,
    not from the original entry quantity (see plan §4.5.3 — the cumulative
    ``partial_taken_ratio_total`` is initial-qty-based, but each rung's
    ``partial_ratio`` is remaining-qty-based).
    """

    target_pct: float
    stop_pct: float
    partial_ratio: float

    def __post_init__(self) -> None:
        if not (0.0 <= self.partial_ratio <= 1.0):
            raise ValueError(
                f"partial_ratio must be in [0, 1], got {self.partial_ratio!r}"
            )


@dataclass(frozen=True)
class ProfitLadderPolicy:
    """Immutable rung-set for one symbol/policy combination."""

    policy_id: str
    rungs: tuple[LadderRung, ...]
    final_take_full: bool = True

    def __post_init__(self) -> None:
        if not self.policy_id:
            raise ValueError("policy_id must be non-empty")
        if not self.rungs:
            raise ValueError("ProfitLadderPolicy requires at least one rung")
        # Monotone target_pct, monotone stop_pct (rung_n stop > rung_{n-1} stop
        # is the ratchet contract — protected stop only ever rises).
        prev = self.rungs[0]
        for rung in self.rungs[1:]:
            if rung.target_pct <= prev.target_pct:
                raise ValueError(
                    "rung target_pct must be strictly increasing; "
                    f"{rung.target_pct} <= {prev.target_pct}"
                )
            if rung.stop_pct < prev.stop_pct:
                raise ValueError(
                    "rung stop_pct must be monotonically non-decreasing; "
                    f"{rung.stop_pct} < {prev.stop_pct}"
                )
            prev = rung


@dataclass(frozen=True)
class LadderState:
    """Per-position ladder state. Persisted in ``positions`` columns."""

    policy_id: str
    activated_rung_index: int = -1
    protected_stop_pct: float = 0.0
    last_event_at: datetime | None = None
    partial_taken_ratio_total: float = 0.0
    completed: bool = False
    pending_order_id: str | None = None
    pending_rung_index: int | None = None

    @property
    def has_pending_order(self) -> bool:
        return self.pending_order_id is not None


@dataclass(frozen=True)
class LadderContext:
    """Per-evaluation inputs.

    ``high`` and ``low`` are the bar's intraday extremes; when both are
    populated the evaluator uses the configured ``fill_model`` to break
    same-bar target+stop ties. Live tick streams pass only
    ``current_price`` and call the evaluator twice (once per event).
    """

    entry_price: float
    current_price: float
    moment: datetime
    high: float | None = None
    low: float | None = None
    fill_model: FillModel = FillModel.STOP_FIRST

    @property
    def gross_return_pct(self) -> float:
        return ((self.current_price - self.entry_price) / self.entry_price) * 100


@dataclass(frozen=True)
class LadderTransition:
    """Outcome of one ladder evaluation.

    ``action`` is ``None`` only when nothing happened (no rung changed,
    no stop, no order). ``rung_index_promoted_to`` is the highest rung
    index reached *this evaluation* (not the cumulative max — that lives
    in ``next_state.activated_rung_index``).
    """

    next_state: LadderState
    action: LadderActionKind | None
    reason: str
    rung_index_promoted_to: int | None
    suggested_partial_ratio: float = 0.0  # of remaining quantity, 0..1


def default_policy() -> ProfitLadderPolicy:
    """The CONFIRM-L1 default 4-rung ladder (plan §3)."""

    return ProfitLadderPolicy(
        policy_id="default_v1",
        rungs=(
            LadderRung(target_pct=1.5, stop_pct=0.0, partial_ratio=0.0),
            LadderRung(target_pct=2.5, stop_pct=1.0, partial_ratio=0.25),
            LadderRung(target_pct=4.0, stop_pct=2.0, partial_ratio=0.25),
            LadderRung(target_pct=6.0, stop_pct=3.5, partial_ratio=1.0),
        ),
        final_take_full=True,
    )


def ladder_order_idempotency_key(
    *,
    position_symbol: str,
    policy_id: str,
    rung_index: int,
    action: LadderActionKind,
) -> str:
    """Deterministic key for broker idempotency (plan §4.5.3).

    Same (symbol, policy, rung, action) maps to the same key — broker
    adapter dedupes via ``OrderRecord.idempotency_key`` UNIQUE.
    """

    return f"ladder:{position_symbol}:{policy_id}:r{rung_index}:{action.value}"


def evaluate_ladder(
    context: LadderContext,
    policy: ProfitLadderPolicy,
    state: LadderState,
) -> LadderTransition:
    """Evaluate one bar against the ladder."""

    if state.policy_id != policy.policy_id:
        raise ValueError(
            "ladder state/policy mismatch: "
            f"state={state.policy_id} policy={policy.policy_id}"
        )

    same_bar_decision = _resolve_same_bar_tiebreak(context, policy, state)
    if isinstance(same_bar_decision, LadderTransition):
        return same_bar_decision

    if not same_bar_decision and state.activated_rung_index >= 0:
        stop_transition = _check_protected_stop(context, state)
        if stop_transition is not None:
            return stop_transition

    new_index = _highest_reached_rung_index(context, policy, state)
    if new_index == state.activated_rung_index:
        return LadderTransition(
            next_state=state,
            action=None,
            reason="no_rung_promotion",
            rung_index_promoted_to=None,
        )
    return _build_promotion_transition(context, policy, state, new_index)


def _resolve_same_bar_tiebreak(
    context: LadderContext, policy: ProfitLadderPolicy, state: LadderState
) -> "LadderTransition | bool":
    """Return a STOP transition (stop-first), True (target-first override
    grant), or False (no same-bar conflict)."""

    if not (
        _same_bar_target_observed(context, policy, state)
        and _same_bar_stop_observed(context, state)
    ):
        return False
    if context.fill_model == FillModel.TARGET_FIRST:
        return True
    return _stop_transition(context, state, reason="same_bar_stop_first")


def _check_protected_stop(
    context: LadderContext, state: LadderState
) -> LadderTransition | None:
    stop_price = context.entry_price * (1 + state.protected_stop_pct / 100)
    probe_low = context.low if context.low is not None else context.current_price
    if probe_low < stop_price:
        return _stop_transition(context, state, reason="protected_stop_breached")
    return None


def _highest_reached_rung_index(
    context: LadderContext, policy: ProfitLadderPolicy, state: LadderState
) -> int:
    probe_high = context.high if context.high is not None else context.current_price
    high_return_pct = ((probe_high - context.entry_price) / context.entry_price) * 100
    new_index = state.activated_rung_index
    for idx, rung in enumerate(policy.rungs):
        if high_return_pct >= rung.target_pct and idx > new_index:
            new_index = idx
    return new_index


def _build_promotion_transition(
    context: LadderContext,
    policy: ProfitLadderPolicy,
    state: LadderState,
    new_index: int,
) -> LadderTransition:
    promoted_rung = policy.rungs[new_index]
    next_state = replace(
        state,
        activated_rung_index=new_index,
        protected_stop_pct=promoted_rung.stop_pct,
        last_event_at=context.moment,
        completed=(
            state.completed
            or (
                new_index == len(policy.rungs) - 1
                and not policy.final_take_full
                and promoted_rung.partial_ratio == 0
            )
        ),
    )
    is_last_rung = new_index == len(policy.rungs) - 1
    if is_last_rung and policy.final_take_full:
        return LadderTransition(
            next_state=next_state,
            action=LadderActionKind.LADDER_TAKE_PROFIT,
            reason="ladder_final_rung_take_full",
            rung_index_promoted_to=new_index,
            suggested_partial_ratio=1.0,
        )
    if promoted_rung.partial_ratio > 0:
        return LadderTransition(
            next_state=next_state,
            action=LadderActionKind.LADDER_PARTIAL,
            reason="ladder_rung_partial_take",
            rung_index_promoted_to=new_index,
            suggested_partial_ratio=promoted_rung.partial_ratio,
        )
    return LadderTransition(
        next_state=next_state,
        action=LadderActionKind.LADDER_HOLD_STOP_PROMOTED,
        reason="ladder_rung_state_only_promotion",
        rung_index_promoted_to=new_index,
        suggested_partial_ratio=0.0,
    )


def _same_bar_target_observed(
    context: LadderContext, policy: ProfitLadderPolicy, state: LadderState
) -> bool:
    if context.high is None:
        return False
    high_return_pct = ((context.high - context.entry_price) / context.entry_price) * 100
    for idx, rung in enumerate(policy.rungs):
        if idx > state.activated_rung_index and high_return_pct >= rung.target_pct:
            return True
    return False


def _same_bar_stop_observed(context: LadderContext, state: LadderState) -> bool:
    if context.low is None or state.activated_rung_index < 0:
        return False
    stop_price = context.entry_price * (1 + state.protected_stop_pct / 100)
    return context.low < stop_price


def _stop_transition(
    context: LadderContext, state: LadderState, *, reason: str
) -> LadderTransition:
    next_state = replace(state, last_event_at=context.moment)
    return LadderTransition(
        next_state=next_state,
        action=LadderActionKind.STOP_LOSS_LADDER,
        reason=reason,
        rung_index_promoted_to=state.activated_rung_index,
        suggested_partial_ratio=1.0,
    )


@dataclass(frozen=True)
class LimitUpHoldState:
    """Limit-up close defer state (PR-L2 plan §5.1.0).

    Lives in its own dataclass so the limit-up hold domain is decoupled
    from the ladder domain even though both back onto the same
    ``positions`` row. ``deferred_session_date`` is KST.
    """

    deferred_session_date: date | None = None

    def already_deferred(self, current_session_date: date) -> bool:
        return (
            self.deferred_session_date is not None
            and self.deferred_session_date == current_session_date
        )


__all__ = [
    "FillModel",
    "LadderActionKind",
    "LadderContext",
    "LadderRung",
    "LadderState",
    "LadderTransition",
    "LimitUpHoldState",
    "ProfitLadderPolicy",
    "default_policy",
    "evaluate_ladder",
    "ladder_order_idempotency_key",
]

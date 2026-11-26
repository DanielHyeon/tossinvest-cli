# stockos-port: source=packages/trading/stockos_trading/exit/stop_controls.py commit=fcad869ecc531a6de80777f59b6fb842a2891101 — verbatim copy, do not edit
"""Resolve general versus externally adopted position stop controls."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Protocol


HARD_STOP_GRACE_MINUTES_MIN = 0
HARD_STOP_GRACE_MINUTES_MAX = 30
HARD_STOP_LOSS_PCT_MIN = -5.0
HARD_STOP_LOSS_PCT_MAX = -0.1
CATASTROPHE_STOP_LOSS_PCT_MIN = -5.0
CATASTROPHE_STOP_LOSS_PCT_MAX = -0.5


class _StopControlSource(Protocol):
    exit_hard_stop_loss_pct: float
    exit_hard_stop_grace_minutes: int
    exit_hard_stop_grace_catastrophe_pct: float
    external_exit_hard_stop_loss_pct: float | None
    external_exit_hard_stop_grace_minutes: int | None
    external_exit_hard_stop_grace_catastrophe_pct: float | None


@dataclass(frozen=True)
class StopControls:
    hard_stop_loss_pct: float
    grace_minutes: int
    catastrophe_pct: float
    source: str


def validate_stop_controls(controls: StopControls) -> None:
    if not (
        HARD_STOP_LOSS_PCT_MIN
        <= controls.hard_stop_loss_pct
        <= HARD_STOP_LOSS_PCT_MAX
    ):
        raise ValueError(
            "hard stop must satisfy "
            f"{HARD_STOP_LOSS_PCT_MIN} <= value <= {HARD_STOP_LOSS_PCT_MAX}"
        )
    if not (
        HARD_STOP_GRACE_MINUTES_MIN
        <= controls.grace_minutes
        <= HARD_STOP_GRACE_MINUTES_MAX
    ):
        raise ValueError(
            "grace minutes must satisfy "
            f"{HARD_STOP_GRACE_MINUTES_MIN} <= value <= "
            f"{HARD_STOP_GRACE_MINUTES_MAX}"
        )
    if not (
        CATASTROPHE_STOP_LOSS_PCT_MIN
        <= controls.catastrophe_pct
        <= CATASTROPHE_STOP_LOSS_PCT_MAX
    ):
        raise ValueError(
            "catastrophe must satisfy "
            f"{CATASTROPHE_STOP_LOSS_PCT_MIN} <= value <= "
            f"{CATASTROPHE_STOP_LOSS_PCT_MAX}"
        )
    if controls.catastrophe_pct > controls.hard_stop_loss_pct:
        raise ValueError("catastrophe must be less than or equal to hard stop")


def resolve_stop_controls(
    runtime: _StopControlSource,
    *,
    external_adopted: bool,
) -> StopControls:
    external_values = (
        getattr(runtime, "external_exit_hard_stop_loss_pct", None),
        getattr(runtime, "external_exit_hard_stop_grace_minutes", None),
        getattr(runtime, "external_exit_hard_stop_grace_catastrophe_pct", None),
    )
    configured_count = sum(value is not None for value in external_values)
    if configured_count not in {0, 3}:
        raise ValueError("external stop controls must be all set or all null")

    if external_adopted and configured_count == 3:
        controls = StopControls(
            hard_stop_loss_pct=float(external_values[0]),
            grace_minutes=int(external_values[1]),
            catastrophe_pct=float(external_values[2]),
            source="external",
        )
    else:
        controls = StopControls(
            hard_stop_loss_pct=float(runtime.exit_hard_stop_loss_pct),
            grace_minutes=int(runtime.exit_hard_stop_grace_minutes),
            catastrophe_pct=float(
                runtime.exit_hard_stop_grace_catastrophe_pct
            ),
            source="general",
        )

    # External overrides are new in a169, so every persisted triplet must obey
    # the new range. General controls predate a169 and historically allowed
    # wider values; rejecting those during a live read could stop the entire
    # exit evaluator. New general PATCHes are validated at the write boundary.
    if controls.source == "external":
        validate_stop_controls(controls)
    return controls

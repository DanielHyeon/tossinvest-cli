# stockos-port: source=packages/trading/stockos_trading/ladder/profile.py commit=fcad869ecc531a6de80777f59b6fb842a2891101 — verbatim copy, do not edit
"""LadderProfile — 4-rung profile with PM-6 ``T4Action`` hook.

Plan reference:
``docs/implementation/013-plan-profit-maximization-exit-architecture.md``
§5 PR-PM5 + §5 PR-PM6.

A profile bundles:

* a ``ProfitLadderPolicy`` (PR-L1 surface — rungs)
* a ``T4Action`` enum value (PR-PM6 hook; PR-PM5 keeps the existing
  PR-L1 ``FULL_EXIT`` semantics for BALANCED only)
* an ``experimental`` flag — non-BALANCED profiles default to
  experimental and are rejected from live allowlists by PR-PM12.

Numeric rungs are frozen v0 default (Q1 lock). Changes require ADR.
"""

from __future__ import annotations

from dataclasses import dataclass
from enum import Enum

from ..profit_ladder import LadderRung, ProfitLadderPolicy


class T4Action(str, Enum):
    """What the ladder does when the highest rung triggers (PR-PM6).

    PR-PM5 ships the enum but only ``FULL_EXIT`` is wired on adapter
    side. PR-PM6 will activate the others.
    """

    FULL_EXIT = "FULL_EXIT"
    PARTIAL_AND_TRAIL = "PARTIAL_AND_TRAIL"
    ENTER_RUNNER_MODE = "ENTER_RUNNER_MODE"
    RAISE_FLOOR_ONLY = "RAISE_FLOOR_ONLY"


@dataclass(frozen=True)
class LadderProfile:
    """One named profile."""

    name: str
    policy: ProfitLadderPolicy
    t4_action: T4Action
    experimental: bool = True

    def __post_init__(self) -> None:
        if not self.name:
            raise ValueError("profile name must be non-empty")
        # WHY: BALANCED.policy.policy_id is "default_v1" (Q1 lock — bit-equal
        # to PR-L1's default_policy()) while profile.name is "balanced" so
        # operators can refer to the user-facing label. Other profiles match
        # name == policy_id for clarity but the contract is "non-empty
        # policy_id" only.
        if not self.policy.policy_id:
            raise ValueError("policy.policy_id must be non-empty")


def _build_balanced() -> LadderProfile:
    """BALANCED — bit-equal to PR-L1 ``default_policy()`` (Q1 lock).

    The hash regression test in tests/test_pm5_ladder_profile.py
    enforces this byte-for-byte.
    """

    # WHY policy_id="default_v1": Q1 lock requires BALANCED bit-equal to
    # PR-L1 ``default_policy()``. policy_id is part of LadderState's
    # equality so reusing it preserves repository compatibility for any
    # state row created before PR-PM5.
    policy = ProfitLadderPolicy(
        policy_id="default_v1",
        rungs=(
            LadderRung(target_pct=1.5, stop_pct=0.0, partial_ratio=0.0),
            LadderRung(target_pct=2.5, stop_pct=1.0, partial_ratio=0.25),
            LadderRung(target_pct=4.0, stop_pct=2.0, partial_ratio=0.25),
            LadderRung(target_pct=6.0, stop_pct=3.5, partial_ratio=1.0),
        ),
        final_take_full=True,
    )
    return LadderProfile(
        name="balanced",
        policy=policy,
        t4_action=T4Action.FULL_EXIT,
        experimental=False,
    )


def _build_conservative() -> LadderProfile:
    """CONSERVATIVE — earlier partials, tighter stops (Q1 lock)."""

    policy = ProfitLadderPolicy(
        policy_id="conservative",
        rungs=(
            LadderRung(target_pct=1.0, stop_pct=0.0, partial_ratio=0.0),
            LadderRung(target_pct=1.8, stop_pct=0.7, partial_ratio=0.40),
            LadderRung(target_pct=3.0, stop_pct=1.5, partial_ratio=0.40),
            LadderRung(target_pct=4.5, stop_pct=2.5, partial_ratio=1.0),
        ),
        final_take_full=True,
    )
    return LadderProfile(
        name="conservative",
        policy=policy,
        t4_action=T4Action.FULL_EXIT,
        experimental=True,
    )


def _build_defensive() -> LadderProfile:
    """DEFENSIVE — fast baseline lock for shakeout-prone entries."""

    policy = ProfitLadderPolicy(
        policy_id="defensive",
        rungs=(
            LadderRung(target_pct=0.8, stop_pct=0.0, partial_ratio=0.0),
            LadderRung(target_pct=1.2, stop_pct=0.3, partial_ratio=0.40),
            LadderRung(target_pct=2.0, stop_pct=0.8, partial_ratio=0.35),
            LadderRung(target_pct=3.2, stop_pct=1.5, partial_ratio=1.0),
        ),
        final_take_full=True,
    )
    return LadderProfile(
        name="defensive",
        policy=policy,
        t4_action=T4Action.FULL_EXIT,
        experimental=True,
    )


def _build_trend() -> LadderProfile:
    """TREND — wider rungs, T4 hands off to runner mode (Q1 lock).

    PM-5 still ships ``T4Action.ENTER_RUNNER_MODE`` but PM-6 wires the
    runner. Until then the adapter treats T4 as a normal full-exit
    rung — captured by an explicit unit test in PR-PM5 so PR-PM6
    swap is a single-line change.
    """

    policy = ProfitLadderPolicy(
        policy_id="trend",
        rungs=(
            LadderRung(target_pct=2.0, stop_pct=0.0, partial_ratio=0.0),
            LadderRung(target_pct=3.5, stop_pct=1.5, partial_ratio=0.20),
            LadderRung(target_pct=5.5, stop_pct=2.5, partial_ratio=0.25),
            # T4 partial=0 means "no extra sell" — the runner takes over.
            # final_take_full=False keeps PR-L1 ladder from forcing a sell.
            LadderRung(target_pct=8.0, stop_pct=4.0, partial_ratio=0.0),
        ),
        final_take_full=False,
    )
    return LadderProfile(
        name="trend",
        policy=policy,
        t4_action=T4Action.ENTER_RUNNER_MODE,
        experimental=True,
    )


def _build_runner() -> LadderProfile:
    """RUNNER — minimal partial, trailing-only T4 (Q1 lock).

    The fourth rung's ``target_pct`` cannot be ``None`` in PR-L1
    ``LadderRung`` (it stays a number); the meaning of "no fixed
    target" is encoded by ``T4Action.RAISE_FLOOR_ONLY`` and a target
    far enough away (we use ``999.0`` as a sentinel) that the ladder
    never auto-promotes there. PR-PM6 wiring will short-circuit T4
    evaluation for this profile.
    """

    policy = ProfitLadderPolicy(
        policy_id="runner",
        rungs=(
            LadderRung(target_pct=2.5, stop_pct=0.0, partial_ratio=0.0),
            LadderRung(target_pct=4.5, stop_pct=2.0, partial_ratio=0.15),
            LadderRung(target_pct=7.0, stop_pct=3.5, partial_ratio=0.0),
            LadderRung(target_pct=999.0, stop_pct=5.0, partial_ratio=0.0),
        ),
        final_take_full=False,
    )
    return LadderProfile(
        name="runner",
        policy=policy,
        t4_action=T4Action.RAISE_FLOOR_ONLY,
        experimental=True,
    )


def _build_hybrid_50() -> LadderProfile:
    """HYBRID_50 — realise half, then trail the remaining half.

    Partial ratios are applied to the remaining quantity.  Selling 25% at T2
    leaves 75%; selling one third of that at T3 leaves exactly 50% before
    integer order rounding.  T4 raises the protected floor and hands the
    remaining position to the persistent high-water runner.
    """

    policy = ProfitLadderPolicy(
        policy_id="hybrid_50",
        rungs=(
            LadderRung(target_pct=1.8, stop_pct=0.0, partial_ratio=0.0),
            LadderRung(target_pct=3.0, stop_pct=1.2, partial_ratio=0.25),
            LadderRung(target_pct=4.8, stop_pct=2.5, partial_ratio=1.0 / 3.0),
            LadderRung(target_pct=6.5, stop_pct=3.8, partial_ratio=0.0),
        ),
        final_take_full=False,
    )
    return LadderProfile(
        name="hybrid_50",
        policy=policy,
        t4_action=T4Action.ENTER_RUNNER_MODE,
        experimental=False,
    )


BALANCED: LadderProfile = _build_balanced()
CONSERVATIVE: LadderProfile = _build_conservative()
DEFENSIVE: LadderProfile = _build_defensive()
TREND: LadderProfile = _build_trend()
RUNNER: LadderProfile = _build_runner()
HYBRID_50: LadderProfile = _build_hybrid_50()


__all__ = [
    "BALANCED",
    "CONSERVATIVE",
    "DEFENSIVE",
    "HYBRID_50",
    "LadderProfile",
    "RUNNER",
    "T4Action",
    "TREND",
]

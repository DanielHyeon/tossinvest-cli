# stockos-port: source=packages/trading/stockos_trading/session_hours.py commit=fcad869ecc531a6de80777f59b6fb842a2891101 — verbatim copy, do not edit
"""Session-hours configuration (PR-L4).

Plan reference: ``docs/implementation/006-plan-exit-ladder-and-session-hours.md``
§7. Loads KRX 2026-09-14 extended-hours wiring + NXT (external ATS)
hours from environment so:

- KRX schedule changes only need a ``.env`` flip, not a code release.
- The NXT/extended hours classification can run in observe-only mode
  (``EXTENDED_HOURS_ENABLED=true`` + ``EXTENDED_HOURS_TRADE_ENABLED=false``)
  before any ADR turns on real routing.

The default values mirror the latest KRX/NXT public schedule at the
time of writing (2026-05-03). Any drift is caught by the Acceptance
Gate (plan §14): "PR-L4 merge 전 KRX/NXT/FSC 최신 공지 재확인".
"""

from __future__ import annotations

from collections.abc import Mapping
from dataclasses import dataclass
from datetime import date, time


class ConfigurationError(Exception):
    """Raised when env-supplied session config is malformed."""


def _parse_time(value: str, *, key: str) -> time:
    parts = value.strip().split(":")
    if len(parts) != 2:
        raise ConfigurationError(
            f"{key} must be HH:MM (got {value!r})"
        )
    try:
        hour = int(parts[0])
        minute = int(parts[1])
    except ValueError as exc:
        raise ConfigurationError(
            f"{key} must be HH:MM with integer fields (got {value!r})"
        ) from exc
    if not (0 <= hour <= 23 and 0 <= minute <= 59):
        raise ConfigurationError(
            f"{key} hour/minute out of range (got {value!r})"
        )
    return time(hour, minute)


def _parse_bool(value: str, *, key: str) -> bool:
    lowered = value.strip().lower()
    if lowered in {"true", "1", "yes", "on"}:
        return True
    if lowered in {"false", "0", "no", "off", ""}:
        return False
    raise ConfigurationError(f"{key} must be a boolean (got {value!r})")


def _parse_date(value: str, *, key: str) -> date:
    try:
        return date.fromisoformat(value.strip())
    except ValueError as exc:
        raise ConfigurationError(
            f"{key} must be ISO date YYYY-MM-DD (got {value!r})"
        ) from exc


@dataclass(frozen=True)
class ExtendedHoursConfig:
    """KRX 2026-09-14 extended-hours configuration.

    ``enabled`` toggles classification (does the gate emit
    ``EXTENDED_KRX_*_PENDING`` for the extra hours?). ``trade_enabled``
    toggles routing on top of classification. v1 default: both False.
    """

    enabled: bool
    trade_enabled: bool
    effective_date: date
    pre_start: time
    pre_end: time
    after_start: time
    after_end: time

    def __post_init__(self) -> None:
        if self.pre_end <= self.pre_start:
            raise ConfigurationError(
                "EXTENDED pre_end must be strictly after pre_start"
            )
        if self.after_end <= self.after_start:
            raise ConfigurationError(
                "EXTENDED after_end must be strictly after after_start"
            )

    @classmethod
    def from_env(cls, env: Mapping[str, str]) -> "ExtendedHoursConfig":
        return cls(
            enabled=_parse_bool(
                env.get("EXTENDED_HOURS_ENABLED", "false"),
                key="EXTENDED_HOURS_ENABLED",
            ),
            trade_enabled=_parse_bool(
                env.get("EXTENDED_HOURS_TRADE_ENABLED", "false"),
                key="EXTENDED_HOURS_TRADE_ENABLED",
            ),
            effective_date=_parse_date(
                env.get("EXTENDED_HOURS_EFFECTIVE_DATE", "2026-09-14"),
                key="EXTENDED_HOURS_EFFECTIVE_DATE",
            ),
            pre_start=_parse_time(
                env.get("KRX_EXTENDED_PRE_MARKET_START", "07:00"),
                key="KRX_EXTENDED_PRE_MARKET_START",
            ),
            pre_end=_parse_time(
                env.get("KRX_EXTENDED_PRE_MARKET_END", "07:50"),
                key="KRX_EXTENDED_PRE_MARKET_END",
            ),
            after_start=_parse_time(
                env.get("KRX_EXTENDED_AFTER_MARKET_START", "16:00"),
                key="KRX_EXTENDED_AFTER_MARKET_START",
            ),
            after_end=_parse_time(
                env.get("KRX_EXTENDED_AFTER_MARKET_END", "20:00"),
                key="KRX_EXTENDED_AFTER_MARKET_END",
            ),
        )


@dataclass(frozen=True)
class NxtHoursConfig:
    """NXT (external ATS) hours. v1 default 08:00-08:50 / 09:00-15:20 / 15:30-20:00.

    Different KRX/FSC public sources publish slightly different start
    times for the after-market window (15:30 vs 15:40); we default to
    15:30 and let operators override via env without a code change.
    """

    pre_start: time
    pre_end: time
    regular_start: time
    regular_end: time
    after_start: time
    after_end: time

    def __post_init__(self) -> None:
        if self.pre_end <= self.pre_start:
            raise ConfigurationError("NXT pre_end must be after pre_start")
        if self.regular_end <= self.regular_start:
            raise ConfigurationError("NXT regular_end must be after regular_start")
        if self.after_end <= self.after_start:
            raise ConfigurationError("NXT after_end must be after after_start")

    @classmethod
    def from_env(cls, env: Mapping[str, str]) -> "NxtHoursConfig":
        return cls(
            pre_start=_parse_time(
                env.get("NXT_PRE_MARKET_START", "08:00"),
                key="NXT_PRE_MARKET_START",
            ),
            pre_end=_parse_time(
                env.get("NXT_PRE_MARKET_END", "08:50"),
                key="NXT_PRE_MARKET_END",
            ),
            regular_start=_parse_time(
                env.get("NXT_REGULAR_START", "09:00"),
                key="NXT_REGULAR_START",
            ),
            regular_end=_parse_time(
                env.get("NXT_REGULAR_END", "15:20"),
                key="NXT_REGULAR_END",
            ),
            after_start=_parse_time(
                env.get("NXT_AFTER_MARKET_START", "15:30"),
                key="NXT_AFTER_MARKET_START",
            ),
            after_end=_parse_time(
                env.get("NXT_AFTER_MARKET_END", "20:00"),
                key="NXT_AFTER_MARKET_END",
            ),
        )


@dataclass(frozen=True)
class SessionHoursConfig:
    """Aggregate session config — KRX extended + NXT.

    Single object passed to ``classify_time_window`` /
    ``evaluate_time_gate`` so callers do not juggle two parameters.
    """

    extended_krx: ExtendedHoursConfig
    nxt: NxtHoursConfig

    @classmethod
    def from_env(cls, env: Mapping[str, str]) -> "SessionHoursConfig":
        return cls(
            extended_krx=ExtendedHoursConfig.from_env(env),
            nxt=NxtHoursConfig.from_env(env),
        )


__all__ = [
    "ConfigurationError",
    "ExtendedHoursConfig",
    "NxtHoursConfig",
    "SessionHoursConfig",
]

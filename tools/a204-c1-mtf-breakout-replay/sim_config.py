"""시뮬레이터 청산 설정 — StockOS 와 같은 이름 · 기본값 · 허용 범위(design §8). TossOS 런타임 설정이 아님.

근거(StockOS fcad869e): 기본 runtime_config.py:370-404, 범위 routes/settings.py:772-826 ·
runtime_config.py:860-872 · 946-1097(교차 규칙), exit/stop_controls.py:9-14 · 34-67.
"""

from __future__ import annotations

import hashlib
import json
import math
from pathlib import Path


class ConfigError(ValueError):
    def __init__(self, field: str, message: str):
        super().__init__(f"{field}: {message}")
        self.field = field


# 이름 → (기본값, 최소, 최대, 정수 여부)
NUMERIC = {
    "exit_hard_stop_loss_pct": (-1.2, -5.0, -0.1, False),
    "exit_weak_trend_stop_loss_pct": (-0.8, -5.0, -0.1, False),
    "exit_trailing_stop_gap_pct": (4.5, 0.2, 10.0, False),
    "exit_trail_stop_pct": (6.5, 0.5, 15.0, False),
    "max_hold_minutes": (90, 1, 390, True),
    "krx_first_take_profit_pct": (1.2, 0.1, 20.0, False),
    "krx_second_take_profit_pct": (2.4, 0.1, 20.0, False),
    "krx_partial_exit_ratio": (0.5, 0.05, 1.0, False),
    # 하한은 settings 라우트(0.0)가 아니라 런타임 안전 하한(runtime_config.py:860-864) — 엄격한 쪽
    "rapid_profit_lock_base_capture_ratio": (0.60, 0.60, 0.95, False),
    "rapid_profit_lock_strong_capture_ratio": (0.65, 0.65, 0.95, False),
    "rapid_profit_lock_tight_capture_ratio": (0.70, 0.70, 0.95, False),
    "exit_hard_stop_grace_minutes": (1, 0, 30, True),
    "exit_hard_stop_grace_catastrophe_pct": (-1.5, -5.0, -0.5, False),
    "exit_breakeven_atr_multiplier": (1.0, 0.0, 3.0, False),
    "exit_stop_confirm_seconds": (0.0, 0.0, 30.0, False),
    "exit_stop_atr_multiplier": (0.0, 0.0, 3.0, False),
}
LADDER_POLICIES = (
    "COMMON_LADDER_BALANCED",
    "COMMON_LADDER_CONSERVATIVE",
    "COMMON_LADDER_DEFENSIVE",
    "COMMON_LADDER_TREND",
    "COMMON_LADDER_RUNNER",
)  # HYBRID_50 은 runner 인계를 재구현하지 않아 거절(design §4.2)


def _number(field: str, value, lo, hi, integer: bool):
    if isinstance(value, bool) or not isinstance(value, (int, float)):
        raise ConfigError(field, "숫자여야 함")
    if integer and not isinstance(value, int):
        raise ConfigError(field, "정수여야 함")
    if not math.isfinite(value) or not lo <= value <= hi:
        raise ConfigError(field, f"{lo:g} 이상 {hi:g} 이하여야 함 (받은 값 {value!r})")
    return value if integer else float(value)  # digest 가 1 과 1.0 을 다르게 세지 않게 실수 필드는 float 로


def validate(values: dict) -> dict:
    """기본값을 채우고 StockOS 와 같은 규칙으로 검사함. 위반은 ConfigError(필드 이름 포함)."""

    for key in values:
        if key not in NUMERIC and key not in ("eod_flatten_enabled", "common_exit_policy"):
            raise ConfigError(key, "알 수 없는 설정")
    cfg = {k: _number(k, values.get(k, d), lo, hi, i) for k, (d, lo, hi, i) in NUMERIC.items()}
    eod = values.get("eod_flatten_enabled", False)
    if not isinstance(eod, bool):
        raise ConfigError("eod_flatten_enabled", "true 또는 false 여야 함")
    cfg["eod_flatten_enabled"] = eod
    ladder = values.get("common_exit_policy")
    if ladder is not None and ladder not in LADDER_POLICIES:
        raise ConfigError("common_exit_policy", f"null 또는 {', '.join(LADDER_POLICIES)} 중 하나여야 함")
    cfg["common_exit_policy"] = ladder

    # 교차 규칙 — runtime_config.py:1018-1025 · 1078-1088 · 979-987, stop_controls.py:66-67
    if cfg["exit_weak_trend_stop_loss_pct"] < cfg["exit_hard_stop_loss_pct"]:
        raise ConfigError("exit_weak_trend_stop_loss_pct", "exit_hard_stop_loss_pct 이상(같거나 더 좁게)이어야 함")
    if cfg["krx_second_take_profit_pct"] < cfg["krx_first_take_profit_pct"]:
        raise ConfigError("krx_second_take_profit_pct", "krx_first_take_profit_pct 이상이어야 함")
    if cfg["exit_hard_stop_grace_catastrophe_pct"] > cfg["exit_hard_stop_loss_pct"]:
        raise ConfigError("exit_hard_stop_grace_catastrophe_pct", "exit_hard_stop_loss_pct 이하여야 함")
    if cfg["rapid_profit_lock_strong_capture_ratio"] < cfg["rapid_profit_lock_base_capture_ratio"]:
        raise ConfigError("rapid_profit_lock_strong_capture_ratio", "base ≤ strong ≤ tight 여야 함")
    if cfg["rapid_profit_lock_tight_capture_ratio"] < cfg["rapid_profit_lock_strong_capture_ratio"]:
        raise ConfigError("rapid_profit_lock_tight_capture_ratio", "base ≤ strong ≤ tight 여야 함")
    return cfg


def load(path: Path) -> dict:
    try:
        values = json.loads(Path(path).read_text(encoding="utf-8"))
    except (OSError, ValueError) as exc:
        raise ConfigError("<file>", f"읽을 수 없음: {exc}") from exc
    if not isinstance(values, dict):
        raise ConfigError("<file>", "JSON 객체여야 함")
    return validate(values)


def digest(cfg: dict) -> str:
    return hashlib.sha256(json.dumps(cfg, sort_keys=True, separators=(",", ":")).encode()).hexdigest()

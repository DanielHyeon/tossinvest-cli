"""설정 검증 시험 — StockOS 와 같은 이름 · 기본 · 범위(design §8). 경계값은 받고 바로 바깥은 필드 이름과 함께 거절."""

from __future__ import annotations

import json
import tempfile
import unittest
from pathlib import Path

import sim_config as sc

RANGES = {
    "exit_hard_stop_loss_pct": (-5.0, -0.1),
    "exit_weak_trend_stop_loss_pct": (-5.0, -0.1),
    "exit_trailing_stop_gap_pct": (0.2, 10.0),
    "exit_trail_stop_pct": (0.5, 15.0),
    "max_hold_minutes": (1, 390),
    "krx_first_take_profit_pct": (0.1, 20.0),
    "krx_second_take_profit_pct": (0.1, 20.0),
    "krx_partial_exit_ratio": (0.05, 1.0),
    "rapid_profit_lock_base_capture_ratio": (0.60, 0.95),
    "rapid_profit_lock_strong_capture_ratio": (0.65, 0.95),
    "rapid_profit_lock_tight_capture_ratio": (0.70, 0.95),
    "exit_hard_stop_grace_minutes": (0, 30),
    "exit_hard_stop_grace_catastrophe_pct": (-5.0, -0.5),
    "exit_breakeven_atr_multiplier": (0.0, 3.0),
    "exit_stop_confirm_seconds": (0.0, 30.0),
    "exit_stop_atr_multiplier": (0.0, 3.0),
}

# 경계에서 교차 규칙을 깨지 않게 함께 바꿀 짝(값 하나만 움직이면 교차 규칙이 먼저 걸리는 필드)
COMPANIONS = {
    ("exit_hard_stop_loss_pct", "min"): {"exit_hard_stop_grace_catastrophe_pct": -5.0},
    ("exit_hard_stop_loss_pct", "max"): {"exit_weak_trend_stop_loss_pct": -0.1, "exit_hard_stop_grace_catastrophe_pct": -0.5},
    ("exit_weak_trend_stop_loss_pct", "min"): {"exit_hard_stop_loss_pct": -5.0, "exit_hard_stop_grace_catastrophe_pct": -5.0},
    ("exit_hard_stop_grace_catastrophe_pct", "max"): {"exit_hard_stop_loss_pct": -0.5, "exit_weak_trend_stop_loss_pct": -0.5},
    ("krx_first_take_profit_pct", "max"): {"krx_second_take_profit_pct": 20.0},
    ("krx_second_take_profit_pct", "min"): {"krx_first_take_profit_pct": 0.1},
    ("rapid_profit_lock_base_capture_ratio", "max"): {"rapid_profit_lock_strong_capture_ratio": 0.95, "rapid_profit_lock_tight_capture_ratio": 0.95},
    ("rapid_profit_lock_strong_capture_ratio", "max"): {"rapid_profit_lock_tight_capture_ratio": 0.95},
}


class DefaultsTest(unittest.TestCase):
    def test_defaults_match_stockos(self):
        cfg = sc.validate({})
        self.assertEqual(cfg["exit_hard_stop_loss_pct"], -1.2)
        self.assertEqual(cfg["exit_trailing_stop_gap_pct"], 4.5)
        self.assertEqual(cfg["exit_trail_stop_pct"], 6.5)
        self.assertEqual(cfg["krx_second_take_profit_pct"], 2.4)
        self.assertEqual(cfg["exit_breakeven_atr_multiplier"], 1.0)
        self.assertIsNone(cfg["common_exit_policy"])
        self.assertFalse(cfg["eod_flatten_enabled"])
        self.assertEqual(set(cfg), set(RANGES) | {"eod_flatten_enabled", "common_exit_policy"})

    def test_default_file_equals_defaults(self):
        path = Path(__file__).resolve().parent.parent / "config.default.json"
        self.assertEqual(sc.load(path), sc.validate({}))


class RangeTest(unittest.TestCase):
    def test_each_boundary_accepted_and_just_outside_rejected(self):
        for field, (lo, hi) in RANGES.items():
            step = 1 if isinstance(lo, int) and not isinstance(lo, float) else 1e-6
            for side, edge, outside in (("min", lo, lo - step), ("max", hi, hi + step)):
                companions = COMPANIONS.get((field, side), {})
                with self.subTest(field=field, side=side):
                    sc.validate({**companions, field: edge})
                    with self.assertRaises(sc.ConfigError) as ctx:
                        sc.validate({**companions, field: outside})
                    self.assertEqual(ctx.exception.field, field)

    def test_cross_field_rules(self):
        cases = [
            ({"exit_hard_stop_loss_pct": -1.0, "exit_weak_trend_stop_loss_pct": -1.1}, "exit_weak_trend_stop_loss_pct"),
            ({"krx_first_take_profit_pct": 2.0, "krx_second_take_profit_pct": 1.9}, "krx_second_take_profit_pct"),
            ({"exit_hard_stop_loss_pct": -1.6}, None),  # catastrophe −1.5 > hard −1.6 → 거절
            ({"rapid_profit_lock_base_capture_ratio": 0.70, "rapid_profit_lock_strong_capture_ratio": 0.69}, "rapid_profit_lock_strong_capture_ratio"),
        ]
        for values, field in cases:
            with self.subTest(values=values), self.assertRaises(sc.ConfigError) as ctx:
                sc.validate(values)
            if field is not None:
                self.assertEqual(ctx.exception.field, field)
        with self.assertRaises(sc.ConfigError) as ctx:
            sc.validate({"exit_hard_stop_loss_pct": -1.6})
        self.assertEqual(ctx.exception.field, "exit_hard_stop_grace_catastrophe_pct")
        sc.validate({"exit_hard_stop_loss_pct": -1.6, "exit_hard_stop_grace_catastrophe_pct": -1.6})

    def test_types_and_unknown_keys(self):
        for values, field in [
            ({"exit_hard_stop_grace_minutes": 1.5}, "exit_hard_stop_grace_minutes"),
            ({"max_hold_minutes": True}, "max_hold_minutes"),
            ({"exit_trail_stop_pct": True}, "exit_trail_stop_pct"),
            ({"exit_trail_stop_pct": "6.5"}, "exit_trail_stop_pct"),
            ({"eod_flatten_enabled": 1}, "eod_flatten_enabled"),
            ({"exit_trail_stop_pct": float("nan")}, "exit_trail_stop_pct"),
            ({"no_such_setting": 1}, "no_such_setting"),
            ({"common_exit_policy": "COMMON_LADDER_HYBRID_50"}, "common_exit_policy"),
            ({"common_exit_policy": "PARKER_VWAP_EXIT"}, "common_exit_policy"),
        ]:
            with self.subTest(values=values), self.assertRaises(sc.ConfigError) as ctx:
                sc.validate(values)
            self.assertEqual(ctx.exception.field, field)
        self.assertEqual(sc.validate({"common_exit_policy": "COMMON_LADDER_BALANCED"})["common_exit_policy"],
                         "COMMON_LADDER_BALANCED")

    def test_integer_valued_float_ok_for_float_fields(self):
        self.assertEqual(sc.validate({"exit_trail_stop_pct": 7})["exit_trail_stop_pct"], 7)


class DigestTest(unittest.TestCase):
    def test_digest_stable_and_sensitive(self):
        a = sc.digest(sc.validate({}))
        self.assertEqual(a, sc.digest(sc.validate({"exit_trail_stop_pct": 6.5})))
        self.assertNotEqual(a, sc.digest(sc.validate({"exit_trail_stop_pct": 6.6})))
        self.assertRegex(a, r"^[0-9a-f]{64}$")

    def test_load_rejects_non_object(self):
        with tempfile.TemporaryDirectory() as d:
            p = Path(d) / "c.json"
            p.write_text(json.dumps([1]))
            with self.assertRaises(sc.ConfigError):
                sc.load(p)


if __name__ == "__main__":
    unittest.main()

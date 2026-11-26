"""AEE 재구현 시험 — 틱 경로 순서 · floor 단조 · 세션 조임 경계 · grace · EOD."""

from __future__ import annotations

import unittest
from datetime import timedelta

import aee_exits as ax
import sim_config
from c1_entry import Bar, Entry
from tests.helpers import DAY, at


def entry(price=10000, stop=9000, minute=(10, 0), atr=None):
    return Entry(time=at(DAY, *minute), price=price, stop=stop, level=price, atr_pct=atr)


def minute_bars(start, specs):
    return [Bar(start + timedelta(minutes=i), *s) for i, s in enumerate(specs)]


CFG = sim_config.validate({})


class PathTest(unittest.TestCase):
    def test_path_walks_low_before_high(self):
        path = ax.intrabar_path(10000, 10030, 9980, 10010)
        self.assertEqual(path, [10000, 9990, 9980, 9990, 10000, 10010, 10020, 10030, 10020, 10010])

    def test_path_across_tick_band(self):
        self.assertEqual(ax.walk(2005, 1998), [2000, 1999, 1998])  # 2000 아래는 1원 단위
        self.assertEqual(ax.walk(1998, 2005), [1999, 2000, 2005])

    def test_stop_first_when_bar_touches_both(self):
        # 진입 다음 봉(grace 밖)이 TP1(+1.2%)과 하드 손절(−1.4%)을 다 건드림 → 손절이 먼저
        bars = minute_bars(at(DAY, 10, 0), [(10000, 10000, 10000, 10000, 1), (10000, 10200, 9850, 10100, 1)])
        t = ax.simulate(entry(), bars, CFG, mode="live")
        self.assertEqual(len(t.fills), 1)
        qty, price, action = t.fills[0].qty, t.fills[0].price, t.fills[0].action
        self.assertEqual(action, "STOP_LOSS")
        self.assertEqual(price, 9860 - 10)  # −1.4% 첫 틱의 bid
        self.assertEqual(qty, ax.Q0)

    def test_profit_first_only_when_low_not_reached(self):
        bars = minute_bars(at(DAY, 10, 0), [(10000, 10000, 10000, 10000, 1), (10000, 10200, 9900, 10100, 1)])
        t = ax.simulate(entry(), bars, CFG, mode="live")
        self.assertEqual(t.fills[0].action, "TAKE_PARTIAL_PROFIT")
        self.assertEqual(t.fills[0].qty, ax.Q0 // 2)


class FloorTest(unittest.TestCase):
    def test_ratchet_only_raises(self):
        st = ax.PositionState.open(entry(stop=9000))
        ax.raise_trailing_floor(st, 11000, 6.5)
        self.assertAlmostEqual(st.current_floor, 11000 * 0.935)
        self.assertTrue(st.protected_floor_enabled)
        ax.raise_trailing_floor(st, 10500, 6.5)
        self.assertAlmostEqual(st.current_floor, 11000 * 0.935)

    def test_ratchet_needs_to_beat_initial_stop(self):
        st = ax.PositionState.open(entry(stop=9900))
        ax.raise_trailing_floor(st, 10500, 6.5)  # 9817.5 < 9900
        self.assertIsNone(st.current_floor)
        self.assertFalse(st.protected_floor_enabled)

    def test_protected_floor_breach_exits_all(self):
        st_entry = entry(stop=9000)
        bars = minute_bars(at(DAY, 10, 0), [(10000, 10000, 10000, 10000, 1), (10000, 12000, 10000, 12000, 1),
                                            (12000, 12000, 11000, 11000, 1)])
        cfg = sim_config.validate({"krx_first_take_profit_pct": 20.0, "krx_second_take_profit_pct": 20.0})
        t = ax.simulate(st_entry, bars, cfg, mode="live")
        last = t.fills[-1]
        self.assertEqual(last.reason, "protected_floor_breached")
        self.assertEqual(last.action, "TAKE_FULL_PROFIT")  # floor 가 진입가 위 → net 이익 라벨(AEE:2843-2847)
        self.assertEqual(sum(f.qty for f in t.fills), ax.Q0)
        # rapid lock 이 floor 를 진입 +65% × 2000 = 11300 으로 올림 → 11300 첫 틱에서 전량, bid 11290
        self.assertEqual(last.price, 11300 - 10)


class SessionTest(unittest.TestCase):
    def test_bucket_boundaries(self):
        self.assertIsNone(ax.session_bucket(at(DAY, 14, 49)))
        self.assertEqual(ax.session_bucket(at(DAY, 14, 50)), "LATE_SESSION_PROTECT")
        self.assertEqual(ax.session_bucket(at(DAY, 15, 19)), "LATE_SESSION_PROTECT")
        self.assertEqual(ax.session_bucket(at(DAY, 15, 20)), "CLOSE_MANAGEMENT")
        self.assertEqual(ax.session_bucket(at(DAY, 15, 30)), "CLOSE_MANAGEMENT")

    def test_policy_overrides_and_tp2_disabled_after(self):
        base = ax.runtime_policy(CFG)
        self.assertEqual(base.first_take_profit_pct, 1.2)
        late = ax.effective_policy(CFG, at(DAY, 14, 50))
        self.assertEqual((late.first_take_profit_pct, late.trailing_stop_gap_pct, late.breakeven_atr_multiplier),
                         (0.8, 0.45, 0.0))
        self.assertEqual(late.second_take_profit_pct, ax.RUNNER_DISABLED_SECOND_TP_PCT)
        close = ax.effective_policy(CFG, at(DAY, 15, 20))
        self.assertEqual((close.first_take_profit_pct, close.partial_exit_ratio), (0.4, 1.0))
        self.assertEqual(close.second_take_profit_pct, ax.RUNNER_DISABLED_SECOND_TP_PCT)
        self.assertEqual(ax.effective_policy(CFG, at(DAY, 14, 49)).first_take_profit_pct, 1.2)

    def test_late_bucket_takes_profit_at_08(self):
        e = entry(minute=(14, 48))
        bars = minute_bars(at(DAY, 14, 48), [(10000, 10000, 10000, 10000, 1), (10000, 10000, 10000, 10000, 1),
                                             (10000, 10090, 10000, 10090, 1)])
        t = ax.simulate(e, bars, CFG, mode="live")
        self.assertEqual(t.fills[0].action, "TAKE_PARTIAL_PROFIT")
        self.assertEqual(t.fills[0].time, at(DAY, 14, 50))
        self.assertEqual(t.fills[0].price, 10080 - 10)


class GraceAndEodTest(unittest.TestCase):
    def test_grace_holds_small_drop_in_entry_minute(self):
        bars = minute_bars(at(DAY, 10, 0), [(10000, 10000, 9860, 9860, 1)])
        t = ax.simulate(entry(), bars, CFG, mode="live")
        self.assertEqual(t.exit_reason, "EOD")

    def test_grace_catastrophe_fires_immediately(self):
        bars = minute_bars(at(DAY, 10, 0), [(10000, 10000, 9840, 9840, 1)])
        t = ax.simulate(entry(), bars, CFG, mode="live")
        self.assertEqual(t.fills[0].action, "STOP_LOSS")
        self.assertEqual(t.fills[0].price, 9850 - 10)

    def test_structural_stop_is_technical_stop(self):
        bars = minute_bars(at(DAY, 10, 0), [(10000, 10000, 10000, 10000, 1), (10000, 10000, 9940, 9950, 1)])
        t = ax.simulate(entry(stop=9950), bars, CFG, mode="live")
        self.assertEqual(t.fills[0].price, 9940 - 10)  # 9950 아래 첫 틱

    def test_eod_exit_at_last_close_bid(self):
        bars = minute_bars(at(DAY, 10, 0), [(10000, 10050, 9990, 10010, 1)] * 3)
        t = ax.simulate(entry(), bars, CFG, mode="live")
        self.assertEqual(t.exit_reason, "EOD")
        self.assertEqual(t.fills[-1].price, 10010 - 10)

    def test_eod_flatten_setting(self):
        e = entry(minute=(15, 19))
        bars = minute_bars(at(DAY, 15, 19), [(10000, 10000, 10000, 10000, 1), (10000, 10000, 10000, 10000, 1)])
        cfg = sim_config.validate({"eod_flatten_enabled": True})
        t = ax.simulate(e, bars, cfg, mode="live")
        self.assertEqual(t.fills[0].action, "EOD_FLATTEN")
        self.assertEqual(t.fills[0].time, at(DAY, 15, 20))


class ResultTest(unittest.TestCase):
    def test_r_uses_nearest_protective_level(self):
        # 기본 neutral 하드 −1.4% → 9860; 구조 손절 9000 은 더 멀다 → R% = 1.4
        self.assertAlmostEqual(ax.risk_pct(entry(stop=9000), CFG), 1.4)
        self.assertAlmostEqual(ax.risk_pct(entry(stop=9950), CFG), 0.5)

    def test_net_includes_stockos_costs(self):
        bars = minute_bars(at(DAY, 10, 0), [(10000, 10000, 10000, 10000, 1)])
        t = ax.simulate(entry(), bars, CFG, mode="live")
        # 매도 9990(bid): (9990−10000) − 수수료 0.015%×2 근사 − 세 0.20%
        expected = (9990 - 10000 - 10000 * 0.00015 - 9990 * 0.00015 - 9990 * 0.0020) / 10000 * 100
        self.assertAlmostEqual(t.net_pct, expected)
        self.assertAlmostEqual(t.r, expected / 1.4)

    def test_as_designed_mode_evaluates_close_only(self):
        # 같은 봉을 종가 한 번으로 평가: 종가 +1.0% < TP1, 봉 저가 −1.5% 는 안 봄 → 보유 → EOD
        bars = minute_bars(at(DAY, 10, 0), [(10000, 10000, 10000, 10000, 1), (10000, 10200, 9850, 10100, 1)])
        t = ax.simulate(entry(), bars, CFG, mode="designed")
        self.assertEqual(t.exit_reason, "EOD")
        self.assertEqual(t.fills[0].price, 10100 - 10)



class ReviewFollowUpExitTest(unittest.TestCase):
    """리뷰(2026-11-27) 지적: §4.2 단계 중 시험이 없던 자리."""

    def test_no_second_partial_after_tp1(self):
        bars = minute_bars(at(DAY, 10, 0), [(10000, 10000, 10000, 10000, 1), (10000, 10150, 10000, 10150, 1),
                                            (10150, 10160, 10130, 10150, 1)])
        t = ax.simulate(entry(), bars, CFG, mode="live")
        self.assertEqual([f.action for f in t.fills].count("TAKE_PARTIAL_PROFIT"), 1)

    def test_ladder_policy_takes_rung_partial(self):
        cfg = sim_config.validate({"common_exit_policy": "COMMON_LADDER_BALANCED"})
        bars = minute_bars(at(DAY, 10, 0), [(10000, 10000, 10000, 10000, 1), (10000, 10260, 10000, 10260, 1)])
        t = ax.simulate(entry(), bars, cfg, mode="live")
        self.assertEqual(t.fills[0].action, "LADDER_PARTIAL")  # BALANCED 2 단(+2.5%, 25%)
        self.assertEqual(t.fills[0].qty, ax.Q0 // 4)

    def test_stop_confirm_seconds_delays_technical_stop(self):
        bars = minute_bars(at(DAY, 10, 0), [(10000, 10000, 10000, 10000, 1), (9940, 9940, 9940, 9940, 1),
                                            (9940, 9940, 9940, 9940, 1)])
        immediate = ax.simulate(entry(stop=9950), bars, CFG, mode="live")
        self.assertEqual(immediate.fills[0].time, at(DAY, 10, 1))
        delayed = ax.simulate(entry(stop=9950), bars, sim_config.validate({"exit_stop_confirm_seconds": 30.0}), mode="live")
        self.assertEqual(delayed.fills[0].time, at(DAY, 10, 2))  # 하회 지속 60초 ≥ 30초

    def test_atr_breakeven_raises_floor_then_breach(self):
        # 진입 ATR 1% · k 1.0: MFE ≥ 1% 에서 floor = max(손익분기, 고점 − 1%) → 그 아래로 내려오면 보호 floor 이탈
        bars = minute_bars(at(DAY, 10, 0), [(10000, 10000, 10000, 10000, 1), (10000, 10110, 10000, 10110, 1),
                                            (10110, 10110, 10020, 10020, 1)])
        cfg = sim_config.validate({"krx_first_take_profit_pct": 5.0, "krx_second_take_profit_pct": 5.0})
        t = ax.simulate(entry(atr=1.0), bars, cfg, mode="live")
        self.assertEqual(t.fills[0].reason, "protected_floor_breached")

    def test_sell_tick_below_band_boundary(self):
        bars = minute_bars(at(DAY, 10, 0), [(20000, 20000, 20000, 20000, 1)])
        t = ax.simulate(entry(price=20000, stop=19000), bars, CFG, mode="live")
        self.assertEqual(t.fills[-1].price, 19990)


class ReviewFollowUpExitTest2(unittest.TestCase):
    def test_ladder_state_persists_between_ticks(self):
        # BALANCED 1 단(+1.5%) 도달 → 보호 손절 0% 로 승격, 그 뒤 진입가 아래 첫 틱에서 사다리 손절
        cfg = sim_config.validate({"common_exit_policy": "COMMON_LADDER_BALANCED"})
        bars = minute_bars(at(DAY, 10, 0), [(10000, 10000, 10000, 10000, 1), (10000, 10160, 10000, 10160, 1),
                                            (10160, 10160, 9990, 9990, 1)])
        t = ax.simulate(entry(), bars, cfg, mode="live")
        self.assertEqual(t.fills[0].action, "STOP_LOSS_LADDER")
        self.assertEqual(t.fills[0].price, 9990 - 10)

    def test_floor_breach_waits_for_confirm_seconds(self):
        bars = minute_bars(at(DAY, 10, 0), [(10000, 10000, 10000, 10000, 1), (10000, 10110, 10000, 10110, 1),
                                            (10110, 10110, 10020, 10020, 1), (10020, 10020, 10020, 10020, 1)])
        cfg = sim_config.validate({"krx_first_take_profit_pct": 5.0, "krx_second_take_profit_pct": 5.0,
                                   "exit_stop_confirm_seconds": 30.0})
        t = ax.simulate(entry(atr=1.0), bars, cfg, mode="live")
        self.assertEqual(t.fills[0].reason, "protected_floor_breached")
        self.assertEqual(t.fills[0].time, at(DAY, 10, 3))  # 하회 시작 10:02, 60초 뒤 확인


if __name__ == "__main__":
    unittest.main()

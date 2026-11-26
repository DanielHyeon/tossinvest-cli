"""C1 진입 규칙 시험 — design §2 의 숫자 하나하나를 경계에서 잼."""

from __future__ import annotations

import unittest
from datetime import timedelta

import c1_entry as ce
from tests.helpers import DAY, PREV, at, expand3, flat_session, rising_prev, setup_day


class AggregationTest(unittest.TestCase):
    def test_3m_bucket_rebuilds_ohlcv(self):
        bars = expand3(DAY, 0, 100, 130, 90, 120, 900)
        (b,) = ce.aggregate(bars, 3)
        self.assertEqual((b.open, b.high, b.low, b.close, b.volume), (100, 130, 90, 120, 900))
        self.assertEqual(b.start, at(DAY, 9, 0))
        self.assertEqual(b.end, at(DAY, 9, 3))

    def test_30m_bar_invisible_before_its_close(self):
        bars = flat_session(DAY, 10000, 10000)
        b30 = ce.aggregate(bars, 30)
        # 09:29 에 닫히는 1분봉이 끝나는 09:30 전에는 첫 30분 봉을 쓸 수 없음
        self.assertEqual(ce.closed(b30, at(DAY, 9, 29)), [])
        self.assertEqual(len(ce.closed(b30, at(DAY, 9, 30))), 1)
        self.assertEqual(len(ce.closed(b30, at(DAY, 10, 0) - timedelta(seconds=1))), 1)

    def test_3m_bar_invisible_before_its_close(self):
        b3 = ce.aggregate(flat_session(DAY, 10000, 10000), 3)
        self.assertEqual(ce.closed(b3, at(DAY, 9, 2)), [])
        self.assertEqual(len(ce.closed(b3, at(DAY, 9, 3))), 1)

    def test_closing_print_joins_last_bucket(self):
        bars = flat_session(DAY, 10000, 10000) + [ce.Bar(at(DAY, 15, 30), 1, 2, 1, 2, 5)]
        b30 = ce.aggregate(bars, 30)
        self.assertEqual(len(b30), 13)
        self.assertEqual(b30[-1].close, 2)


class EmaAndZTest(unittest.TestCase):
    def test_ema_hand_fixture(self):
        # α = 2/(3+1) = 0.5 : 10 → 10.5 → 11.25
        self.assertAlmostEqual(ce.ema([10, 11, 12], 3), 11.25)
        # α = 1/3 : 3 → 3 + (6-3)/3 = 4 → 4 + (9-4)/3 = 5.666…
        self.assertAlmostEqual(ce.ema([3, 6, 9], 5), 17 / 3)

    def test_macro_needs_15_closed_bars(self):
        bars = [ce.Bar(at(DAY, 9, 0) + timedelta(minutes=30 * i), p, p, p, p, 1, 30) for i, p in enumerate(range(100, 114))]
        self.assertFalse(ce.macro_ok(bars))  # 14 개
        bars.append(ce.Bar(at(DAY, 16, 0), 200, 200, 200, 200, 1, 30))
        self.assertTrue(ce.macro_ok(bars))  # 15 개, 상승
        self.assertFalse(ce.macro_ok([ce.Bar(b.start, 300 - b.close, 0, 0, 300 - b.close, 1, 30) for b in bars]))

    def test_macro_uses_only_last_15(self):
        rising = [ce.Bar(at(DAY, 9, 0), p, p, p, p, 1, 30) for p in range(100, 115)]
        falling_then = [ce.Bar(at(DAY, 9, 0), p, p, p, p, 1, 30) for p in range(200, 100, -1)]
        self.assertTrue(ce.macro_ok(falling_then + rising))

    def test_zscore_edges(self):
        self.assertIsNone(ce.volume_z(500, [100] * 20))  # 모표준편차 0 → 신호 없음
        prev = [3000, 3600] * 10  # 평균 3300, 모표준편차 300
        self.assertAlmostEqual(ce.volume_z(3900, prev), 2.0)
        self.assertAlmostEqual(ce.volume_z(3899, prev), 599 / 300)


class BreakoutBarTest(unittest.TestCase):
    prev = [ce.Bar(at(DAY, 9, 0), 10000, 10050, 9950, 10000, v, 3) for v in [3000, 3600] * 10]

    def bar(self, o, h, l, c, v):
        return ce.Bar(at(DAY, 10, 0), o, h, l, c, v, 3)

    def test_valid_breakout_returns_level(self):
        self.assertEqual(ce.breakout_level(self.bar(10000, 10120, 9990, 10100, 3900), self.prev), 10050)

    def test_each_condition_rejects(self):
        self.assertIsNone(ce.breakout_level(self.bar(10000, 10120, 9990, 10050, 3900), self.prev))  # 종가 = L
        self.assertIsNone(ce.breakout_level(self.bar(10100, 10120, 9990, 10060, 3900), self.prev))  # 음봉
        self.assertIsNone(ce.breakout_level(self.bar(10000, 10201, 9990, 10100, 3900), self.prev))  # 윗꼬리 101 > 몸통 100
        self.assertEqual(ce.breakout_level(self.bar(10000, 10200, 9990, 10100, 3900), self.prev), 10050)  # 꼬리 = 몸통
        self.assertIsNone(ce.breakout_level(self.bar(10000, 10120, 9990, 10100, 3899), self.prev))  # z < 2
        self.assertIsNone(ce.breakout_level(self.bar(10000, 10120, 9990, 10100, 3900), self.prev[1:]))  # 앞 19 봉


class EntryScanTest(unittest.TestCase):
    def run_scan(self, today, prior_close=9900, history=None):
        return ce.find_entry(today, rising_prev() if history is None else history, prior_close)

    def test_entry_at_next_minute_open(self):
        entry, funnel = self.run_scan(setup_day())
        self.assertIsNotNone(entry)
        self.assertEqual(entry.time, at(DAY, 10, 6))
        self.assertEqual(entry.price, 10100)
        self.assertEqual(entry.level, 10050)
        self.assertEqual(entry.stop, 10060 - 10)  # 되돌림 저가 − 1틱
        self.assertEqual(funnel["entry"], 1)

    def test_retest_band_boundaries(self):
        # L=10050: L×0.997=10019.85, L×1.003=10080.15
        _, f = self.run_scan(setup_day(pullbacks=((10070, 10110, 10019, 10100, 3300),)))
        self.assertEqual(f["pullback_failed"], 1)  # 너무 깊음
        e, _ = self.run_scan(setup_day(pullbacks=((10070, 10110, 10020, 10100, 3300),)))
        self.assertIsNotNone(e)
        e, _ = self.run_scan(setup_day(pullbacks=((10090, 10110, 10080, 10100, 3300),)))
        self.assertIsNotNone(e)
        _, f = self.run_scan(setup_day(pullbacks=((10090, 10110, 10090, 10100, 3300),) * 10))
        self.assertEqual(f["timeout"], 1)  # 10080.15 아래로 안 옴

    def test_only_first_pullback_counts(self):
        bad_then_good = ((10100, 10110, 10060, 10070, 3300), (10070, 10110, 10060, 10100, 3300))
        entry, f = self.run_scan(setup_day(pullbacks=bad_then_good))
        self.assertIsNone(entry)
        self.assertEqual(f["pullback_failed"], 1)

    def test_close_must_exceed_level(self):
        _, f = self.run_scan(setup_day(pullbacks=((10040, 10110, 10030, 10050, 3300),)))
        self.assertEqual(f["pullback_failed"], 1)

    def test_timeout_after_10_bars(self):
        high_bars = ((10100, 10120, 10090, 10100, 3300),) * 10
        _, f = self.run_scan(setup_day(pullbacks=high_bars + ((10070, 10110, 10060, 10100, 3300),)))
        self.assertEqual(f["timeout"], 1)
        e, _ = self.run_scan(setup_day(pullbacks=high_bars[:9] + ((10070, 10110, 10060, 10100, 3300),)))
        self.assertIsNotNone(e)  # 10 번째 봉은 아직 창 안

    def test_surge_exclusion_boundary(self):
        self.assertTrue(ce.surge_excluded(10800, 10000))
        self.assertFalse(ce.surge_excluded(10799, 10000))
        _, f = self.run_scan(setup_day(), prior_close=9350)  # 10100/9350 = +8.02%
        self.assertEqual(f["surge_rejected"], 1)
        e, _ = self.run_scan(setup_day(), prior_close=9352)  # +7.99%
        self.assertIsNotNone(e)

    def test_macro_reject(self):
        _, f = self.run_scan(setup_day(), history=flat_session(PREV, 11500, 10100))
        self.assertEqual(f["macro_rejected"], 1)

    def test_macro_without_history_has_too_few_bars(self):
        _, f = self.run_scan(setup_day(), history=[])
        self.assertEqual(f["macro_rejected"], 1)

    def test_1445_cutoff(self):
        # 되돌림 봉 끝 S = 09:00 + 3×(rise+20+2) 분. rise=93 → 14:45, 94 → 14:48
        e, _ = self.run_scan(setup_day(rise_bars=93, tail_minutes=30))
        self.assertIsNotNone(e)
        self.assertEqual(e.time, at(DAY, 14, 45))
        e, f = self.run_scan(setup_day(rise_bars=94, tail_minutes=30))
        self.assertIsNone(e)
        self.assertEqual(f["cutoff"], 1)

    def test_missing_entry_minute(self):
        today = [b for b in setup_day() if b.start != at(DAY, 10, 6)]
        e, f = self.run_scan(today)
        self.assertIsNone(e)
        self.assertEqual(f["no_entry_bar"], 1)

    def test_one_entry_per_session(self):
        e, f = self.run_scan(setup_day())
        self.assertEqual(f["entry"], 1)

    def test_rearm_after_failed_setup(self):
        # 실패한 첫 setup 뒤, 새 고점으로 두 번째 돌파 → 그 되돌림에서 진입
        bars = setup_day(pullbacks=((10100, 10110, 10060, 10070, 3300),), tail_minutes=0)
        k = (len(bars)) // 3
        p = 10070
        for i in range(19):
            bars += expand3(DAY, k + i, p, p + 50, p - 50, p, 3000 if i % 2 == 0 else 3600)
        k += 19
        bars += expand3(DAY, k, 10100, 10250, 10090, 10230, 9000)  # 앞 20 최고가 10120(첫 돌파 봉) 돌파
        bars += expand3(DAY, k + 1, 10140, 10180, 10120, 10160, 3300)
        bars.append(ce.Bar(at(DAY, 9, 0) + timedelta(minutes=3 * (k + 2)), 10160, 10160, 10160, 10160, 1))
        e, f = self.run_scan(bars)
        self.assertEqual(f["pullback_failed"], 1)
        self.assertIsNotNone(e)
        self.assertEqual(e.level, 10120)


class AtrAndTickTest(unittest.TestCase):
    def test_kr_tick_table(self):
        for price, tick in [(1999, 1), (2000, 5), (4999, 5), (5000, 10), (19999, 10), (20000, 50),
                            (49999, 50), (50000, 100), (199999, 100), (200000, 500), (499999, 500), (500000, 1000)]:
            self.assertEqual(ce.kr_tick(price), tick, price)

    def test_atr_pct_needs_15_closed_5m_bars(self):
        bars = flat_session(DAY, 10000, 10000)  # 5분 봉 TR = 100 (고 +50 저 −50)
        self.assertIsNone(ce.atr_pct_5m(bars, at(DAY, 10, 10)))  # 닫힌 5분 봉 14 개
        self.assertAlmostEqual(ce.atr_pct_5m(bars, at(DAY, 10, 15)), 1.0)



class ReviewFollowUpEntryTest(unittest.TestCase):
    """리뷰(2026-11-27) 지적: 규칙이 시험에 못 박히지 않았던 자리."""

    def b30(self, closes):
        return [ce.Bar(at(DAY, 9, 0) + timedelta(minutes=30 * i), p, p, p, p, 1, 30) for i, p in enumerate(closes)]

    def test_macro_strict_greater_and_last_15_only(self):
        self.assertFalse(ce.macro_ok(self.b30([100] * 15)))  # EMA5 == EMA15 → 거래 없음
        # 전체로는 EMA5 < EMA15 지만 마지막 15 개만 보면 EMA5 > EMA15
        closes = [1000] * 30 + list(range(100, 115))
        self.assertGreater(ce.ema(closes[-15:], 5), ce.ema(closes[-15:], 15))
        self.assertLess(ce.ema(closes, 5), ce.ema(closes, 15))
        self.assertTrue(ce.macro_ok(self.b30(closes)))

    def exact_day(self, retest):
        # 앞 20 봉 최고가 10000 → L = 10000, 띠 9970 ~ 10030 이 정수
        bars = []
        for k in range(20):
            bars += expand3(DAY, k, 9950, 10000, 9900, 9950, 3000 if k % 2 == 0 else 3600)
        bars += expand3(DAY, 20, 9950, 10120, 9940, 10100, 9000)
        bars += expand3(DAY, 21, *retest)
        bars.append(ce.Bar(at(DAY, 10, 6), 10050, 10050, 10050, 10050, 1))
        return bars

    def test_band_edges_exact(self):
        for low, ok in ((9970, True), (9969, False), (10030, True)):
            e, f = ce.find_entry(self.exact_day((10040, 10060, low, 10050, 3300)), rising_prev(), 9900)
            self.assertEqual(e is not None, ok, low)
        _, f = ce.find_entry(self.exact_day((10040, 10060, 10031, 10050, 3300)), rising_prev(), 9900)
        self.assertEqual(f["timeout"], 1)  # 10031 > 10030 → 되돌림 아님

    def test_breakout_inside_live_setup_is_ignored(self):
        # 시한 창(10 봉) 안의 두 번째 돌파는 세지 않음 → 돌파 1, 시한 1
        hi = ((10100, 10120, 10090, 10100, 3300),) * 4 + ((10100, 10400, 10095, 10380, 20000),) + ((10380, 10400, 10370, 10390, 3300),) * 5
        _, f = ce.find_entry(setup_day(pullbacks=hi, tail_minutes=0), rising_prev(), 9900)
        self.assertEqual(f["breakout"], 1)
        self.assertEqual(f["timeout"], 1)

    def test_stop_fallback_when_retest_low_not_below_entry(self):
        today = [b for b in setup_day() if b.start != at(DAY, 10, 6)] + [ce.Bar(at(DAY, 10, 6), 10050, 10050, 10050, 10050, 1)]
        e, _ = ce.find_entry(sorted(today, key=lambda b: b.start), rising_prev(), 9900)
        self.assertAlmostEqual(e.stop, 10050 * 0.97)  # 되돌림 저가 − 1틱 = 10050 ≥ 진입가

    def test_stop_one_tick_below_band_boundary(self):
        self.assertEqual(20000 - ce.kr_tick(20000 - 1), 19990)

    def test_atr_uses_previous_close_gap(self):
        bars = flat_session(DAY, 10000, 10000)
        # 5분 봉 하나를 위로 갭: 이전 종가 대비 TR 이 고가−저가보다 큼
        gapped = [ce.Bar(b.start, b.open + 300, b.high + 300, b.low + 300, b.close + 300, b.volume)
                  if at(DAY, 10, 10) <= b.start < at(DAY, 10, 15) else b for b in bars]
        # 마지막 봉 TR = |저가 10250 − 이전 종가 10000|·|고가 10350 − 10000| 중 큰 350(고가−저가 100 이 아님)
        self.assertAlmostEqual(ce.atr_pct_5m(gapped, at(DAY, 10, 15)), (13 * 100 + 350) / 14 / 10300 * 100)


if __name__ == "__main__":
    unittest.main()

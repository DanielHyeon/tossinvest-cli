"""판정 시험 — design §7 의 각 조건을 경계에서 넘기고 안 넘김."""

from __future__ import annotations

import unittest
from datetime import date, datetime, timedelta

import verdict as vd
from c1_entry import KST

SESSIONS = [date(2026, 1, 1) + timedelta(days=i) for i in range(100)]
DEV, HOLD = SESSIONS[:50], SESSIONS[50:]


def rows(sessions, rs):
    out = []
    for i, (d, r) in enumerate(zip(sessions, rs)):
        out.append(vd.Row(session=d, time=datetime(d.year, d.month, d.day, 10, 0, tzinfo=KST), symbol=f"{i:06d}", r=r))
    return out


def passing_holdout():
    # 22 승(+1) · 18 패(−1) 교대 → 평균 0.10 · 양의 세션 55% · 연속 손실 1 · 낙폭 1
    rs = [1.0, -1.0] * 18 + [1.0] * 4
    return rows(HOLD, rs)


def run(hold, dev=None, gaps=0, sample=SESSIONS):
    dev = rows(DEV, [0.5]) if dev is None else dev
    universe = {d: 30 for d in sample}
    gap_map = {sample[-1]: gaps} if gaps else {}
    return vd.evaluate(sample, dev + hold, universe, gap_map)


class VerdictTest(unittest.TestCase):
    def test_pass_at_every_boundary(self):
        res = run(passing_holdout())
        self.assertEqual(res["verdict"], "PASS", res)
        self.assertEqual(res["holdout"]["trades"], 40)
        self.assertAlmostEqual(res["holdout"]["mean_r"], 0.10)
        self.assertAlmostEqual(res["holdout"]["positive_session_ratio"], 0.55)
        self.assertEqual(res["score_e"], 40)

    def test_mean_just_below(self):
        h = passing_holdout()
        h[-1] = h[-1]._replace(r=0.996)  # 평균 0.0999
        res = run(h)
        self.assertEqual(res["verdict"], "FAIL")
        self.assertFalse(res["checks"]["c1_mean_r"])
        self.assertEqual(res["score_e"], 24)

    def test_positive_session_ratio_below(self):
        rs = [2.0, -1.0] * 18 + [2.0] + [-1.0] * 3  # 19/40 양
        res = run(rows(HOLD, rs))
        self.assertFalse(res["checks"]["c3_positive_sessions"])
        self.assertEqual(res["verdict"], "FAIL")
        self.assertEqual(res["score_e"], 30)

    def test_session_ratio_uses_daily_mean(self):
        h = passing_holdout()
        # 같은 세션에 +1 · −1 → 일 평균 0 은 양이 아님
        same_day = [h[0], h[1]._replace(session=h[0].session)]
        res = vd.metrics(same_day)
        self.assertEqual(res["sessions_with_trades"], 1)
        self.assertEqual(res["positive_session_ratio"], 0.0)

    def test_consecutive_losses(self):
        rs6 = [3.0] * 20 + [-1.0] * 6 + [1.0] * 14
        self.assertEqual(run(rows(HOLD, rs6))["verdict"], "PASS")
        rs7 = [3.0] * 20 + [-1.0] * 7 + [1.0] * 13
        res = run(rows(HOLD, rs7))
        self.assertFalse(res["checks"]["c4_losses_and_drawdown"])
        self.assertEqual(res["score_e"], 32)

    def test_drawdown_boundary(self):
        rs = [3.0] * 22 + [-4.0, 1.0, -5.0] + [0.5] * 15  # 낙폭 4 − 1 + 5 = 8
        res = run(rows(HOLD, rs))
        self.assertAlmostEqual(res["holdout"]["max_drawdown_r"], 8.0)
        self.assertTrue(res["checks"]["c4_losses_and_drawdown"])
        rs[24] = -5.01
        self.assertFalse(run(rows(HOLD, rs))["checks"]["c4_losses_and_drawdown"])

    def test_development_sign(self):
        self.assertFalse(run(passing_holdout(), dev=rows(DEV, [-0.1]))["checks"]["c5_dev_same_sign"])
        res = run(passing_holdout(), dev=[])
        self.assertFalse(res["checks"]["c5_dev_same_sign"])
        self.assertEqual(res["verdict"], "FAIL")
        self.assertEqual(run(passing_holdout(), dev=rows(DEV, [0.0]))["verdict"], "FAIL")

    def test_insufficient_trades(self):
        res = run(passing_holdout()[:39])
        self.assertEqual(res["verdict"], "INSUFFICIENT")
        self.assertEqual(res["score_e"], 0)

    def test_insufficient_gaps(self):
        self.assertEqual(run(passing_holdout(), gaps=300)["verdict"], "PASS")  # 300/1500 = 20%
        self.assertEqual(run(passing_holdout(), gaps=301)["verdict"], "INSUFFICIENT")

    def test_dev_gaps_do_not_count(self):
        universe = {d: 30 for d in SESSIONS}
        res = vd.evaluate(SESSIONS, rows(DEV, [0.5]) + passing_holdout(), universe, {DEV[0]: 30})
        self.assertEqual(res["verdict"], "PASS")

    def test_insufficient_small_sample(self):
        sample = HOLD
        res = vd.evaluate(sample, passing_holdout(), {d: 30 for d in sample}, {})
        self.assertEqual(res["verdict"], "INSUFFICIENT")

    def test_holdout_is_last_50(self):
        res = run(passing_holdout())
        self.assertEqual(res["split"]["holdout_first"], HOLD[0].isoformat())
        self.assertEqual(res["split"]["development_sessions"], 50)



class ReviewFollowUpVerdictTest(unittest.TestCase):
    def test_both_negative_is_same_sign(self):
        hold = rows(HOLD, [-1.0] * 40)
        self.assertTrue(run(hold, dev=rows(DEV, [-0.5]))["checks"]["c5_dev_same_sign"])

    def test_positive_ratio_just_below_55(self):
        rs = [3.0] * 21 + [-1.0] * 19  # 21/40 = 0.525 (c4 도 실패하지만 c3 만 봄)
        res = run(rows(HOLD, rs))
        self.assertAlmostEqual(res["holdout"]["positive_session_ratio"], 0.525)
        self.assertFalse(res["checks"]["c3_positive_sessions"])

    def test_duplicate_sample_dates_do_not_shrink_holdout(self):
        res = vd.evaluate(SESSIONS + SESSIONS[-10:], rows(DEV, [0.5]) + passing_holdout(), {d: 30 for d in SESSIONS}, {})
        self.assertEqual(res["split"]["holdout_sessions"], 50)
        self.assertEqual(res["verdict"], "PASS")

    def test_reports_net_and_eod(self):
        res = vd.metrics([vd.Row(DEV[0], datetime(2026, 1, 1, 10, tzinfo=KST), "1", 1.0, 0.3, "EOD")])
        self.assertEqual((res["mean_net_pct"], res["eod_exits"]), (0.3, 1))


if __name__ == "__main__":
    unittest.main()

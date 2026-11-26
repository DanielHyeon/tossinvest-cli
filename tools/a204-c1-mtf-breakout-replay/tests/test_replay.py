"""CLI 종단 시험 — 픽스처 형식 읽기 · 보고서의 설정 digest · 종료 코드 2 · 덮어쓰기 거절 · 결손 셈."""

from __future__ import annotations

import io
import json
import tempfile
import unittest
from contextlib import redirect_stderr
from pathlib import Path

import replay
import sim_config
from tests.helpers import DAY, PREV, rising_prev, setup_day


def bars_json(bars):
    return [[b.start.strftime("%H:%M"), b.open, b.high, b.low, b.close, b.volume] for b in bars]


def write_fixture(root: Path, *, missing_symbol=False):
    days = [PREV.isoformat(), DAY.isoformat()]
    manifest = {
        "format": "a204-fixture/v1",
        "sessions": days,
        "sample": [DAY.isoformat()],
        "universe": {DAY.isoformat(): ["000001", "000002"]},
    }
    (root / "manifest.json").write_text(json.dumps(manifest))
    for sym in ("000001", "000002"):
        if sym == "000002" and missing_symbol:
            continue
        d = root / "bars" / sym
        d.mkdir(parents=True)
        (d / f"{PREV.isoformat()}.json").write_text(json.dumps({"prior_close": 9000, "bars": bars_json(rising_prev())}))
        (d / f"{DAY.isoformat()}.json").write_text(json.dumps({"prior_close": 9900, "bars": bars_json(setup_day())}))


class ReplayTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.root = Path(self.tmp.name)
        self.src = self.root / "src"
        self.src.mkdir()

    def tearDown(self):
        self.tmp.cleanup()

    def test_report_contains_digest_and_trades(self):
        write_fixture(self.src)
        out = self.root / "report.json"
        self.assertEqual(replay.main(["--source", str(self.src), "--output", str(out)]), 0)
        rep = json.loads(out.read_text())
        self.assertEqual(rep["config_digest"], sim_config.digest(sim_config.validate({})))
        self.assertEqual(rep["stockos_commit"], "fcad869ecc531a6de80777f59b6fb842a2891101")
        self.assertTrue(rep["limitations"][0].startswith("과거 호가 없음"))
        self.assertEqual(rep["modes"]["live-effective"]["funnel"]["entry"], 2)
        self.assertEqual(len(rep["modes"]["live-effective"]["trades"]), 2)
        self.assertEqual(rep["verdict_mode"], "live-effective")
        self.assertEqual(rep["verdict"]["verdict"], "INSUFFICIENT")  # 표본 1 세션
        self.assertIn("as-designed", rep["modes"])

    def test_config_error_exit_code_2(self):
        write_fixture(self.src)
        cfg = self.root / "c.json"
        cfg.write_text(json.dumps({"exit_trail_stop_pct": 99}))
        err = io.StringIO()
        with redirect_stderr(err):
            code = replay.main(["--source", str(self.src), "--config", str(cfg), "--output", str(self.root / "r.json")])
        self.assertEqual(code, 2)
        self.assertIn("exit_trail_stop_pct", err.getvalue())
        self.assertFalse((self.root / "r.json").exists())

    def test_refuses_existing_output(self):
        write_fixture(self.src)
        out = self.root / "r.json"
        out.write_text("keep")
        with redirect_stderr(io.StringIO()):
            self.assertEqual(replay.main(["--source", str(self.src), "--output", str(out)]), 2)
        self.assertEqual(out.read_text(), "keep")

    def test_missing_symbol_counts_as_gap(self):
        write_fixture(self.src, missing_symbol=True)
        out = self.root / "r.json"
        self.assertEqual(replay.main(["--source", str(self.src), "--output", str(out)]), 0)
        rep = json.loads(out.read_text())
        self.assertEqual(rep["gaps"], {DAY.isoformat(): 1})
        self.assertEqual(rep["modes"]["live-effective"]["funnel"]["entry"], 1)

    def test_bad_manifest_format(self):
        (self.src / "manifest.json").write_text(json.dumps({"format": "other"}))
        with redirect_stderr(io.StringIO()):
            self.assertEqual(replay.main(["--source", str(self.src), "--output", str(self.root / "r.json")]), 2)



class ReviewFollowUpReplayTest(unittest.TestCase):
    """리뷰(2026-11-27) 지적: 어댑터 경계의 입력 검증 · 결손 셈 · 입력 지문."""

    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.root = Path(self.tmp.name)
        self.src = self.root / "src"
        self.src.mkdir()
        write_fixture(self.src)

    def tearDown(self):
        self.tmp.cleanup()

    def edit_manifest(self, **changes):
        p = self.src / "manifest.json"
        m = json.loads(p.read_text())
        m.update(changes)
        p.write_text(json.dumps(m))

    def edit_day(self, fn, sym="000001", day=DAY):
        p = self.src / "bars" / sym / f"{day.isoformat()}.json"
        data = json.loads(p.read_text())
        fn(data)
        p.write_text(json.dumps(data))

    def code(self, name="r.json"):
        err = io.StringIO()
        with redirect_stderr(err):
            code = replay.main(["--source", str(self.src), "--output", str(self.root / name)])
        return code, err.getvalue()

    def test_rejects_bad_manifest_values(self):
        cases = [
            {"sample": [DAY.isoformat(), DAY.isoformat()]},
            {"universe": {}},
            {"universe": {DAY.isoformat(): ["000001", "000001"]}},
            {"universe": {DAY.isoformat(): ["../000001"]}},
            {"universe": {DAY.isoformat(): "000001"}},
        ]
        for change in cases:
            with self.subTest(change=change):
                original = (self.src / "manifest.json").read_text()
                self.edit_manifest(**change)
                self.assertEqual(self.code()[0], 2)
                (self.src / "manifest.json").write_text(original)

    def test_rejects_bad_bar_rows(self):
        def nan(d):
            d["bars"][0][4] = float("nan")

        def dup(d):
            d["bars"].insert(1, d["bars"][0])

        def ohlc(d):
            d["bars"][0][2] = d["bars"][0][3] - 1  # 고가 < 저가

        def short(d):
            d["bars"][0] = d["bars"][0][:4]

        for fn in (nan, dup, ohlc, short):
            with self.subTest(fn=fn.__name__):
                p = self.src / "bars" / "000001" / f"{DAY.isoformat()}.json"
                original = p.read_text()
                self.edit_day(fn)
                code, err = self.code()
                self.assertEqual(code, 2)
                self.assertIn("000001", err)
                p.write_text(original)

    def test_missing_previous_session_is_a_gap_not_macro_reject(self):
        (self.src / "bars" / "000002" / f"{PREV.isoformat()}.json").unlink()
        self.assertEqual(self.code()[0], 0)
        rep = json.loads((self.root / "r.json").read_text())
        self.assertEqual(rep["gaps"], {DAY.isoformat(): 1})
        self.assertNotIn("macro_rejected", rep["modes"]["live-effective"]["funnel"])

    def test_auction_minutes_are_dropped(self):
        def add(d):
            d["bars"] += [["15:25", 1, 1, 1, 1, 1], ["15:30", 1, 1, 1, 1, 1]]
        self.edit_day(add)
        starts = [b.start.strftime("%H:%M") for b in replay.FixtureSource(self.src).minute_bars("000001", DAY)]
        self.assertNotIn("15:25", starts)
        self.assertEqual(starts[-1], "15:30")

    def test_input_fingerprint_binds_bar_bytes(self):
        self.assertEqual(self.code("a.json")[0], 0)
        self.edit_day(lambda d: d["bars"].__setitem__(0, d["bars"][0][:5] + [d["bars"][0][5] + 1]), sym="000002")
        self.assertEqual(self.code("b.json")[0], 0)
        a = json.loads((self.root / "a.json").read_text())
        b = json.loads((self.root / "b.json").read_text())
        self.assertEqual(a["sample_fingerprint"], b["sample_fingerprint"])
        self.assertNotEqual(a["input_fingerprint"], b["input_fingerprint"])

    def test_missing_output_directory_is_exit_2(self):
        self.assertEqual(self.code("no/such/dir.json")[0], 2)


if __name__ == "__main__":
    unittest.main()

import importlib.util
import hashlib
import json
import pathlib
import tempfile
import unittest

HERE = pathlib.Path(__file__).parent
spec = importlib.util.spec_from_file_location("profile_eval", HERE / "profile_eval.py")
evalmod = importlib.util.module_from_spec(spec)
spec.loader.exec_module(evalmod)

class ProfileEvalTest(unittest.TestCase):
    def test_missing_and_exact_fallback_and_later_override(self):
        parse = evalmod.prereg.parse_at
        states = {
            "early": {"market":"KR", "symbol":"X", "observed_at":parse("2026-10-01T01:01:00Z"), "bars":[{"t":"09:55", "c":100}, {"t":"14:05", "c":105}]},
            "late": {"market":"KR", "symbol":"X", "observed_at":parse("2026-10-01T05:15:00Z"), "bars":[{"t":"14:05", "c":110}]},
        }
        book=evalmod.prereg.build_bars(states)[("KR","2026-10-01","X")]
        ret, why=evalmod.prereg.kr_trade(book)
        self.assertIsNone(why); self.assertAlmostEqual(ret,.1)
        self.assertEqual(evalmod.prereg.kr_trade({"09:54":(None,100), "14:05":(None,110)})[1],"entry_missing")
        self.assertEqual(evalmod.prereg.kr_trade({"09:55":(None,100), "14:04":(None,110)})[1],"exit_missing")

    def test_later_kr_warning_and_us_first_flagged(self):
        with tempfile.TemporaryDirectory() as tmp:
            root=pathlib.Path(tmp); states=root/"states"; states.mkdir(); records=[]
            def state(market,symbol,when,minute,bars):
                raw=json.dumps({"market":market,"symbol":symbol,"observed_at":when,"minutes_since_open":minute,"intraday":{"bars_1m":bars}}).encode()
                digest=hashlib.sha256(raw).hexdigest(); (states/(digest+".json")).write_bytes(raw+b"\n"); return "sha256:"+digest
            for market,symbol,when,minute,current,initial,ret in (("KR","X","2026-10-01T00:15:00Z",15,.2,.2,.1),("KR","X","2026-10-01T00:30:00Z",30,.2,.7,.1),("US","Y","2026-10-01T14:00:00Z",30,.7,.2,.01),("US","Y","2026-10-01T14:10:00Z",40,.2,.7,.03)):
                digest=state(market,symbol,when,minute,[{"t":"10:00","c":100},{"t":"14:10","c":110}] if market=="KR" else [])
                for profile,p in (("current",current),("initial",initial)):
                    ident=market+when+profile
                    records.append({"kind":"judgment","id":ident,"cohort":"c","source_state_digest":digest,"state_digest":digest,"market":market,"symbol":symbol,"at":when,"question_id":evalmod.prereg.J2,"question_rev":"v","question_digest":"q","model_requested":"m","model_answered":"m","horizon_minutes":60,"base_price":100,"input_profile":profile,"p":p})
                    records.append({"kind":"label","judgment_id":ident,"realized_return":ret})
            ledger=root/"ledger";ledger.write_text("\n".join(json.dumps(r) for r in records))
            result=evalmod.evaluate(ledger,states)
            self.assertEqual(result["us_entry_time_difference_stockdays"],1)
            self.assertIn("first flagged",result["us_selection_rule"])
            got=result["groups"]
            self.assertIn("KR/initial/signal/미분류",got)
            self.assertIn("KR/current/non-signal/미분류",got)
            self.assertAlmostEqual(got["US/current/signal/미분류"]["gross_date_equal"],.01)
            self.assertAlmostEqual(got["US/initial/signal/미분류"]["gross_date_equal"],.03)

    def test_failed_later_source_contributes_price(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = pathlib.Path(tmp); states = root / "states"; states.mkdir()
            digests = {}
            for name, when, bars in (("source", "2026-10-01T00:15:00Z", []), ("later", "2026-10-01T05:15:00Z", [{"t":"10:00","c":100},{"t":"14:10","c":110}]), ("input", "2026-10-01T00:15:00Z", [])):
                raw = json.dumps({"market":"KR","symbol":"X","observed_at":when,"minutes_since_open":15,"intraday":{"bars_1m":bars}}).encode()
                digest=hashlib.sha256(raw).hexdigest(); digests[name]=digest
                (states / (digest + ".json")).write_bytes(raw+b"\n")
            base={"kind":"judgment","cohort":"c","source_state_digest":"sha256:"+digests["source"],"state_digest":"sha256:"+digests["input"],"market":"KR","symbol":"X","at":"2026-10-01T00:15:00Z","question_id":evalmod.prereg.J2,"question_rev":"v","question_digest":"q","model_requested":"m","model_answered":"m","horizon_minutes":60,"base_price":100}
            a=dict(base,id="a",input_profile="current",p=.59); b=dict(base,id="b",input_profile="initial",p=.61)
            gap={"kind":"gap","cohort":"c","input_profile":"initial","source_state_digest":"sha256:"+digests["later"]}
            ledger=root/"ledger.jsonl"; ledger.write_text("\n".join(json.dumps(r) for r in (a,b,gap)))
            got=evalmod.evaluate(ledger, states)
            self.assertEqual(got["paired_stockdays"],1)
            self.assertEqual(got["price_missing"],{})
            result=got["groups"]["KR/initial/signal/미분류"]
            self.assertAlmostEqual(result["gross_date_equal"],.1)
            self.assertAlmostEqual(result["net_date_equal"],.095)
            missing_cohort=dict(a,id="bad",cohort="")
            with ledger.open("a") as f: f.write("\n"+json.dumps(missing_cohort))
            self.assertEqual(evalmod.evaluate(ledger, states)["invalid_rows"],1)
            ledger.write_text("\n".join(json.dumps(r) for r in (a,b,gap)))
            (states / (digests["source"]+".json")).write_text("{}")
            invalid=evalmod.evaluate(ledger, states)
            self.assertEqual(invalid["paired_stockdays"],0)
            self.assertEqual(invalid["invalid_rows"],2)
            self.assertEqual(invalid["missing_source_states"],1)


if __name__ == "__main__": unittest.main()

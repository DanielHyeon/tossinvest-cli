import copy
import json
from pathlib import Path
import subprocess
import sys
import unittest

ROOT = Path(__file__).resolve().parents[2]
TOOL = ROOT / 'tools/a153-evaluate-cost-aware-strategy-variants/analysis.py'
FIXTURE = ROOT / 'tools/a151-compare-stop-policies-offline/testdata/kr-us-synthetic.json'

class Contract(unittest.TestCase):
    def manifest(self):
        data = json.loads(FIXTURE.read_text())
        data['version'], data['mode'] = 'a153-v1', 'independent_filters'
        return data

    def run_tool(self, obj):
        return subprocess.run([sys.executable, str(TOOL), '-'], input=json.dumps(obj), text=True, capture_output=True)

    def test_cli_market_and_inconclusive(self):
        run = self.run_tool(self.manifest())
        self.assertEqual(run.returncode, 0, run.stderr)
        result = json.loads(run.stdout)
        self.assertEqual(result['profitability'], 'inconclusive')
        self.assertFalse(result['production_authorized'])
        self.assertEqual(set(result['markets']), {'KR', 'US'})
        self.assertEqual(len(result['candidates']), 2)

    def test_duplicate_opportunity(self):
        obj = self.manifest(); dup = copy.deepcopy(obj['candidates'][0]); dup['id'] = 'different'
        obj['candidates'].append(dup)
        self.assertNotEqual(self.run_tool(obj).returncode, 0)

    def test_invalid_timestamp_precision(self):
        obj = self.manifest(); obj['candidates'][0]['signal_close_at'] = '2026-01-05T10:15:00.000000001+09:00'
        self.assertNotEqual(self.run_tool(obj).returncode, 0)


class Detail(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        import importlib.util
        spec = importlib.util.spec_from_file_location('a153', TOOL)
        cls.tool = importlib.util.module_from_spec(spec); spec.loader.exec_module(cls.tool)
        cls.data = json.loads(FIXTURE.read_text())

    def test_feature_truth_unknown_and_future_independence(self):
        c = copy.deepcopy(self.data['candidates'][0])
        c['bars_5m'][-1]['close'] = '10010'
        c['bars_5m'][-1]['volume'] = '6'
        del c['bars_1m'][-3]
        self.assertEqual(self.tool.features(c), {'candle': None, 'volume': None, 'mfi': None})
        c['bars_1m'][-1]['close'] = '777777'
        self.assertEqual(self.tool.features(c), {'candle': None, 'volume': None, 'mfi': None})
        c['bars_5m'][-1]['received_at'] = '2026-01-05T10:15:01+09:00'
        self.assertEqual(self.tool.features(c), {'candle': None, 'volume': None, 'mfi': None})

    def test_duplicate_normalized_instant_with_different_confirmation(self):
        data = copy.deepcopy(self.data); data['version']='a153-v1'; data['mode']='independent_filters'
        c=copy.deepcopy(data['candidates'][0]); c['id']='another'; c['signal_close_at']='2026-01-05T01:15:00Z'; c['signal_confirmed_at']='2026-01-05T10:15:01+09:00'; data['candidates'].append(c)
        with self.assertRaisesRegex(ValueError,'duplicate opportunity'): self.tool.compare(data)

    def test_session_weighting_and_unknown_exclusion(self):
        def row(ident,date,r,selected=True):
            outcome={'status':'eligible','arms':{'fixed':{'risk':'1','base':{'status':'closed','pnl':str(r)}}}}
            return {'id':ident,'date':date,'filters':{'candle':selected,'volume':True,'mfi':True},'base':outcome,'stress':outcome}
        rows=[row('a','day1',2,False),row('b','day1',2,False),row('c','day2',-2,False)]
        result=self.tool.report_market(rows)
        self.assertEqual(result['summaries']['base']['candle']['difference'],'0')
        self.assertEqual(result['summaries']['base']['candle']['selected_count'],0)
        rows.append(row('d','day2',1,None))
        result=self.tool.report_market(rows)
        self.assertEqual(result['complete_sessions'],1)
        self.assertEqual(result['excluded_sessions']['day2']['ids'],['c','d'])
        self.assertEqual(result['summaries']['base']['candle']['difference'],'-2')
        self.assertEqual(result['summaries']['base']['candle']['bootstrap_95'],['-2','-2'])
        self.assertEqual(result['sample_sufficiency'],'insufficient')

    def test_missing_and_zero_risk_are_not_zero_return(self):
        self.assertIsNone(self.tool.base_r({'status':'eligible','arms':{'fixed':{'risk':'0','base':{'status':'closed','pnl':'0'}}}}))
        self.assertIsNone(self.tool.base_r({'status':'unfilled'}))

    def test_timestamp_offset_bounds(self):
        for bad in ('2026-01-05T10:15:00+08:60','2026-01-05T10:15:00+24:00','2026-01-05 10:15:00+09:00'):
            with self.assertRaises(ValueError): self.tool.timestamp(bad)

    def test_fee_monotonicity(self):
        from fractions import Fraction
        c=copy.deepcopy(self.data['candidates'][0]['cost']); before=self.tool.stop.pnl(Fraction(10000),Fraction(10200),3,c)
        c['buy_fixed']='10'; after=self.tool.stop.pnl(Fraction(10000),Fraction(10200),3,c)
        self.assertLess(after,before)

    def test_mfi_direction_and_late_or_stale(self):
        import datetime as dt
        c=copy.deepcopy(self.data['candidates'][0]); end=self.tool.timestamp(c['signal_close_at'])
        for i,b in enumerate(c['bars_1m'][-20:-2]):
            price=100+i if i<15 else (114-(i-14)*3)
            b.update(open=str(price),high=str(price),low=str(price),close=str(price),volume='1')
        self.assertFalse(self.tool.features(c)['mfi'])
        c['bars_1m'][-3]['received_at']=(end+dt.timedelta(seconds=1)).isoformat()
        self.assertIsNone(self.tool.features(c)['mfi'])
        c['bars_1m'][-3]['received_at']=end.isoformat()
        c['signal_confirmed_at']=(end+dt.timedelta(seconds=61)).isoformat()
        self.assertIsNone(self.tool.features(c)['mfi'])

    def test_slippage_reexecutes_with_new_entry(self):
        data=copy.deepcopy(self.data); data['version']='a153-v1'; data['mode']='independent_filters'
        for c in data['candidates']: c['entry_slippage_ticks']=1; c['exit_slippage_ticks']=1
        out=self.tool.compare(data)
        for row in out['candidates']:
            self.assertEqual(row['base']['status'],'eligible')
            self.assertEqual(row['stress']['status'],'eligible')
            self.assertNotEqual(row['base']['entry'],row['stress']['entry'])
            self.assertNotEqual(row['base']['arms']['fixed']['stop'],row['stress']['arms']['fixed']['stop'])

    def test_full_session_unknown_and_inherited_exclusion(self):
        data=copy.deepcopy(self.data); data['version']='a153-v1'; data['mode']='independent_filters'
        data['candidates'][0]['bars_5m']=data['candidates'][0]['bars_5m'][:14]
        out=self.tool.compare(data); row=out['candidates'][0]
        self.assertEqual(row['base']['status'],'excluded')
        self.assertEqual(row['base']['reason'],'missing_signal_bar')
        self.assertEqual(out['markets']['KR']['complete_sessions'],0)
        self.assertIsNone(out['markets']['KR']['summaries']['base']['candle']['difference'])

    def test_fixed_sensitivity_does_not_change_base_score(self):
        c=copy.deepcopy(self.data['candidates'][0]); outcome=self.tool.evaluate(c)
        self.assertTrue(outcome['arms']['fixed']['base']['ambiguous'])
        self.assertNotEqual(outcome['arms']['fixed']['base']['pnl'],outcome['arms']['fixed']['sensitivity']['pnl'])
        self.assertEqual(self.tool.base_r(outcome),self.tool.Fraction(outcome['arms']['fixed']['base']['simulated_r']))

    def test_bootstrap_percentile_and_reproducibility(self):
        values=[self.tool.Fraction(-2),self.tool.Fraction(2)]
        self.assertEqual(self.tool.bootstrap(values),['-2','2'])
        self.assertEqual(self.tool.bootstrap(values),self.tool.bootstrap(values))
        self.assertIsNone(self.tool.bootstrap([]))

    def test_cli_files_immutable_exclusive_and_symlink(self):
        import tempfile
        with tempfile.TemporaryDirectory() as d:
            path=Path(d)/'input.json'; output=Path(d)/'output.json'; data=copy.deepcopy(self.data)
            data['version']='a153-v1'; data['mode']='independent_filters'; path.write_text(json.dumps(data)); original=path.read_bytes()
            cmd=[sys.executable,str(TOOL),str(path),'--output',str(output)]
            run=subprocess.run(cmd,capture_output=True,text=True)
            self.assertEqual(run.returncode,0,run.stderr)
            self.assertEqual(path.read_bytes(),original); saved=output.read_bytes()
            self.assertNotEqual(subprocess.run(cmd,capture_output=True).returncode,0)
            self.assertEqual(output.read_bytes(),saved)
            link=Path(d)/'link'; link.symlink_to(Path(d),target_is_directory=True)
            self.assertNotEqual(subprocess.run([sys.executable,str(TOOL),str(link/'input.json')],capture_output=True).returncode,0)

    def test_cli_duplicate_keys_and_nonfinite(self):
        for raw in ('{"version":"a153-v1","version":"a153-v1"}', '{"x":NaN}'):
            run=subprocess.run([sys.executable,str(TOOL),'-'],input=raw,text=True,capture_output=True)
            self.assertNotEqual(run.returncode,0)

    def test_independent_unequal_session_and_bootstrap_oracle(self):
        def row(name,date,pnl):
            state={'status':'eligible','arms':{'fixed':{'risk':'1','base':{'status':'closed','pnl':str(pnl)}}}}
            return {'id':name,'date':date,'filters':{'candle':False,'volume':True,'mfi':True},'base':state,'stress':state}
        rows=[row('a','1',1),row('b','1',2),row('c','1',3),row('d','2',-4)]
        result=self.tool.report_market(rows)['summaries']['base']['candle']
        self.assertEqual(result['difference'],'1')
        self.assertEqual(result['bootstrap_95'],['-2','4'])

    def test_independent_fee_oracle(self):
        from fractions import Fraction as F
        cost={'buy_rate':'0.001','sell_rate':'0.001','tax_rate':'0.001','buy_fixed':'0.01','sell_fixed':'0.01','quantum':'0.01'}
        self.assertEqual(self.tool.stop.costs(F(100),F(102),3,cost),(F(31,100),F(32,100),F(31,100)))
        self.assertEqual(self.tool.stop.pnl(F(100),F(102),3,cost),F(253,50))
        self.assertEqual(self.tool.stop.pnl(F(100),F(102),3,cost)/3,F(253,150))

    def test_future_unused_bad_timestamp_is_input_error(self):
        data=copy.deepcopy(self.data); data['version']='a153-v1'; data['mode']='independent_filters'
        for timestamp in ('2026-01-05T10:17:00.0000001+09:00','2026-01-05T10:17:00+08:60'):
            data['candidates'][0]['bars_1m'][-1]['received_at']=timestamp
            with self.assertRaisesRegex(ValueError,'invalid timestamp'): self.tool.compare(data)

    def test_contradictory_signal_aggregate_is_unknown(self):
        c=copy.deepcopy(self.data['candidates'][0])
        c['bars_5m'][-1]['close']='10010'
        self.assertIsNone(self.tool.features(c)['candle'])

    def test_mfi_positive_direction_with_independent_flow_oracle(self):
        c=copy.deepcopy(self.data['candidates'][0])
        bars=c['bars_1m'][-20:-2]
        prices=[120-i for i in range(15)]+[108,110,112]
        for b,price in zip(bars,prices):
            b.update(open=str(price),high=str(price),low=str(price),close=str(price),volume='1')
        self.assertTrue(self.tool.features(c)['mfi'])
        converted=[{k:(b[k] if k=='close_at' else float(b[k])) for k in ('close_at','open','high','low','close','volume')} for b in bars]
        value=self.tool.mfi.mfi(converted,c['signal_confirmed_at'],60)
        self.assertAlmostEqual(value['mfi'],100*(108+110+112)/sum(prices[4:18]))
        self.assertEqual(value['direction'],'rising')

    def test_late_or_missing_signal_constituents_unknown(self):
        for kind in ('missing','late'):
            c=copy.deepcopy(self.data['candidates'][0])
            if kind == 'missing': del c['bars_1m'][-3]
            else: c['bars_1m'][-3]['received_at']='2026-01-05T10:15:01+09:00'
            self.assertIsNone(self.tool.features(c)['candle'])
            self.assertIsNone(self.tool.features(c)['volume'])

    def test_missing_previous_constituent_only_volume_unknown(self):
        c=copy.deepcopy(self.data['candidates'][0]); del c['bars_1m'][-8]
        self.assertFalse(self.tool.features(c)['candle'])
        self.assertIsNone(self.tool.features(c)['volume'])

    def test_complete_synthetic_has_known_signal_features(self):
        data=json.loads((ROOT/'tools/a153-evaluate-cost-aware-strategy-variants/testdata/kr-us-synthetic.json').read_text())
        for c in data['candidates']:
            expected={'candle':True,'volume':False,'mfi':False}
            self.assertEqual(self.tool.features(c), expected)
            c['bars_1m'][-1].update(open='1',high='1',low='1',close='1',volume='99999')
            self.assertEqual(self.tool.features(c), expected)

    def test_session_exclusion_preserves_specific_outcome_reason(self):
        bad={'status':'unfilled','reason':'missing_next_bar'}
        row={'id':'bad','date':'d','filters':{'candle':True,'volume':True,'mfi':True},'base':bad,'stress':bad}
        reasons=self.tool.report_market([row])['excluded_sessions']['d']['reasons']
        self.assertIn('bad:base:unfilled:missing_next_bar',reasons)

    def test_independent_stress_execution_oracle(self):
        data=copy.deepcopy(self.data); data['version']='a153-v1'; data['mode']='independent_filters'
        c=data['candidates'][1]; c['tick']['value']='1'; c['entry_slippage_ticks']=1; c['exit_slippage_ticks']=1
        c['bars_1m'][75].update(open='100',high='100',low='100',close='100')
        row=self.tool.compare(data)['candidates'][1]
        for scenario,entry,stop,exit_price,r in (('base','101','100','99','-2'),('stress','102','101','98','-4')):
            execution=row[scenario]; arm=execution['arms']['fixed']; outcome=arm['base']
            self.assertEqual((execution['entry'],arm['stop'],arm['quantity'],outcome['exit'],outcome['simulated_r']), (entry,stop,70,exit_price,r))

if __name__ == '__main__': unittest.main()

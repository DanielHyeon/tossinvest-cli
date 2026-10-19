import copy
import datetime as dt
import importlib.util
import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

HERE = Path(__file__).parent
spec = importlib.util.spec_from_file_location('analysis', HERE / 'analysis.py')
a = importlib.util.module_from_spec(spec)
spec.loader.exec_module(a)


def fixture(market='KR'):
    from zoneinfo import ZoneInfo
    zone = ZoneInfo('Asia/Seoul' if market == 'KR' else 'America/New_York')
    start = dt.datetime(2026, 1, 5, 9, 0 if market == 'KR' else 30, tzinfo=zone)
    tick, price, low, high, budget, curr = ('10','10000','9900','10100','7000','KRW') if market == 'KR' else ('0.1','100','99','101','70','USD')
    def bar(end, minutes):
        return {'close_at': end.isoformat(), 'received_at': end.isoformat(), 'open': price, 'high': high, 'low': low, 'close': price, 'volume': str(minutes)}
    minute = [bar(start + dt.timedelta(minutes=i), 1) for i in range(1, 78)]
    five = [bar(start + dt.timedelta(minutes=i*5), 5) for i in range(1, 16)]
    signal = start + dt.timedelta(minutes=75)
    meta = {'source':'synthetic','version':'v1','digest':'fixture','as_of':signal.isoformat()}
    return {'version':'a151-v1','synthetic':True,'mode':'initial_stops','candidates':[{
        'id':market, 'market':market, 'currency':curr, 'account_currency':curr, 'symbol':'TEST', 'date':'2026-01-05', 'basis':'unadjusted', 'data_source':'synthetic', 'data_version':'1', 'signal_confirmed_at':signal.isoformat(), 'signal_close_at':signal.isoformat(),
        'session':dict(meta, timezone=str(zone), kind='regular', open_at=start.isoformat(), close_at=(start+dt.timedelta(minutes=390)).isoformat()),
        'tick':dict(meta, market=market, symbol='TEST', valid_from=start.isoformat(), valid_until=(start+dt.timedelta(minutes=390)).isoformat(), value=tick, min_price='1', max_price='20000'),
        'cost':dict(meta, currency=curr, buy_rate='0', sell_rate='0', tax_rate='0', buy_fixed='0', sell_fixed='0', quantum='0.01'),
        'limits':dict(meta, currency=curr, risk_budget=budget, max_quantity='10000', max_notional='10000000'),
        'entry_slippage_ticks':0,'exit_slippage_ticks':0,'bars_1m':minute,'bars_5m':five}]}

class ContractTests(unittest.TestCase):
    def test_exact_geometry_and_atr(self):
        for tick, entry, low, atr, budget, expected in [
            ('10','10000','9900','80','7000',['9930','9890','9880']),
            ('0.1','100','99','0.8','70',['99.3','98.9','98.8'])]:
            arms = a.geometry(a.Fraction(entry), a.Fraction(low), a.Fraction(atr), a.Fraction(tick), a.Fraction(budget), 10000, a.Fraction('10000000'))
            self.assertEqual([arms[x]['quantity'] for x in ('fixed','structural','atr')],[100,63,58])
            self.assertEqual([arms[x]['stop'] for x in ('fixed','structural','atr')],expected)

    def test_atr_wilder(self):
        self.assertEqual(a.atr14([(101,99,100)]*15 + [(104,100,102)]), a.Fraction(15,7))

    def test_late_signal_and_missing_entry(self):
        x=fixture(); x['candidates'][0]['bars_5m'][-1]['received_at']='2026-01-05T10:16:00+09:00'
        self.assertNotEqual(a.compare(x)['candidates'][0]['status'],'eligible')
        x=fixture(); x['candidates'][0]['bars_1m']=x['candidates'][0]['bars_1m'][:75]
        self.assertEqual(a.compare(x)['candidates'][0]['status'],'unfilled')

    def test_currency_and_zero_cap(self):
        x=fixture('US'); x['candidates'][0]['limits']['currency']='KRW'
        self.assertNotEqual(a.compare(x)['candidates'][0]['status'],'eligible')
        x=fixture(); x['candidates'][0]['limits']['max_quantity']='0'
        self.assertNotEqual(a.compare(x)['candidates'][0]['status'],'eligible')

    def test_fee_quantum(self):
        c={'buy_rate':'0.001','sell_rate':'0.001','tax_rate':'0.001','buy_fixed':'0.01','sell_fixed':'0.01','quantum':'0.01'}
        self.assertEqual(a.pnl(a.Fraction(100),a.Fraction(102),3,c),a.Fraction(253,50))

    def test_reject_duplicate_keys_and_output_overwrite(self):
        with tempfile.TemporaryDirectory() as d:
            src=Path(d)/'input.json'; out=Path(d)/'out.json'
            src.write_text('{"version":"x","version":"y"}')
            result=subprocess.run([sys.executable,str(HERE/'analysis.py'),str(src)],capture_output=True)
            self.assertNotEqual(result.returncode,0)
            src.write_text(json.dumps(fixture())); out.write_text('keep')
            result=subprocess.run([sys.executable,str(HERE/'analysis.py'),str(src),'--output',str(out)],capture_output=True)
            self.assertNotEqual(result.returncode,0); self.assertEqual(out.read_text(),'keep')

    def test_entry_gap_stop_in_entry_bar(self):
        x=fixture('US'); c=x['candidates'][0]
        c['tick']['value']='1'; c['entry_slippage_ticks']=2; c['exit_slippage_ticks']=1
        for bar in c['bars_1m'][70:75]: bar.update(open='102',high='103',low='102',close='102')
        c['bars_5m'][-1].update(open='102',high='103',low='102',close='102')
        y=a.compare(x)['candidates'][0]
        self.assertEqual(y['entry'],'102')
        self.assertEqual(y['arms']['structural']['stop'],'101')
        self.assertEqual(y['arms']['structural']['base']['exit'],'99')
        self.assertEqual(y['arms']['structural']['base']['simulated_r'],'-3')
        self.assertEqual(y['entry_open_at'],'2026-01-05T10:46:00-05:00')

    def test_same_bar_ambiguity_and_complete_cohort(self):
        x=fixture('US'); c=x['candidates'][0]; c['bars_1m'][76].update(high='104',low='96')
        out=a.compare(x); row=out['candidates'][0]
        self.assertEqual(out['markets']['US']['complete'],1)
        for arm in row['arms'].values():
            self.assertTrue(arm['base']['ambiguous'])
            self.assertEqual(arm['base']['reason'],'stop_touch')
            self.assertEqual(arm['sensitivity']['reason'],'target_touch')
            self.assertEqual(arm['sensitivity']['exit'],'102.1')

    def test_censored_denominator_and_input_unchanged(self):
        x=fixture('US'); y=copy.deepcopy(x); out=a.compare(x)
        self.assertEqual(x,y); self.assertEqual(out['markets']['US']['complete'],0)
        self.assertIsNone(out['markets']['US']['paired_mean_simulated_r']['atr_minus_fixed'])
        self.assertEqual(out['candidates'][0]['arms']['atr']['base']['status'],'censored')

    def test_input_bounds_and_unknown(self):
        x=fixture(); x['unknown']=1
        with self.assertRaises(ValueError): a.compare(x)
        x=fixture(); x['candidates'][0]['tick']['value']='1p10000000'
        self.assertEqual(a.compare(x)['candidates'][0]['status'],'invalid')
        x=fixture(); x['candidates'].append(copy.deepcopy(x['candidates'][0]))
        with self.assertRaises(ValueError): a.compare(x)
        x=fixture(); x['mode']='ema9'
        with self.assertRaises(ValueError): a.compare(x)

    def test_late_snapshot_and_inconsistent_bars(self):
        x=fixture(); x['candidates'][0]['cost']['as_of']='2026-01-05T10:16:00+09:00'
        self.assertEqual(a.compare(x)['candidates'][0]['reason'],'future_snapshot')
        x=fixture(); x['candidates'][0]['bars_5m'][0]['volume']='4'
        self.assertEqual(a.compare(x)['candidates'][0]['reason'],'bar_aggregation_mismatch')

    def test_us_dst_early_close_holiday_and_offset(self):
        x=fixture('US'); c=x['candidates'][0]
        c['session']['kind']='early_close'; c['session']['close_at']='2026-01-05T13:00:00-05:00'
        self.assertEqual(a.compare(x)['candidates'][0]['status'],'eligible')
        c['session']['kind']='holiday'; self.assertEqual(a.compare(x)['candidates'][0]['reason'],'holiday')
        c['session']['kind']='regular'; c['session']['open_at']='2026-01-05T09:30:00-04:00'
        self.assertEqual(a.compare(x)['candidates'][0]['reason'],'session_date_or_offset')
        x=fixture('US'); c=x['candidates'][0]
        def shift(obj):
            for key,val in obj.items():
                if isinstance(val,str) and val.startswith('2026-01-05T'):
                    obj[key]=val.replace('2026-01-05T','2026-06-01T').replace('-05:00','-04:00')
        c['date']='2026-06-01'; shift(c)
        for key in ('session','tick','cost','limits'): shift(c[key])
        for b in c['bars_1m']+c['bars_5m']: shift(b)
        self.assertEqual(a.compare(x)['candidates'][0]['status'],'eligible')

    def test_symlink_output_and_input_hash(self):
        with tempfile.TemporaryDirectory() as d:
            root=Path(d); src=root/'input.json'; output=root/'link'; target=root/'target'; target.write_text('keep')
            src.write_text(json.dumps(fixture('US'))); before=src.read_bytes(); output.symlink_to(target)
            result=subprocess.run([sys.executable,str(HERE/'analysis.py'),str(src),'--output',str(output)],capture_output=True)
            self.assertNotEqual(result.returncode,0); self.assertEqual(target.read_text(),'keep'); self.assertEqual(src.read_bytes(),before)
            result=subprocess.run([sys.executable,str(HERE/'analysis.py'),str(src)],capture_output=True,check=True)
            import hashlib
            self.assertEqual(json.loads(result.stdout)['input_sha256'],hashlib.sha256(before).hexdigest())

    def test_paired_cohorts_and_exclusion(self):
        x=fixture('US'); closed=x['candidates'][0]; closed['bars_1m'][76].update(high='104',low='96')
        censored=copy.deepcopy(fixture('US')['candidates'][0]); censored['id']='censored'
        zero=copy.deepcopy(censored); zero['id']='zero'; zero['limits']['max_quantity']='0'
        x['candidates'] += [censored,zero]; m=a.compare(x)['markets']['US']
        self.assertEqual((m['total'],m['eligible'],m['complete']),(3,2,1))
        self.assertEqual(m['excluded']['zero_quantity'],1)

    def test_tick_scope_and_signal_snapshot(self):
        x=fixture('US'); x['candidates'][0]['tick']['symbol']='OTHER'
        self.assertEqual(a.compare(x)['candidates'][0]['reason'],'tick_scope_or_validity')
        x=fixture('US'); x['candidates'][0]['tick']['valid_until']='2026-01-05T10:30:00-05:00'
        self.assertEqual(a.compare(x)['candidates'][0]['status'],'unsupported')
        x=fixture('US'); x['candidates'][0]['bars_1m'][0]['received_at']='2026-01-05T08:00:00-05:00'
        self.assertEqual(a.compare(x)['candidates'][0]['reason'],'received_before_close')
        x=fixture('US'); x['candidates'][0]['bars_1m'][1]=copy.deepcopy(x['candidates'][0]['bars_1m'][0])
        self.assertEqual(a.compare(x)['candidates'][0]['reason'],'duplicate_or_reversed_bars')

    def test_no_lookahead_in_entry_geometry(self):
        x=fixture('US'); c=x['candidates'][0]; first=a.compare(x)['candidates'][0]
        c['bars_1m'][76].update(high='104',low='96',close='101')
        later=a.compare(x)['candidates'][0]
        self.assertEqual(first['entry'],later['entry'])
        for name in a.ARMS:
            for field in ('stop','target','quantity','risk'):
                self.assertEqual(first['arms'][name][field],later['arms'][name][field])

    def test_gap_fill_and_no_optimistic_target_gap(self):
        x=fixture('US'); c=x['candidates'][0]; c['bars_1m'][76]['low']='99.5'
        b=copy.deepcopy(c['bars_1m'][76]); b.update(close_at='2026-01-05T10:48:00-05:00',received_at='2026-01-05T10:48:00-05:00',open='96',high='100',low='95',close='99'); c['bars_1m'].append(b)
        row=a.compare(x)['candidates'][0]
        self.assertEqual(row['arms']['fixed']['base']['reason'],'stop_gap')
        self.assertEqual(row['arms']['fixed']['base']['exit'],'96')
        x=fixture('US'); c=x['candidates'][0]; c['bars_1m'][76]['low']='99.5'
        b=copy.deepcopy(c['bars_1m'][76]); b.update(close_at='2026-01-05T10:48:00-05:00',received_at='2026-01-05T10:48:00-05:00',open='104',high='105',low='103',close='104'); c['bars_1m'].append(b)
        row=a.compare(x)['candidates'][0]
        self.assertTrue(all(arm['base']['exit']=='102.1' for arm in row['arms'].values()))

    def test_no_network_or_live_imports(self):
        import ast
        tree=ast.parse((HERE/'analysis.py').read_text())
        names={n.names[0].name for n in ast.walk(tree) if isinstance(n,ast.Import)}
        self.assertTrue(names <= {'argparse','datetime','hashlib','json','os','re','sys'})
        self.assertTrue({n.module for n in ast.walk(tree) if isinstance(n,ast.ImportFrom)} <= {'fractions','pathlib','zoneinfo'})

    def test_invalid_exit_and_missing_path(self):
        x=fixture('US'); c=x['candidates'][0]; c['exit_slippage_ticks']=980; c['bars_1m'][76].update(low='96',high='104')
        out=a.compare(x); row=out['candidates'][0]
        self.assertEqual(row['status'],'eligible')
        self.assertEqual(row['arms']['atr']['base']['status'],'invalid')
        self.assertEqual(row['arms']['atr']['sensitivity']['status'],'closed')
        self.assertEqual(row['arms']['fixed']['base']['status'],'closed')
        self.assertEqual(out['markets']['US']['complete'],0)
        self.assertEqual(out['markets']['US']['arms']['atr']['invalid'],1)
        x=fixture('US'); c=x['candidates'][0]; c['bars_1m'][76]['low']='99.5'
        b=copy.deepcopy(c['bars_1m'][76]); b.update(close_at='2026-01-05T10:49:00-05:00',received_at='2026-01-05T10:49:00-05:00',high='104',low='96'); c['bars_1m'].append(b)
        row=a.compare(x)['candidates'][0]
        self.assertEqual(row['arms']['fixed']['base']['status'],'censored')
        self.assertEqual(row['arms']['fixed']['base']['mark_at'],'2026-01-05T10:47:00-05:00')

    def test_unknown_candidate_field_is_structural_failure(self):
        x=fixture('US'); x['candidates'][0]['unexpected']=1
        with self.assertRaises(a.ManifestError): a.compare(x)

    def test_tick_validity_covers_warmup_and_full_session(self):
        x=fixture('US'); x['candidates'][0]['tick']['valid_from']='2026-01-05T10:00:00-05:00'
        self.assertEqual(a.compare(x)['candidates'][0]['status'],'unsupported')

    def test_unverifiable_future_aggregate_does_not_erase_outcomes(self):
        x=fixture('US'); c=x['candidates'][0]
        b=copy.deepcopy(c['bars_5m'][-1]); b.update(close_at='2026-01-05T10:50:00-05:00',received_at='2026-01-05T10:50:00-05:00'); c['bars_5m'].append(b)
        result=a.compare(x)['candidates'][0]
        self.assertEqual(result['status'],'eligible')
        self.assertEqual(result['arms']['fixed']['base']['status'],'closed')
        self.assertEqual(result['arms']['atr']['base']['status'],'censored')
        self.assertEqual(result['unverified_future_5m'],1)

    def test_complete_future_aggregate_contradiction_is_rejected(self):
        x=fixture('US'); c=x['candidates'][0]
        for minute in (48,49,50):
            b=copy.deepcopy(c['bars_1m'][-1]); stamp=f'2026-01-05T10:{minute}:00-05:00'; b.update(close_at=stamp,received_at=stamp); c['bars_1m'].append(b)
        b=copy.deepcopy(c['bars_5m'][-1]); b.update(close_at='2026-01-05T10:50:00-05:00',received_at='2026-01-05T10:50:00-05:00',volume='4'); c['bars_5m'].append(b)
        self.assertEqual(a.compare(x)['candidates'][0]['reason'],'bar_aggregation_mismatch')

if __name__ == '__main__': unittest.main()

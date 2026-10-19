import copy
import importlib.util
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest

HERE = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location('a152', HERE / 'analysis.py')
a = importlib.util.module_from_spec(spec)
spec.loader.exec_module(a)

def dataset(kind='trades', market='KR'):
    tz, currency, offset = ('Asia/Seoul','KRW','+09:00') if market == 'KR' else ('America/New_York','USD','-04:00')
    date='2026-10-15'; op=f'{date}T09:00:00{offset}'; end=f'{date}T10:00:00{offset}'; now=f'{date}T10:05:00{offset}'; close=f'{date}T15:00:00{offset}'
    declared={'trades':'executions','bins':'executed_price_bins','ohlcv':'ohlcv','orderbook':'orderbook','unavailable':'unknown'}[kind]
    records=[{'event_id':'1','action':'trade','reference_id':None,'executed_at':f'{date}T09:15:00{offset}','received_at':f'{date}T09:15:01{offset}','price':'100','volume':'2'}, {'event_id':'2','action':'trade','reference_id':None,'executed_at':f'{date}T09:20:00{offset}','received_at':f'{date}T09:20:01{offset}','price':'101','volume':'3'}] if kind=='trades' else []
    method=None
    if kind=='bins':
        records=[{'lower':'100','upper':'101','volume':'2'},{'lower':'101','upper':'102','volume':'0'},{'lower':'102','upper':'103','volume':'3'}]
        method={'name':'supplied','volume_basis':'executed','unit':'shares','boundary':'[lower,upper)','width':'1','price_min':'100','price_max':'103','zero_bins_explicit':True,'built_at':f'{date}T10:01:00{offset}','received_at':f'{date}T10:02:00{offset}'}
    return {'id':market+kind,'kind':kind,'market':market,'symbol':'X','currency':currency,'basis':'unadjusted','profile_mode':'current_prefix','session':{'date':date,'timezone':tz,'open_at':op,'close_at':close},'cutoff_at':end,'as_of':now,'source':{'name':'supplied','version':'1','reference':'local','declared_kind':declared},'coverage':{'start_at':op,'end_at':end,'status':'complete','gaps':[],'sequence_status':'complete','venues':['X'],'trade_types':['regular']},'correction_policy':'none_declared','price_increment':'1','records':records,'bin_method':method,'access':{'license':'provided claim','availability':'provided','cost_note':'unknown'}}

def manifest(*items, synthetic=True): return {'version':'a152-v1','synthetic':synthetic,'datasets':list(items)}
def result(d,synthetic=False): return a.qualify(manifest(d,synthetic=synthetic))['datasets'][0]

class Qualification(unittest.TestCase):
    def test_trade_and_bins_exact_oracles_both_markets(self):
        for market in ('KR','US'):
            for kind in ('trades','bins'):
                row=result(dataset(kind,market)); self.assertEqual(row['status'],'eligible_unverified'); self.assertEqual(row['total_volume'],'5'); self.assertTrue(row['vp_input_eligible'])
    def test_output_contains_validated_interpretation_metadata(self):
        d=dataset('bins','US'); row=result(d); self.assertEqual((row['market'],row['symbol'],row['currency'],row['basis']),('US','X','USD','unadjusted')); self.assertEqual(row['session'],d['session']); self.assertEqual(row['bin_method'],d['bin_method'])
        d['bin_method']['zero_bins_explicit']=False; row=result(d); self.assertEqual(row['status'],'unsupported'); self.assertEqual(row['bin_method'],d['bin_method'])
        d=dataset('bins'); d['coverage']['status']='partial'; row=result(d); self.assertEqual(row['status'],'not_ready'); self.assertEqual(row['bin_method'],d['bin_method'])
    def test_synthetic_never_empirical(self):
        out=a.qualify(manifest(dataset())); row=out['datasets'][0]; self.assertFalse(row['vp_input_eligible']); self.assertIn('synthetic_only',row['vp_exclusion_reasons']); self.assertFalse(out['production_authorized']); self.assertFalse(out['non_profile_research_blocked'])
    def test_proxies(self):
        for k,c,reason in [('ohlcv','ohlcv_proxy','ohlcv_is_not_execution_profile'),('orderbook','unavailable','book_size_is_not_executed_volume'),('unavailable','unavailable','no_profile_evidence')]:
            r=result(dataset(k)); self.assertEqual((r['classification'],r['status']),(c,'unsupported')); self.assertIn(reason,r['reasons'])
    def test_corrections_and_policy(self):
        for action in ('correct','cancel'):
            d=dataset(); d['records'][0].update(action=action,reference_id='original'); self.assertEqual(result(d)['status'],'unsupported')
        for p in ('unknown','unresolved'):
            d=dataset(); d['correction_policy']=p; self.assertEqual(result(d)['status'],'unsupported')
    def test_late_bin_and_trade(self):
        d=dataset('bins'); d['bin_method']['received_at']='2026-10-15T10:06:00+09:00'; self.assertNotEqual(result(d)['status'],'eligible_unverified')
        d=dataset(); d['records'][0]['received_at']='2026-10-15T10:06:00+09:00'; self.assertNotEqual(result(d)['status'],'eligible_unverified')
    def test_coverage_and_duplicate(self):
        d=dataset(); d['coverage']['status']='partial'; self.assertEqual(result(d)['status'],'not_ready')
        d=dataset(); d['records'][1]['event_id']='1'; self.assertEqual(result(d)['status'],'invalid')
    def test_bins_gaps_and_zero(self):
        d=dataset('bins'); d['records'].pop(1); self.assertNotEqual(result(d)['status'],'eligible_unverified')
        d=dataset('bins'); [b.update(volume='0') for b in d['records']]; self.assertEqual(result(d)['status'],'not_ready')
    def test_invalid_sibling_isolated(self):
        d=dataset(); d['id']='bad'; d['price_increment']='1e3'; out=a.qualify(manifest(d,dataset())); self.assertEqual([x['status'] for x in out['datasets']],['invalid','eligible_unverified'])
    def test_input_unchanged(self):
        m=manifest(dataset()); old=copy.deepcopy(m); a.qualify(m); self.assertEqual(m,old)
    def test_prior_session(self):
        d=dataset(); d['profile_mode']='prior_session_fixed'; d['cutoff_at']=d['session']['close_at']; d['coverage']['end_at']=d['cutoff_at']; d['as_of']='2026-10-16T09:00:00+09:00'; self.assertEqual(result(d)['status'],'eligible_unverified')
    def test_cli_hash_and_exclusive_output(self):
        with tempfile.TemporaryDirectory() as tmp:
            p=Path(tmp)/'in.json'; p.write_text(json.dumps(manifest(dataset()))); before=p.read_bytes(); output=Path(tmp)/'out.json'; cmd=[sys.executable,str(HERE/'analysis.py'),str(p),'--output',str(output)]
            self.assertEqual(subprocess.run(cmd,capture_output=True).returncode,0); self.assertEqual(p.read_bytes(),before); self.assertEqual(subprocess.run(cmd,capture_output=True).returncode,1)
            link=Path(tmp)/'link'; link.symlink_to(output); self.assertEqual(subprocess.run([sys.executable,str(HERE/'analysis.py'),str(p),'--output',str(link)],capture_output=True).returncode,1)
    def test_invalid_json_duplicate_key(self):
        with tempfile.TemporaryDirectory() as tmp:
            p=Path(tmp)/'x'; p.write_text('{"version":"a152-v1","version":"a152-v1","synthetic":true,"datasets":[]}'); self.assertEqual(subprocess.run([sys.executable,str(HERE/'analysis.py'),str(p)],capture_output=True).returncode,1)
    def test_trade_timestamp_boundaries_and_receipt_order(self):
        for field,value in [('executed_at','2026-10-15T10:00:00+09:00'),('received_at','2026-10-15T09:14:00+09:00')]:
            d=dataset(); d['records'][0][field]=value; self.assertEqual(result(d)['status'],'invalid')
        d=dataset(); d['records'][0]['received_at']='2026-10-15T09:21:00+09:00'; self.assertEqual(result(d)['status'],'invalid')
        d=dataset(); d['records'][0]['executed_at']='2026-10-15T09:19:00+09:00'; d['records'][0]['received_at']='2026-10-15T09:21:00+09:00'; d['records'][1]['executed_at']='2026-10-15T09:18:00+09:00'; d['records'][1]['received_at']='2026-10-15T09:22:00+09:00'; self.assertEqual(result(d)['status'],'eligible_unverified')
    def test_prefix_delay_and_cutoff(self):
        d=dataset('bins'); self.assertEqual(result(d)['status'],'eligible_unverified')
        d['bin_method']['built_at']='2026-10-15T09:59:00+09:00'; self.assertEqual(result(d)['status'],'invalid')
        d=dataset(); d['cutoff_at']='2026-10-15T10:06:00+09:00'; self.assertEqual(result(d)['status'],'invalid')
    def test_declared_gap_partial_and_scope(self):
        d=dataset(); d['coverage']['gaps']=[{'start_at':'2026-10-15T09:10:00+09:00','end_at':'2026-10-15T09:11:00+09:00'}]; self.assertEqual(result(d)['status'],'not_ready')
        d=dataset(); d['coverage']['end_at']='2026-10-15T09:59:00+09:00'; self.assertEqual(result(d)['status'],'not_ready')
        d=dataset(); d['coverage']['sequence_status']='unknown'; self.assertEqual(result(d)['status'],'not_ready')
    def test_bin_overlap_grid_method_and_empty(self):
        d=dataset('bins'); d['records'][1]['lower']='100'; self.assertEqual(result(d)['status'],'invalid')
        d=dataset('bins'); d['bin_method']['zero_bins_explicit']=False; self.assertEqual(result(d)['status'],'unsupported')
        d=dataset('bins'); d['records']=[]; self.assertEqual(result(d)['status'],'not_ready')
        d=dataset(); d['records']=[]; self.assertEqual(result(d)['status'],'not_ready')
        d=dataset(); d['records'][0]['price']='100.5'; self.assertEqual(result(d)['status'],'invalid')
    def test_restricted_mismatch_and_bad_fields(self):
        d=dataset(); d['access']['availability']='restricted'; self.assertEqual(result(d)['status'],'unsupported')
        d=dataset(); d['source']['declared_kind']='ohlcv'; self.assertEqual(result(d)['status'],'unsupported')
        d=dataset(); d['session']['extra']='x'; self.assertEqual(result(d)['status'],'invalid')
        d=dataset('ohlcv'); d['records']=[{}]; self.assertEqual(result(d)['status'],'invalid')
    def test_numeric_boundary(self):
        for value in ('1e2','1p2','0x1','1_0','-1','1'*65,True,1):
            d=dataset(); d['price_increment']=value; self.assertEqual(result(d)['status'],'invalid',str(value))
    def test_us_dst_and_prior_date(self):
        d=dataset(market='US'); d['session']['open_at']='2026-10-15T09:00:00-05:00'; self.assertEqual(result(d)['status'],'invalid')
        d=dataset(); d['profile_mode']='prior_session_fixed'; d['cutoff_at']=d['session']['close_at']; d['coverage']['end_at']=d['cutoff_at']; d['as_of']='2026-10-15T16:00:00+09:00'; self.assertEqual(result(d)['status'],'invalid')
    def test_invalid_envelope_and_duplicate_dataset(self):
        for obj in [manifest(dataset(),dataset()), {'version':'a152-v1','synthetic':1,'datasets':[]}, {'version':'a152-v1','synthetic':True,'datasets':[{}]*101}]:
            with self.assertRaises(a.Invalid): a.qualify(obj)
    def test_zero_lookahead_and_authority(self):
        d=dataset(); d['records'][1]['executed_at']=d['cutoff_at']; row=result(d); self.assertFalse(row['vp_input_eligible']); self.assertNotIn('total_volume',row)
        out=a.qualify(manifest(dataset(),synthetic=False)); self.assertFalse(out['authenticity_verified']); self.assertFalse(out['coverage_verified']); self.assertFalse(out['production_authorized']); self.assertEqual(out['profitability'],'inconclusive')
    def test_parent_symlink_and_missing_output(self):
        with tempfile.TemporaryDirectory() as tmp:
            root=Path(tmp); src=root/'in'; src.write_text(json.dumps(manifest(dataset()))); (root/'real').mkdir(); (root/'alias').symlink_to(root/'real',target_is_directory=True)
            p=subprocess.run([sys.executable,str(HERE/'analysis.py'),str(src),'--output',str(root/'alias'/'out')],capture_output=True); self.assertEqual(p.returncode,1); self.assertFalse((root/'real'/'out').exists())
    def test_nested_types_and_limits(self):
        for key in ('session','source','coverage','access','records','bin_method'):
            d=dataset('bins'); d[key]=42; self.assertEqual(result(d)['status'],'invalid',key)
        d=dataset(); d['source']['name']='x'*513; self.assertEqual(result(d)['status'],'invalid')
        d=dataset(); d['coverage']['venues']=['x']*33; self.assertEqual(result(d)['status'],'invalid')
        d=dataset(); d['coverage']['gaps']=[{}]*101; self.assertEqual(result(d)['status'],'invalid')
        d=dataset(); d['records']=[{}]*100001
        with self.assertRaises(a.Invalid): a.qualify(manifest(d))
    def test_cli_rejects_oversize_and_constants(self):
        with tempfile.TemporaryDirectory() as tmp:
            p=Path(tmp)/'input'; p.write_bytes(b' '* (16*1024*1024+1)); self.assertEqual(subprocess.run([sys.executable,str(HERE/'analysis.py'),str(p)],capture_output=True).returncode,1)
            p.write_text('{"version":"a152-v1","synthetic":true,"datasets":[],"bad":NaN}'); self.assertEqual(subprocess.run([sys.executable,str(HERE/'analysis.py'),str(p)],capture_output=True).returncode,1)
    def test_future_submicrosecond_and_fractional_offset_rejected(self):
        d=dataset(); d['records'][1]['received_at']='2026-10-15T10:05:00.000000001+09:00'; self.assertEqual(result(d)['status'],'invalid')
        d=dataset('bins'); d['bin_method']['received_at']='2026-10-15T10:05:00.000000001+09:00'; self.assertEqual(result(d)['status'],'invalid')
        d=dataset(); d['records'][1]['received_at']='2026-10-15T09:20:01+09:00:00.000000001'; self.assertEqual(result(d)['status'],'invalid')
    def test_timestamp_strict_subset(self):
        for value in ('2026-10-15 09:20:01+09:00','20261015T092001+0900','2026-10-15T09:20+09:00','2026-10-15T09:20:01.1234567+09:00'):
            d=dataset(); d['records'][1]['received_at']=value; self.assertEqual(result(d)['status'],'invalid',value)
        d=dataset(); d['records'][1]['received_at']='2026-10-15T09:20:01.123456+09:00'; self.assertEqual(result(d)['status'],'eligible_unverified')
    def test_offset_minutes_are_not_normalized(self):
        d=dataset(); d['session']['open_at']='2026-10-15T09:00:00+08:60'; self.assertEqual(result(d)['status'],'invalid')
        d=dataset(); d['records'][1]['received_at']='2026-10-15T09:20:01+08:60'; self.assertEqual(result(d)['status'],'invalid')
        d=dataset('bins'); d['bin_method']['received_at']='2026-10-15T10:02:00+08:60'; self.assertEqual(result(d)['status'],'invalid')
if __name__=='__main__': unittest.main()

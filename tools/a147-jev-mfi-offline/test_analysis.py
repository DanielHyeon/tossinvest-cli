import datetime as dt
import hashlib
import json
import os
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path
import analysis

class MfiTest(unittest.TestCase):
 def bars(self):
  return [{'close_at':f'2026-10-20T09:{i:02d}:00+09:00','open':100+i,'high':100+i,'low':100+i,'close':100+i,'volume':2} for i in range(1,19)]
 def test_rising_and_boundary(self):
  result=analysis.mfi(self.bars(),'2026-10-20T09:18:00+09:00',60)
  self.assertEqual((result['mfi'],result['direction']),(100,'unchanged'))
 def test_all_negative_flow_is_zero(self):
  bars=self.bars()
  for i,bar in enumerate(bars):
   value=200-i;bar.update(open=value,high=value,low=value,close=value)
  self.assertEqual(analysis.mfi(bars,'2026-10-20T09:18:00+09:00',60)['mfi'],0)
 def test_down_and_change(self):
  b=self.bars()
  for i in range(15,18): b[i].update(open=80,high=80,low=80,close=80)
  self.assertEqual(analysis.mfi(b,'2026-10-20T09:18:00+09:00',60)['direction'],'falling')
 def test_invalid_and_temporal(self):
  b=self.bars()
  for sample,reason in [(b[:17],'insufficient'),(b[:-1]+[dict(b[-1],close_at='2026-10-20T09:19:00+09:00')],'future'),(b[:-1]+[dict(b[-1],close_at=b[-2]['close_at'])],'nonconsecutive'),(b[:-1]+[dict(b[-1],volume=-1)],'invalid_bar')]:
   with self.subTest(reason=reason):self.assertEqual(analysis.mfi(sample,'2026-10-20T09:18:00+09:00',60)['reason'],reason)
 def test_epsilon_and_invalid_open(self):
  for x in (-1e-9, 0, 1e-9):self.assertEqual(analysis.delta_direction(x),'unchanged')
  self.assertEqual(analysis.delta_direction(1.000001e-9),'rising')
  self.assertEqual(analysis.delta_direction(-1.000001e-9),'falling')
  for value in (float('nan'),0,999):
   bars=self.bars();bars[-1]['open']=value
   self.assertEqual(analysis.mfi(bars,'2026-10-20T09:18:00+09:00',60)['reason'],'invalid_bar')
 def test_zero_flow_and_stale(self):
  b=self.bars()
  for x in b:x.update(open=100,high=100,low=100,close=100)
  self.assertEqual(analysis.mfi(b,'2026-10-20T09:18:00+09:00',60)['reason'],'undefined_flow')
  b=self.bars();b[-1].update(open=1e308, high=1e308, low=1e308, close=1e308)
  self.assertEqual(analysis.mfi(b,'2026-10-20T09:18:00+09:00',60)['reason'],'invalid_bar')
  self.assertEqual(analysis.mfi(self.bars(),'2026-10-20T09:20:00+09:00',60)['reason'],'stale')
 def test_offset_required(self):
  b=self.bars();b[-1]['close_at']='2026-10-20T09:18:00';self.assertEqual(analysis.mfi(b,'2026-10-20T09:18:00+09:00',60)['reason'],'invalid_bar')
 def test_exclusive_write(self):
  with tempfile.TemporaryDirectory() as d:
   p=Path(d)/'plan.json';analysis.write_once(p,{'a':1})
   with self.assertRaises(FileExistsError):analysis.write_once(p,{'a':2})
   self.assertIn('1',p.read_text())
 def test_manifest_limits(self):
  with self.assertRaises(ValueError):analysis.validate_plan({'max_bar_age_seconds':None,'max_label_delay_seconds':10})
 def test_prevent_symlink_output(self):
  with tempfile.TemporaryDirectory() as d:
   root=Path(d);src=root/'src';src.mkdir();link=root/'link';link.symlink_to(src,target_is_directory=True)
   with self.assertRaises(ValueError):analysis.safe_output(root/'src',link/'output')
class PlanTest(unittest.TestCase):
 def test_required_plan_fields_and_horizon(self):
  fields=('completeness_rules','candidate_rule','cycle_schedule','interruption_rules','replacement_rules','interim_reporting_rules','stop_rules','analysis_version')
  plan={key:'fixed' for key in fields};plan['session_plan']={'KR':[{'date':'2026-10-20','open_at':'2026-10-20T09:00:00+09:00','close_at':'2026-10-20T15:30:00+09:00','cadence_seconds':60,'max_cycle_skew_seconds':2}], 'US':[{'date':'2026-10-20','open_at':'2026-10-20T09:30:00-04:00','close_at':'2026-10-20T16:00:00-04:00','cadence_seconds':60,'max_cycle_skew_seconds':2}]}
  plan.update(coverage_thresholds={'min_cycle_fraction':1,'min_mfi_fraction':1,'min_label_fraction':1},sources={'KR':{'name':'fixture','timezone':'Asia/Seoul'},'US':{'name':'fixture','timezone':'America/New_York'}},max_bar_age_seconds=60,max_label_delay_seconds=90,horizon_minutes=60)
  analysis.validate_plan(plan)
  for key in fields:
   bad=dict(plan);bad[key]=''
   with self.subTest(key=key), self.assertRaises(ValueError):analysis.validate_plan(bad)
  bad=dict(plan,horizon_minutes=30)
  with self.assertRaises(ValueError):analysis.validate_plan(bad)

class CycleTest(unittest.TestCase):
 def test_skew_normalizes_without_joining_next_cycle(self):
  window={'open_at':'2026-10-20T09:00:00+09:00','close_at':'2026-10-20T10:00:00+09:00','cadence_seconds':60,'max_cycle_skew_seconds':2}
  slot=lambda t:analysis.cycle_slot(analysis.timestamp('2026-10-20T09:'+t+':00+09:00'),window)
  first=analysis.timestamp('2026-10-20T09:18:01+09:00')
  self.assertEqual(analysis.cycle_slot(first,window),slot('18'))
  self.assertNotEqual(analysis.cycle_slot(first,window),slot('19'))
  with self.assertRaises(ValueError):analysis.cycle_slot(analysis.timestamp('2026-10-20T09:18:03+09:00'),window)

class DuplicateTest(unittest.TestCase):
 def test_same_slot_and_symbol_collision_is_unknown(self):
  window={'open_at':'2026-10-20T09:00:00+09:00','close_at':'2026-10-20T10:00:00+09:00','cadence_seconds':60,'max_cycle_skew_seconds':2}
  def row(symbol,seconds):
   observed=analysis.timestamp('2026-10-20T09:18:'+seconds+'+09:00')
   return {'market':'KR','session':'2026-10-20','cycle':analysis.cycle_slot(observed,window),'symbol':symbol,'selection':'selected','direction':'rising','return':.01}
  rows,count=analysis.deduplicate_records([row('A','00'),row('A','01'),row('B','01')])
  self.assertEqual((len(rows),count),(2,1))
  conflict=next(r for r in rows if r['symbol']=='A')
  self.assertEqual((conflict['selection'],conflict['mfi_reason']),('unknown','duplicate_cycle_identity'))
  self.assertEqual(next(r for r in rows if r['symbol']=='B')['selection'],'selected')

class SidecarTest(unittest.TestCase):
 def test_append_no_overwrite_and_symlink(self):
  with tempfile.TemporaryDirectory() as d:
   p=Path(d)/'events';event={'kind':'capture','state_digest':'a'*64,'capture_digest':'b'*64}
   analysis.append_event(p,event,'1'*64);before=p.read_bytes()
   with self.assertRaises(ValueError):analysis.append_event(p,event,'1'*64)
   self.assertEqual(p.read_bytes(),before)
   link=Path(d)/'link';link.symlink_to(p)
   with self.assertRaises(OSError):analysis.append_event(link,dict(event,state_digest='c'*64),'1'*64)
   self.assertEqual(p.read_bytes(),before)
   with self.assertRaises(ValueError):analysis.append_event(p,dict(event,state_digest='e'*64),'2'*64)
   self.assertEqual(p.read_bytes(),before)
   hard=Path(d)/'hard';os.link(p,hard)
   with self.assertRaises(ValueError):analysis.append_event(hard,dict(event,state_digest='d'*64),'1'*64)
   self.assertEqual(p.read_bytes(),before)

class CandidateTest(unittest.TestCase):
 def test_unknown_and_inconsistent_remain_in_denominator(self):
  common={'state_rev':'s','market':'KR','at':'2026-10-20T09:18:00+09:00','symbol':'A','state_digest':'d'}
  one=dict(common,question_id='j1_up_within_horizon',screen=None)
  two=dict(common,question_id='j2_hold_off_red_flag',screen=None)
  result=analysis.pair_candidates([one,two])
  self.assertEqual((len(result),result[0]['selection'],result[0]['pair_valid']),(1,'unknown',True))
  two['screen']={'finalist':False}
  result=analysis.pair_candidates([one,two])
  self.assertEqual((len(result),result[0]['selection'],result[0]['pair_valid']),(1,'unknown',False))
  other=[dict(one,id='3',state_digest='different'),dict(two,id='4',state_digest='different')]
  result=analysis.pair_candidates([one,two,*other])
  self.assertEqual((len(result),result[0]['selection'],result[0]['pair_valid']),(1,'unknown',False))
 def test_label_due_and_delay(self):
  j={'at':'2026-10-20T09:18:00+09:00','label_due_at':'2026-10-20T10:18:00+09:00','horizon_minutes':60,'base_price':100}
  label={'due_at':j['label_due_at'],'at':'2026-10-20T10:18:10+09:00','realized_price':101,'realized_return':.01}
  self.assertEqual(analysis.label_return(j,label,10),(.01,None))
  self.assertEqual(analysis.label_return(j,None,10),(None,'missing_label'))
  invalid=dict(label,realized_price=-1)
  self.assertEqual(analysis.label_return(j,invalid,10),(None,'invalid_label'))
  self.assertEqual(analysis.label_return(j,label,9),(None,'late_label'))
  gap={'due_at':j['label_due_at'],'at':'2026-10-20T10:18:00+09:00','label_gap':True}
  self.assertEqual(analysis.label_return(j,gap,10),(None,'label_gap'))
  gap['realized_price']=101
  self.assertEqual(analysis.label_return(j,gap,10),(None,'invalid_label'))
  label['at']='2026-10-20T10:17:59+09:00'
  self.assertEqual(analysis.label_return(j,label,10),(None,'invalid_label'))
class SummaryTest(unittest.TestCase):
 def test_equal_cycle_weight_and_unknown_counts(self):
  rows=[]
  for cycle,direction,value in [('a','rising',.1),('a','rising',.3),('a','falling',.1),('b','rising',.05),('b','falling',.02)]:
   rows.append(dict(market='KR',session='2026-10-20',selection='selected',cycle=cycle,direction=direction,return_value=value))
  for row in rows:row['return']=row.pop('return_value')
  rows.append(dict(market='KR',session='2026-10-20',selection='unknown',cycle='a',mfi_reason='unproven',label_reason='missing_label'))
  result=analysis.summarize(rows)
  selected=next(x for x in result if x['selection']=='selected')
  unknown=next(x for x in result if x['selection']=='unknown')
  self.assertEqual((selected['candidates'],selected['matched_cycles'],unknown['candidates']),(5,2,1))
  self.assertAlmostEqual(selected['equal_cycle_rising_minus_falling_pp'],6.5)
class CliTest(unittest.TestCase):
 def test_fixture_end_to_end_and_missing_capture(self):
  with tempfile.TemporaryDirectory() as directory:
   root=Path(directory); source=root/'snapshot';source.mkdir();(source/'states').mkdir();(source/'captures').mkdir()
   bars=MfiTest().bars()
   for bar in bars:bar['open']=bar['close']
   state={'market':'KR','symbol':'A','observed_at':'2026-10-20T09:18:00+09:00','base_price':100,'intraday':{'bars_1m':[{'t':b['close_at'][11:16],'o':b['open'],'h':b['high'],'l':b['low'],'c':b['close'],'v':b['volume']} for b in bars]}}
   state_bytes=json.dumps(state).encode();digest=hashlib.sha256(state_bytes).hexdigest();(source/'states'/f'{digest}.json').write_bytes(state_bytes+b'\n')
   evidence={'state_digest':digest,'market':'KR','symbol':'A','observed_at':state['observed_at'],'source_timezone':'Asia/Seoul','source':'fixture','received_at':'2026-10-20T09:18:00+09:00','bars':bars}
   capture=json.dumps(evidence).encode();capture_digest=hashlib.sha256(capture).hexdigest();(source/'captures'/f'{capture_digest}.json').write_bytes(capture)
   common={'kind':'judgment','state_rev':'test','market':'KR','at':state['observed_at'],'symbol':'A','state_digest':digest,'screen':{'finalist':True},'horizon_minutes':60,'label_due_at':'2026-10-20T10:18:00+09:00','base_price':100}
   one=dict(common,id='1',question_id='j1_up_within_horizon');two=dict(common,id='2',question_id='j2_hold_off_red_flag')
   labels=[{'kind':'label','judgment_id':ident,'due_at':common['label_due_at'],'at':'2026-10-20T10:18:01+09:00','realized_price':101,'realized_return':.01} for ident in ('1','2')]
   (source/'ledger.jsonl').write_text(''.join(json.dumps(x)+'\n' for x in [one,two,*labels]))
   fields=('completeness_rules','candidate_rule','cycle_schedule','interruption_rules','replacement_rules','interim_reporting_rules','stop_rules','analysis_version')
   plan={key:'fixed' for key in fields};plan['session_plan']={'KR':[{'date':'2026-10-20','open_at':'2026-10-20T09:00:00+09:00','close_at':'2026-10-20T15:30:00+09:00','cadence_seconds':60,'max_cycle_skew_seconds':2}], 'US':[{'date':'2026-10-20','open_at':'2026-10-20T09:30:00-04:00','close_at':'2026-10-20T16:00:00-04:00','cadence_seconds':60,'max_cycle_skew_seconds':2}]};plan.update(coverage_thresholds={'min_cycle_fraction':1,'min_mfi_fraction':1,'min_label_fraction':1},sources={'KR':{'name':'fixture','timezone':'Asia/Seoul'},'US':{'name':'fixture','timezone':'America/New_York'}},registered_at='2026-10-19T00:00:00+09:00',first_observation_not_before='2026-10-20T00:00:00+09:00',max_bar_age_seconds=60,max_label_delay_seconds=90,horizon_minutes=60)
   (root/'plan').write_text(json.dumps(plan));(root/'sidecar').write_text(json.dumps({'kind':'manifest','sha256':hashlib.sha256((root/'plan').read_bytes()).hexdigest()})+'\n'+json.dumps({'kind':'capture','state_digest':digest,'capture_digest':capture_digest})+'\n')
   command=[sys.executable,analysis.__file__,str(source),str(root/'plan'),str(root/'sidecar'),str(root/'output')]
   (root/'sitecustomize.py').write_text("import socket\ndef blocked(*a, **kw): raise RuntimeError('network blocked')\nsocket.socket.connect=blocked\nsocket.create_connection=blocked\n")
   env=dict(os.environ,PYTHONPATH=str(root))
   completed=subprocess.run(command,capture_output=True,text=True,env=env)
   self.assertEqual(completed.returncode,0,completed.stderr)
   output=json.loads((root/'output').read_text());group=output['groups'][0]
   self.assertEqual((group['candidates'],group['mfi_valid'],group['label_valid']),(1,1,1))
   original_ledger=(source/'ledger.jsonl').read_bytes()
   for finalist in (True,False):
    with self.subTest(finalist=finalist):
     invalid=[dict(one,state_digest='invalid',screen={'finalist':finalist}),dict(two,state_digest='invalid',screen={'finalist':finalist}),*labels]
     (source/'ledger.jsonl').write_text(''.join(json.dumps(x)+'\n' for x in invalid))
     command[-1]=str(root/('invalid-'+str(finalist)))
     completed=subprocess.run(command,capture_output=True,text=True,env=env)
     self.assertEqual(completed.returncode,0,completed.stderr)
     group=json.loads(Path(command[-1]).read_text())['groups'][0]
     self.assertEqual((group['selection'],group['mfi_excluded'],group['label_excluded']),('unknown',{'invalid_identity':1},{'invalid_identity':1}))
   (source/'ledger.jsonl').write_bytes(original_ledger)
   partial={'kind':'session','market':'KR','date':'2026-10-20','status':'complete'}
   with (root/'sidecar').open('a') as stream:stream.write(json.dumps(partial)+'\n')
   command[-1]=str(root/'partial');self.assertEqual(subprocess.run(command,capture_output=True).returncode,0)
   kr=next(x for x in json.loads((root/'partial').read_text())['sessions'] if x['market']=='KR')
   self.assertEqual((kr['status'],kr['supplied_status'],kr['observed_cycles']),('incomplete','complete',1))
   partial.update(status='partial',reason='fixture interruption')
   sidecar=(root/'sidecar').read_text().splitlines()
   (root/'sidecar').write_text('\n'.join(sidecar[:-1]+[json.dumps(partial)])+'\n')
   command[-1]=str(root/'interrupted');self.assertEqual(subprocess.run(command,capture_output=True).returncode,0)
   kr=next(x for x in json.loads((root/'interrupted').read_text())['sessions'] if x['market']=='KR')
   self.assertEqual((kr['status'],kr['supplied_status']),('incomplete','partial'))
   (root/'sidecar').write_text(json.dumps({'kind':'manifest','sha256':hashlib.sha256((root/'plan').read_bytes()).hexdigest()})+'\n');command[-1]=str(root/'output2')
   completed=subprocess.run(command,capture_output=True,text=True,env=env)
   self.assertEqual(completed.returncode,0,completed.stderr)
   self.assertEqual(json.loads((root/'output2').read_text())['groups'][0]['mfi_excluded'],{'unproven_asof':1})
   self.assertEqual(json.loads((root/'output2').read_text())['sessions'][1]['supplied_status'],'unknown')
   plan['session_plan']['KR'][0]['date']='2026-10-21';plan['session_plan']['KR'][0]['open_at']='2026-10-21T09:00:00+09:00';plan['session_plan']['KR'][0]['close_at']='2026-10-21T15:30:00+09:00'
   (root/'plan').write_text(json.dumps(plan));command[-1]=str(root/'output3')
   mismatch=subprocess.run(command,capture_output=True,text=True,env=env)
   self.assertNotEqual(mismatch.returncode,0)
   self.assertIn('sidecar manifest hash mismatch',mismatch.stderr)
   self.assertFalse((root/'output3').exists())
   (root/'sidecar').write_text(json.dumps({'kind':'manifest','sha256':hashlib.sha256((root/'plan').read_bytes()).hexdigest()})+'\n')
   self.assertNotEqual(subprocess.run(command,capture_output=True,env=env).returncode,0)
   self.assertFalse((root/'output3').exists())
   plan['registered_at']='2026-10-20T09:19:00+09:00'
   (root/'plan').write_text(json.dumps(plan))
   (root/'event').write_text(json.dumps({'kind':'capture','state_digest':digest,'capture_digest':capture_digest}))
   record=[sys.executable,analysis.__file__,'record',str(source),str(root/'plan'),str(root/'event'),str(root/'rejected-sidecar')]
   rejected=subprocess.run(record,capture_output=True,text=True,env=env)
   self.assertNotEqual(rejected.returncode,0)
   self.assertIn('observation predates registration',rejected.stderr)
   self.assertFalse((root/'rejected-sidecar').exists())
class SourceTest(unittest.TestCase):
 def test_capture_identity_and_asof(self):
  bars=MfiTest().bars()
  for b in bars:b['open']=b['close']
  state={'market':'KR','symbol':'A','intraday':{'bars_1m':[{'t':b['close_at'][11:16],'o':b['open'],'h':b['high'],'l':b['low'],'c':b['close'],'v':b['volume']} for b in bars]}}
  evidence={'state_digest':'digest','market':'KR','symbol':'A','observed_at':'2026-10-20T09:18:00+09:00','source':'fixture','source_timezone':'Asia/Seoul','received_at':'2026-10-20T09:18:00+09:00','bars':bars}
  source={'name':'fixture','timezone':'Asia/Seoul'}
  window={'open_at':'2026-10-20T09:00:00+09:00','close_at':'2026-10-20T15:30:00+09:00'}
  self.assertIsNone(analysis.source_bars(state,evidence,evidence['observed_at'],'digest',source,window)[1])
  for change,reason in [({'symbol':'B'},'source_mismatch'),({'source':'other'},'source_mismatch'),({'received_at':'2026-10-20T09:17:59+09:00'},'unproven_asof'),({'received_at':'2026-10-20T09:18:01+09:00'},'unproven_asof')]:
   with self.subTest(change=change):self.assertEqual(analysis.source_bars(state,dict(evidence,**change),evidence['observed_at'],'digest',source,window)[1],reason)
  prior=[dict(b,close_at=b['close_at'].replace('2026-10-20','2026-10-19')) for b in bars]
  self.assertEqual(analysis.source_bars(state,dict(evidence,bars=prior),evidence['observed_at'],'digest',source,window)[1],'outside_session')
if __name__=='__main__':unittest.main()

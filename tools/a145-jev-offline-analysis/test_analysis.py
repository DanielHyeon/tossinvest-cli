import importlib.util
import pathlib
import json
import hashlib
import tempfile
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location('analysis', pathlib.Path(__file__).with_name('analysis.py'))
a = importlib.util.module_from_spec(spec)
spec.loader.exec_module(a)

class AnalysisTests(unittest.TestCase):
    def test_output_overlap_rejected_including_symlink(self):
        with tempfile.TemporaryDirectory() as root:
            source = pathlib.Path(root) / 'source'; source.mkdir()
            link = pathlib.Path(root) / 'link'; link.symlink_to(source)
            with self.assertRaises(ValueError): a.validate_paths(source, link / 'out')
            with self.assertRaises(ValueError): a.validate_paths(source, pathlib.Path(root))
            other = pathlib.Path(root) / 'other'; other.mkdir()
            (source / 'states').symlink_to(other, target_is_directory=True)
            with self.assertRaisesRegex(ValueError, 'symlink'): a.validate_paths(source, other / 'out')

    def test_quality_preserves_unknown_empty_and_unsupported(self):
        self.assertEqual(a.quality({'market':'US','news':'present','news_headlines':{'symbol':[]}})['symbol_news'], 'empty')
        self.assertEqual(a.quality({'market':'KR','news':'missing'})['symbol_news'], 'unavailable')
        self.assertEqual(a.quality({'market':'US'})['stock_flows'], 'unsupported')
        self.assertEqual(a.quality({'market':'KR'})['symbol_news'], 'unknown')

    def test_missing_response_blocks_and_partial_cycle_is_excluded(self):
        cycle = [{'key':'one','j1':.6,'j2':.2,'return':.1,'rank':1}, {'key':'two','j1':.4,'j2':.7,'return':-.1,'rank':2}]
        self.assertEqual(a.r2_cycle(cycle, {}, 1)['status'], 'incomplete')
        self.assertEqual(a.r2_cycle(cycle, {'one':.2,'two':.8}, 2)['status'], 'insufficient_k')
        self.assertEqual(a.r2_cycle(cycle, {'one':.2,'two':.2}, 1)['selected'], ['one'])

    def test_future_or_unknown_news_is_not_safe(self):
        state={'market':'KR','observed_at':'2026-10-16T01:00:00Z','news':'present','news_headlines':{'symbol':[{'created_at':'2026-10-16T01:01:00Z'}], 'market':[{}]}}
        result=a.quality(state)
        self.assertIs(result['symbol_future_news'], True)
        self.assertIsNone(result['market_future_news'])

    def test_bad_label_due_and_bad_state_are_rejected(self):
        with tempfile.TemporaryDirectory() as root:
            p=pathlib.Path(root); (p/'states').mkdir()
            state={'market':'KR','symbol':'X','observed_at':'2026-10-16T01:00:00Z','base_price':100,'ranking_rank':1}
            data=json.dumps(state,separators=(',',':')).encode(); digest=hashlib.sha256(data).hexdigest()
            (p/'states'/f'{digest}.json').write_bytes(data+b'\n')
            common={'kind':'judgment','at':state['observed_at'],'market':'KR','symbol':'X','state_digest':'sha256:'+digest,'p':.4,'base_price':100,'horizon_minutes':60,'label_due_at':'2026-10-16T02:00:00Z'}
            rows=[dict(common,id='one',question_id=a.J1),dict(common,id='two',question_id=a.J2),{'kind':'label','judgment_id':'one','due_at':'2026-10-16T01:10:00Z','at':'2026-10-16T02:01:00Z','realized_price':101,'realized_return':.01}]
            ledger=p/'ledger.jsonl'
            duplicate=dict(rows[0],id='different-id')
            ledger.write_text(''.join(json.dumps(x)+'\n' for x in rows[:2]+[duplicate]))
            with self.assertRaisesRegex(ValueError,'duplicate question'): a.read_snapshot(p)
            ledger.write_text(''.join(json.dumps(x)+'\n' for x in rows))
            with self.assertRaisesRegex(ValueError,'timing'): a.read_snapshot(p)
            rows[0]['horizon_minutes']=30
            ledger.write_text(''.join(json.dumps(x)+'\n' for x in rows))
            with self.assertRaisesRegex(ValueError,'60-minute'): a.read_snapshot(p)
            rows[0]['horizon_minutes']=60
            rows[-1]['due_at']='2026-10-16T02:00:00Z'
            rows[0]['model_version']='v1'; rows[1]['model_version']='v2'
            ledger.write_text(''.join(json.dumps(x)+'\n' for x in rows))
            with self.assertRaisesRegex(ValueError,'paired provenance'): a.read_snapshot(p)
            rows[1]['model_version']='v1'
            rows[-1]['label_gap']='missing'
            ledger.write_text(''.join(json.dumps(x)+'\n' for x in rows))
            with self.assertRaisesRegex(ValueError,'gap'): a.read_snapshot(p)
            del rows[-1]['label_gap']
            other=dict(rows[-1],judgment_id='two',realized_price=102,realized_return=.02); rows.append(other)
            ledger.write_text(''.join(json.dumps(x)+'\n' for x in rows))
            with self.assertRaisesRegex(ValueError,'labels disagree'): a.read_snapshot(p)
            (p/'states'/f'{digest}.json').write_bytes(b'{}\n')
            with self.assertRaisesRegex(ValueError,'digest'): a.read_snapshot(p)

    def test_missing_labels_exclude_entire_cycle(self):
        cycle=[{'key':'one','j1':.6,'j2':.2,'return':.1,'rank':1,'screen':{'finalist':True,'red_flag_excluded':False},'quality':{}}, {'key':'two','j1':.4,'j2':.7,'return':None,'rank':2,'screen':{'finalist':False,'red_flag_excluded':True},'quality':{}}]
        reports,_=a.summarize({('s3','KR','2026-10-16','t','rule'):cycle})
        self.assertEqual(reports[0]['incomplete_cycles'],1)
        self.assertIsNone(reports[0]['current_mean_pct'])

    def test_r2_synthetic_provenance_and_j1_are_fixed(self):
        quality={'market_future_news':False,'symbol_future_news':False}
        def candidate(key, j1, j2, rank, finalist):
            return {'key':key,'j1':j1,'j2':j2,'rank':rank,'return':.1*rank,'screen':{'finalist':finalist},'quality':quality,'model':'m','model_version':'v1','model_requested':'m','request_controls_sha256':'controls','j1_digest':'j1','question_rev':'r1','observed_at':'t'}
        cycle=[candidate('a',.9,.1,1,True),candidate('b',.2,.9,2,False)]
        def response(key,p):
            return {'state_digest':key,'p':p,'model_answered':'m','model_version':'v1','model_requested':'m','request_controls_sha256':'controls','request_sha256':'req','j1_question_digest':'j1','j1_question_rev':'r1','observed_at':'t'}
        with tempfile.TemporaryDirectory() as root:
            path=pathlib.Path(root)/'r2.json'; bundle={'evidence_kind':'synthetic','question':a.R2,'responses':[response('a',.1),response('b',.1)]};path.write_text(json.dumps(bundle))
            cycles={('s3','KR','2026-10-16','t','rule'):cycle}
            result=a.compare_r2(cycles,path)
            self.assertEqual(result['status'],'FIXTURE_ONLY')
            self.assertAlmostEqual(result['cycles'][0]['r2_mean_pct'],10)
            bundle['responses'][0]['model_version']='wrong';path.write_text(json.dumps(bundle))
            self.assertEqual(a.compare_r2(cycles,path)['cycles'][0]['status'],'incomplete')
            bundle['responses'][0]['model_version']='v1'; del cycle[0]['model_version'];cycle[0]['model_version']=None;path.write_text(json.dumps(bundle))
            self.assertEqual(a.compare_r2(cycles,path)['status'],'BLOCKED')

    def test_recorded_requires_anchored_request_and_unchanged_j1(self):
        q1={'type':'noul','instructions':'up?','criteria':{'true':'yes','false':'no'}}
        q2={'type':'noul','instructions':'red?','criteria':{'true':'yes','false':'no'}}
        state={'observed_at':'2026-10-16T01:00:00Z','symbol':'X'}
        base={'state':state,'model':'v1','questions':{a.J1:q1,a.J2:q2}}
        changed={'state':state,'model':'v1','questions':{a.J1:q1,a.J2:a.R2}}
        raw=json.dumps(base,separators=(',',':')); row={'baseline_request_json':raw,'r2_request_json':json.dumps(changed),'model_version':'v1','r2_response_json':json.dumps({'model':'v1','answers':{a.J2:{'type':'noul','noul':.2}}})}
        digest=lambda q:'sha256:'+hashlib.sha256(json.dumps(q,ensure_ascii=False,separators=(',',':')).encode()).hexdigest()
        c={'baseline_request_sha256':hashlib.sha256(raw.encode()).hexdigest(),'model':'v1','model_version':'v1','model_requested':'v1','j1_digest':digest(q1),'j2_digest':digest(q2),'observed_at':state['observed_at'],'state':state}
        self.assertEqual(a.verified_recorded_response(c,row),.2)
        c.update({'key':'a','j1':.9,'j2':.1,'rank':1,'return':.1,
                  'screen':{'finalist':True},'quality':{'market_future_news':False,'symbol_future_news':False}})
        row['state_digest']='a'
        with tempfile.TemporaryDirectory() as root:
            path=pathlib.Path(root)/'r2.json'
            bundle={'evidence_kind':'recorded','question':a.R2,'responses':[row]}
            path.write_text(json.dumps(bundle))
            cycles={('s3','KR','2026-10-16','t','rule'):[c]}
            self.assertEqual(a.compare_r2(cycles,path)['status'],'AVAILABLE')
            other=dict(c,key='b',rank=2,screen={'finalist':False})
            cycles[('s3','KR','2026-10-16','t','rule')].append(other)
            self.assertEqual(a.compare_r2(cycles,path)['cycles'][0]['status'],'incomplete')
        del row['model_version']
        self.assertIsNone(a.verified_recorded_response(c,row))
        row['model_version']='v1'
        changed['questions'][a.J1]={'type':'noul','instructions':'different'};row['r2_request_json']=json.dumps(changed)
        self.assertIsNone(a.verified_recorded_response(c,row))
        c['baseline_request_sha256']=None
        self.assertIsNone(a.verified_recorded_response(c,row))

    def test_temporal_rejects_future_and_outcomes(self):
        state={'observed_at':'2026-10-16T01:00:00Z','daily':{'last_prior_bar_at':'2026-10-16T00:59:00Z'}}
        self.assertTrue(a.temporal_clean(state))
        state['daily']['last_prior_bar_at']='2026-10-16T01:01:00Z'
        self.assertFalse(a.temporal_clean(state))
        state['daily']={'realized_return':.2}
        self.assertFalse(a.temporal_clean(state))

    def test_zero_k_and_missing_label_both_counted(self):
        cycle=[{'key':'a','j1':.2,'j2':.8,'return':None,'rank':1,'screen':{'finalist':False,'red_flag_excluded':True},'quality':{}}]
        rows,_=a.summarize({('s3','KR','2026-10-16','t','rule'):cycle})
        self.assertEqual(rows[0]['zero_k_cycles'],1)
        self.assertEqual(rows[0]['incomplete_cycles'],1)

    def test_equal_j1_uses_rank_with_k_below_ten(self):
        cs=[{'key':'late','j1':.5,'j2':.1,'return':.2,'rank':2,'screen':{'finalist':False,'red_flag_excluded':False},'quality':{}},
            {'key':'early','j1':.5,'j2':.1,'return':.1,'rank':1,'screen':{'finalist':True,'red_flag_excluded':False},'quality':{}}]
        rows,_=a.summarize({('s3','KR','2026-10-16','t','rule'):cs})
        self.assertEqual(rows[0]['complete_cycles'],1)
        self.assertAlmostEqual(rows[0]['current_mean_pct'],10)
        self.assertEqual(a.r2_cycle(cs,{'early':.1,'late':.1},1)['selected'],['early'])

    def test_duplicate_and_orphan_records_fail(self):
        with tempfile.TemporaryDirectory() as root:
            p=pathlib.Path(root); (p/'states').mkdir(); ledger=p/'ledger.jsonl'
            j={'kind':'judgment','id':'one','question_id':a.J1,'p':.5,'base_price':1,'horizon_minutes':60}
            ledger.write_text(json.dumps(j)+'\n'+json.dumps(j)+'\n')
            with self.assertRaisesRegex(ValueError,'duplicate judgment'): a.read_snapshot(p)
            label={'kind':'label','judgment_id':'unknown'}
            ledger.write_text(json.dumps(label)+'\n')
            with self.assertRaisesRegex(ValueError,'orphan'): a.read_snapshot(p)
            ledger.write_text(json.dumps(label)+'\n'+json.dumps(label)+'\n')
            with self.assertRaisesRegex(ValueError,'duplicate label'): a.read_snapshot(p)

    def test_valid_cli_runs_with_socket_disabled(self):
        with tempfile.TemporaryDirectory() as root:
            p=pathlib.Path(root); source=p/'snapshot'; source.mkdir(); (source/'states').mkdir()
            state={'market':'KR','symbol':'X','observed_at':'2026-10-16T01:00:00Z','base_price':100,'ranking_rank':1,'news':'absent'}
            raw=json.dumps(state,separators=(',',':')).encode(); digest=hashlib.sha256(raw).hexdigest()
            (source/'states'/f'{digest}.json').write_bytes(raw+b'\n')
            screen={'finalist':True,'red_flag_excluded':False,'screen_rank':1}
            common={'kind':'judgment','at':state['observed_at'],'market':'KR','symbol':'X','state_digest':'sha256:'+digest,'p':.2,'base_price':100,'horizon_minutes':60,'label_due_at':'2026-10-16T02:00:00Z','screen':screen}
            rows=[dict(common,id='one',question_id=a.J1),dict(common,id='two',question_id=a.J2)]
            rows.extend({'kind':'label','judgment_id':i,'due_at':'2026-10-16T02:00:00Z','at':'2026-10-16T02:00:01Z','realized_price':101,'realized_return':.01} for i in ('one','two'))
            (source/'ledger.jsonl').write_text(''.join(json.dumps(x)+'\n' for x in rows))
            output=p/'out'
            with patch('socket.socket', side_effect=AssertionError('network forbidden')), patch('sys.argv',['analysis',str(source),str(output)]): a.main()
            self.assertEqual(json.loads((output/'r2.json').read_text())['status'],'BLOCKED')
            self.assertEqual(json.loads((output/'sessions.json').read_text())[0]['complete_cycles'],1)

if __name__ == '__main__': unittest.main()

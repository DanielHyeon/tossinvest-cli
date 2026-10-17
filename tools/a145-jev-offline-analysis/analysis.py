#!/usr/bin/env python3
"""Offline, source-preserving analysis of a128 shadow snapshots."""
import argparse
import collections
import datetime as dt
import hashlib
import json
import math
import pathlib
import statistics
import sys
from zoneinfo import ZoneInfo

J1 = 'j1_up_within_horizon'
J2 = 'j2_hold_off_red_flag'
R2 = {'type': 'noul', 'instructions': 'Does this state show a clear, explicit adverse signal relevant to buying this stock now? Distinguish an unavailable news feed from a successfully read feed with no matching symbol headlines. Do not treat missing information by itself as an adverse signal, and do not infer safety or sufficient information from missing data.', 'criteria': {'true': 'An explicit adverse signal relevant to buying now is visible in the supplied state.', 'false': 'No such explicit adverse signal is visible in the supplied state.'}}
REPO = pathlib.Path(__file__).resolve().parents[2]


def validate_paths(source, output):
    source, output = source.resolve(strict=True), output.resolve()
    if (source / 'states').is_symlink():
        raise ValueError('states directory symlink rejected')
    if source == output or source in output.parents or output in source.parents or output == REPO or REPO in output.parents:
        raise ValueError('output must be outside input and repository trees')
    if output.exists():
        raise ValueError('output already exists')
    return source, output


def number(v):
    return isinstance(v, (int, float)) and not isinstance(v, bool) and math.isfinite(v)


def timestamp(v):
    t = dt.datetime.fromisoformat(v.replace('Z', '+00:00'))
    if t.tzinfo is None: raise ValueError('timestamp missing timezone')
    return t


def quality(state):
    news = state.get('news'); heads = state.get('news_headlines') or {}
    def news_kind(key):
        if news not in ('present', 'absent', 'missing', 'unavailable'): return 'unknown'
        if news != 'present': return 'unavailable'
        v = heads.get(key)
        return 'unknown' if not isinstance(v, list) else ('present' if v else 'empty')
    result = {'symbol_news': news_kind('symbol'), 'market_news': news_kind('market')}
    for name, field in [('intraday','intraday'), ('daily','daily'), ('quote','top_of_book'), ('market_context','market_context'), ('stock_flows','investor_flows_per_stock')]:
        missing = state.get('missing_inputs') or []
        aliases = {'intraday':'minute_bars:', 'daily':'daily_bars:', 'quote':'top_of_book:', 'market_context':'market_context:', 'stock_flows':'investor_flows_per_stock:'}
        result[name] = 'unsupported' if name == 'stock_flows' and state.get('market') == 'US' else ('present' if state.get(field) is not None else ('missing' if any(str(x).startswith(aliases[name]) for x in missing) else 'unknown'))
    observed = timestamp(state['observed_at']) if state.get('observed_at') else None
    for key in ('symbol', 'market'):
        ages = []; unknown = False
        for item in heads.get(key, []) if isinstance(heads.get(key), list) else []:
            try: ages.append((observed - timestamp(item['created_at'])).total_seconds())
            except (KeyError, ValueError, TypeError): unknown = True
        result[key + '_news_age_seconds'] = min(ages) if ages else None
        result[key + '_future_news'] = None if unknown else any(age < 0 for age in ages)
        age = result[key + '_news_age_seconds']
        result[key + '_news_freshness'] = 'unknown' if unknown or age is None else ('future' if age < 0 else ('under_1h' if age < 3600 else ('1h_to_24h' if age < 86400 else 'at_least_24h')))
    return result


def r2_cycle(cycle, responses, k):
    if any(c['key'] not in responses for c in cycle): return {'status':'incomplete'}
    survivors = [c for c in cycle if responses[c['key']] < .5]
    if len(survivors) < k: return {'status':'insufficient_k', 'survivors':len(survivors)}
    if not k: return {'status':'zero_k'}
    chosen = sorted(survivors, key=lambda c:(-c['j1'], c['rank']))[:k]
    return {'status':'ok', 'selected':[c['key'] for c in chosen]}


def read_snapshot(source):
    ledger_path = source/'ledger.jsonl'
    if ledger_path.is_symlink(): raise ValueError('ledger symlink rejected')
    raw = ledger_path.read_bytes()
    judgments, labels = {}, {}
    for line in raw.splitlines():
        x = json.loads(line)
        if x.get('kind') == 'judgment':
            if x['id'] in judgments: raise ValueError('duplicate judgment id')
            if x.get('question_id') not in (J1,J2) or not number(x.get('p')) or not 0 <= x['p'] <= 1 or not number(x.get('base_price')) or x['base_price'] <= 0: raise ValueError('invalid judgment')
            if x.get('horizon_minutes') != 60:
                raise ValueError('only 60-minute horizons are supported')
            judgments[x['id']] = x
        elif x.get('kind') == 'label':
            if x['judgment_id'] in labels: raise ValueError('duplicate label')
            labels[x['judgment_id']] = x
    if labels.keys() - judgments.keys(): raise ValueError('orphan label')
    for ident, label in labels.items():
        j = judgments[ident]
        if j.get('horizon_minutes') != 60:
            raise ValueError('only 60-minute horizons are supported')
        due = timestamp(j['label_due_at'])
        expected_due = timestamp(j['at']) + dt.timedelta(minutes=j['horizon_minutes'])
        if due != expected_due or timestamp(label['due_at']) != due or timestamp(label['at']) < due:
            raise ValueError('label timing mismatch')
        has_value = label.get('realized_price') is not None or label.get('realized_return') is not None
        if label.get('label_gap') and has_value:
            raise ValueError('label gap has realized value')
        if not label.get('label_gap'):
            price, value = label.get('realized_price'), label.get('realized_return')
            if not number(price) or price <= 0 or not number(value) or abs(value-(price-j['base_price'])/j['base_price']) > 5.1e-9: raise ValueError('invalid label')
    states, pairs = {}, collections.defaultdict(dict)
    for x in judgments.values():
        digest = x['state_digest'].removeprefix('sha256:')
        if len(digest) != 64 or any(c not in '0123456789abcdef' for c in digest): raise ValueError('invalid state digest')
        if digest not in states:
            state_path = source/'states'/f'{digest}.json'
            if state_path.is_symlink(): raise ValueError('state symlink rejected')
            data = state_path.read_bytes()
            if not data.endswith(b'\n') or hashlib.sha256(data[:-1]).hexdigest() != digest: raise ValueError('state digest mismatch')
            states[digest] = json.loads(data)
        state = states[digest]
        if state.get('market') != x['market'] or state.get('symbol') != x['symbol'] or timestamp(state['observed_at']) != timestamp(x['at']) or state.get('base_price') != x['base_price']: raise ValueError('state identity mismatch')
        key = (x.get('state_rev') or 'legacy', x['market'], x['at'], x['symbol'], digest)
        if x['question_id'] in pairs[key]: raise ValueError('duplicate question in state')
        pairs[key][x['question_id']] = x
    cycles = collections.defaultdict(list)
    for key, pair in pairs.items():
        if set(pair) != {J1,J2}: raise ValueError('unpaired questions')
        one, two = pair[J1], pair[J2]; state = states[key[-1]]
        if one.get('screen') != two.get('screen') or any(one.get(f) != two.get(f) for f in ('model_answered','model_version','model_requested','question_rev','label_due_at','horizon_minutes','selection')): raise ValueError('inconsistent paired provenance')
        label = labels.get(one['id']); other_label = labels.get(two['id']); value = None
        label_fields = ('realized_price', 'realized_return', 'label_gap')
        if label and other_label and any(label.get(f) != other_label.get(f) for f in label_fields):
            raise ValueError('paired labels disagree')
        if not other_label: label = None
        if label and not label.get('label_gap'):
            price, value = label.get('realized_price'), label.get('realized_return')
            if not number(price) or price <= 0 or not number(value) or abs(value-(price-one['base_price'])/one['base_price']) > 5.1e-9 or timestamp(label['at']) < timestamp(label['due_at']): raise ValueError('invalid label')
        screen = one.get('screen')
        if screen is None: continue  # Legacy lots have no comparable screening.
        rank = state.get('ranking_rank')
        if not isinstance(rank,int) or rank <= 0: raise ValueError('invalid ranking')
        day = timestamp(one['at']).astimezone(ZoneInfo('Asia/Seoul' if one['market']=='KR' else 'America/New_York')).date().isoformat()
        cycle = (key[0], key[1], day, one['at'], (one.get('selection') or {}).get('rule','unknown'), (one.get('selection') or {}).get('seed'), state.get('session_source'), one.get('model_answered'), one.get('model_requested'), one.get('question_rev'), one.get('question_digest'), two.get('question_digest'), one.get('model_version') or 'unknown')
        cycles[cycle].append({'key':key[-1], 'j1':one['p'], 'j2':two['p'], 'rank':rank, 'return':value, 'screen':screen, 'quality':quality(state), 'state':state, 'model':one.get('model_answered'), 'model_version':one.get('model_version'), 'request_controls_sha256':one.get('request_controls_sha256'), 'model_requested':one.get('model_requested'), 'j1_digest':one.get('question_digest'), 'j2_digest':two.get('question_digest'), 'baseline_request_sha256':one.get('request_sha256'), 'question_rev':one.get('question_rev'), 'observed_at':one['at']})
    observation_times=sorted(timestamp(x['at']) for x in judgments.values())
    manifest = {'analysis_at_utc':dt.datetime.now(dt.timezone.utc).isoformat(), 'snapshot_path':str(source), 'observation_min_utc':observation_times[0].isoformat() if observation_times else None, 'observation_max_utc':observation_times[-1].isoformat() if observation_times else None, 'ledger_sha256': hashlib.sha256(raw).hexdigest(), 'state_count':len(states), 'judgments':len(judgments), 'labels':len(labels), 'state_manifest_sha256':hashlib.sha256(''.join(d+'\n' for d in sorted(states)).encode()).hexdigest()}
    return cycles, manifest


def summarize(cycles):
    sessions = collections.defaultdict(lambda: {'cycles':0, 'excluded_incomplete':0, 'zero_k':0, 'current':[], 'j1_only':[], 'survivor_uniform':[], 'pool_uniform':[]})
    strata = collections.defaultdict(lambda: {'states':0,'excluded':0,'selected':0,'valid':0,'missing_label':0,'returns':[]})
    for key, cs in cycles.items():
        session=key[:3]+key[4:5]+key[6:]; out=sessions[session]; out['cycles']+=1
        for c in cs:
            for name,value in c['quality'].items():
                if name.endswith('_seconds'): continue
                q=strata[(*session,name,str(value))]; q['states']+=1; q['excluded']+=bool(c['screen']['red_flag_excluded']); q['selected']+=bool(c['screen']['finalist'])
                if c['return'] is None: q['missing_label']+=1
                else: q['valid']+=1; q['returns'].append(c['return']*100)
        k=sum(bool(c['screen']['finalist']) for c in cs)
        incomplete = any(c['return'] is None for c in cs)
        if not k: out['zero_k'] += 1
        if incomplete: out['excluded_incomplete'] += 1
        if not k or incomplete: continue
        survivors=[c for c in cs if c['j2'] < .5]
        expected=sorted(survivors,key=lambda c:(-c['j1'],c['rank']))[:k]
        selected=[c for c in cs if c['screen']['finalist']]
        if {c['key'] for c in expected} != {c['key'] for c in selected}: raise ValueError('screen replay mismatch')
        j1=sorted(cs,key=lambda c:(-c['j1'],c['rank']))[:k]
        for name,items in [('current',selected),('j1_only',j1),('survivor_uniform',survivors),('pool_uniform',cs)]: out[name].append(statistics.mean(c['return']*100 for c in items))
    def avg(values): return statistics.mean(values) if values else None
    reports=[]
    for key,v in sorted(sessions.items()):
        reports.append({'lot':key[0],'market':key[1],'local_session':key[2], 'provenance':list(key[3:]), 'cycles':v['cycles'],'incomplete_cycles':v['excluded_incomplete'],'zero_k_cycles':v['zero_k'], 'complete_cycles':len(v['current']), **{name+'_mean_pct':avg(v[name]) for name in ('current','j1_only','survivor_uniform','pool_uniform')}})
    quality_rows=[]
    for key,v in sorted(strata.items()):
        quality_rows.append({'lot':key[0],'market':key[1],'local_session':key[2],'provenance':list(key[3:-2]),'field':key[-2],'class':key[-1], 'states':v['states'],'exclusion_rate':v['excluded']/v['states'],'selection_rate':v['selected']/v['states'],'missing_labels':v['missing_label'],'row_weighted_mean_pct':avg(v['returns'])})
    return reports,quality_rows


def temporal_clean(state):
    try: observed=timestamp(state['observed_at'])
    except (KeyError, TypeError, ValueError): return False
    forbidden={'label','labels','label_due_at','realized_price','realized_return','future_return','due_at'}
    def walk(value):
        if isinstance(value,dict):
            for k,v in value.items():
                if k in forbidden: return False
                if (k.endswith('_at') or k in ('timestamp','time')) and v is not None:
                    try:
                        if timestamp(v)>observed: return False
                    except (ValueError,TypeError,AttributeError): return False
                if not walk(v): return False
        elif isinstance(value,list): return all(walk(v) for v in value)
        return True
    return walk(state)


def verified_recorded_response(c,row):
    # The original ledger must anchor the exact captured request bytes; a self-reported hash is insufficient.
    anchor=c.get('baseline_request_sha256')
    if not temporal_clean(c['state']): return None
    required = (anchor, c.get('model'), c.get('model_version'), c.get('model_requested'),
                c.get('j1_digest'), c.get('j2_digest'), row.get('model_version'))
    if not all(required) or row['model_version'] != c['model_version']:
        return None
    try:
        baseline_raw=row['baseline_request_json'].encode('utf-8'); revised_raw=row['r2_request_json'].encode('utf-8')
        if hashlib.sha256(baseline_raw).hexdigest() != anchor.removeprefix('sha256:'): return None
        baseline=json.loads(baseline_raw); revised=json.loads(revised_raw); response=json.loads(row['r2_response_json'])
        if baseline.get('model') != c['model_requested'] or response.get('model') != c['model']: return None
        if baseline.get('state') != c['state'] or baseline.get('state') != revised.get('state'):
            return None
        if baseline.get('model') != revised.get('model'):
            return None
        if baseline.get('state',{}).get('observed_at') != c['observed_at']: return None
        bq=baseline.get('questions',{}); rq=revised.get('questions',{})
        if set(bq)!={J1,J2} or set(rq)!={J1,J2} or bq[J1]!=rq[J1] or rq[J2]!=R2: return None
        for question,digest in ((bq[J1],c['j1_digest']),(bq[J2],c['j2_digest'])):
            raw=json.dumps(question,ensure_ascii=False,separators=(',',':')).encode('utf-8')
            if hashlib.sha256(raw).hexdigest() != digest.removeprefix('sha256:'): return None
        baseline_controls = {k:v for k,v in baseline.items() if k != 'questions'}
        revised_controls = {k:v for k,v in revised.items() if k != 'questions'}
        if baseline_controls != revised_controls:
            return None
        answer=response['answers'][J2]
        p=answer['noul']
        if answer.get('type') != 'noul' or not number(p) or not 0 <= p <= 1: return None
        return p
    except (KeyError,TypeError,ValueError,json.JSONDecodeError): return None


def compare_r2(cycles, path):
    if path is None: return {'status':'BLOCKED','reason':'No compatible saved r2 responses supplied.', 'question':R2}
    bundle=json.loads(path.read_text())
    if bundle.get('question') != R2 or not isinstance(bundle.get('responses'), list) or bundle.get('evidence_kind') not in ('synthetic','recorded'): raise ValueError('r2 question, evidence kind or response schema mismatch')

    responses={}
    for row in bundle['responses']:
        digest=row['state_digest'].removeprefix('sha256:')
        if digest in responses: raise ValueError('duplicate r2 response')
        if bundle['evidence_kind']=='synthetic' and (not number(row.get('p')) or not 0 <= row['p'] <= 1): raise ValueError('invalid r2 probability')
        responses[digest]=row
    reports=[]
    for key,cs in cycles.items():
        k=sum(bool(c['screen']['finalist']) for c in cs)
        matching={}
        for c in cs:
            row=responses.get(c['key'])
            if row and c['quality']['market_future_news'] is False and c['quality']['symbol_future_news'] is False:
                if bundle['evidence_kind'] == 'synthetic':
                    required=(c['model'], c['model_version'], c['model_requested'], c['j1_digest'], c['question_rev'], c['request_controls_sha256'], row.get('model_version'), row.get('request_sha256'), row.get('request_controls_sha256'))
                    if all(required) and row.get('model_version') == c['model_version'] and row.get('request_controls_sha256') == c['request_controls_sha256'] and row.get('model_answered') == c['model'] and row.get('model_requested') == c['model_requested'] and row.get('j1_question_digest') == c['j1_digest'] and row.get('j1_question_rev') == c['question_rev'] and row.get('observed_at') == c['observed_at']:
                        matching[c['key']]=row['p']
                else:
                    value=verified_recorded_response(c,row)
                    if value is not None: matching[c['key']]=value
        result=r2_cycle(cs,matching,k)
        if result['status']=='ok' and any(c['return'] is None for c in cs): result={'status':'missing_label'}
        if result['status']=='ok':
            selected=set(result.pop('selected'))
            result['r2_mean_pct']=statistics.mean(c['return']*100 for c in cs if c['key'] in selected)
            result['r1_mean_pct']=statistics.mean(c['return']*100 for c in cs if c['screen']['finalist'])
        reports.append({'lot':key[0], 'market':key[1], 'local_session':key[2], 'cycle':key[3], 'provenance':list(key[4:]), 'original_k':k, **result})
    return {'status':('FIXTURE_ONLY' if bundle['evidence_kind']=='synthetic' else 'AVAILABLE') if any(x['status']=='ok' for x in reports) else 'BLOCKED', 'question':R2, 'cycles':reports, 'interpretation':'fixed J1 sensitivity only; supplied-record consistency does not establish model-call or response authenticity; not joint r2 performance or causal evidence'}


def main():
    parser=argparse.ArgumentParser(description=__doc__); parser.add_argument('snapshot',type=pathlib.Path); parser.add_argument('output',type=pathlib.Path); parser.add_argument('--r2-responses',type=pathlib.Path, help='offline JSON response bundle')
    args=parser.parse_args(); source,out=validate_paths(args.snapshot,args.output)
    cycles,manifest=read_snapshot(source); sessions,strata=summarize(cycles)
    comparison = compare_r2(cycles, args.r2_responses)
    sidecar = [{'state_digest':c['key'], 'lot':key[0], 'market':key[1], 'local_session':key[2], 'quality':c['quality']} for key,cs in cycles.items() for c in cs]
    # Write only after complete validation, and never overwrite an existing output.
    out.mkdir(parents=True,exist_ok=False)
    for name,obj in [('manifest.json',manifest),('sessions.json',sessions),('quality.json',strata),('sidecar.json',sidecar),('r2.json',comparison)]:
        (out/name).write_text(json.dumps(obj,ensure_ascii=False,indent=2)+'\n')
    print('Wrote offline results; r2 status:', comparison['status'])


if __name__ == '__main__':
    try: main()
    except (OSError,ValueError,KeyError,TypeError,json.JSONDecodeError) as exc:
        print(f'error: {exc}',file=sys.stderr); sys.exit(1)

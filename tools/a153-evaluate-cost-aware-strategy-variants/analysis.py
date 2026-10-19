"""Offline exploratory independent-filter comparison; no collection or trading."""
import argparse
import copy
import datetime as dt
from fractions import Fraction
import hashlib
import importlib.util
import json
import math
import os
from pathlib import Path
import random
import re
import sys

ROOT = Path(__file__).resolve().parents[1]
def load(name, path):
    spec = importlib.util.spec_from_file_location(name, path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module

stop = load('a151', ROOT / 'a151-compare-stop-policies-offline/analysis.py')
mfi = load('a147', ROOT / 'a147-jev-mfi-offline/analysis.py')
TIME = re.compile(r'^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d{1,6})?(?:Z|[+-](?:[01]\d|2[0-3]):[0-5]\d)$')
FILTERS = ('candle', 'volume', 'mfi')
SOURCES = (ROOT / 'a151-compare-stop-policies-offline/analysis.py', ROOT / 'a147-jev-mfi-offline/analysis.py', ROOT / 'a145-jev-offline-analysis/analysis.py')

def timestamp(value):
    if not isinstance(value, str) or not TIME.fullmatch(value): raise ValueError('invalid timestamp')
    return dt.datetime.fromisoformat(value.replace('Z', '+00:00'))

def validate_values(value):
    if isinstance(value, str) and len(value) > 256: raise ValueError('string limit')
    if isinstance(value, float) and not math.isfinite(value): raise ValueError('nonfinite number')
    if isinstance(value, dict):
        for key, item in value.items():
            if not isinstance(key, str) or len(key) > 256: raise ValueError('invalid key')
            if key in ('signal_close_at','signal_confirmed_at','open_at','close_at','received_at','as_of','valid_from','valid_until'): timestamp(item)
            validate_values(item)
    elif isinstance(value, list):
        for item in value: validate_values(item)

def features(c):
    result = {name: None for name in FILTERS}
    try:
        end, observed = timestamp(c['signal_close_at']), timestamp(c['signal_confirmed_at'])
        if end > observed: return result
        def indexed(key):
            out = {}
            for bar in c[key]:
                t = timestamp(bar['close_at'])
                if t > end: continue
                if t in out: raise ValueError('duplicate bar')
                out[t] = bar
            return out
        def available(bar):
            closed, received = timestamp(bar['close_at']), timestamp(bar['received_at'])
            if not closed <= received <= observed: return False
            o,h,l,close,v = (stop.exact(bar[k]) for k in ('open','high','low','close','volume'))
            return l > 0 and l <= o <= h and l <= close <= h and v >= 0
        five = indexed('bars_5m')
        one = indexed('bars_1m')
        def coherent(bar):
            if not bar or not available(bar): return False
            close = timestamp(bar['close_at'])
            group = [one.get(close-dt.timedelta(minutes=i)) for i in range(4,-1,-1)]
            if not all(x is not None and available(x) for x in group): return False
            q = lambda b,k: stop.exact(b[k])
            actual = (q(bar,'open'),q(bar,'high'),q(bar,'low'),q(bar,'close'),q(bar,'volume'))
            expected = (q(group[0],'open'),max(q(x,'high') for x in group),min(q(x,'low') for x in group),q(group[-1],'close'),sum((q(x,'volume') for x in group),Fraction()))
            return actual == expected
        current, previous = five.get(end), five.get(end-dt.timedelta(minutes=5))
        if coherent(current):
            result['candle'] = stop.exact(current['close']) > stop.exact(current['open'])
            if coherent(previous): result['volume'] = stop.exact(current['volume']) > stop.exact(previous['volume'])
    except (ValueError, TypeError, KeyError, OverflowError):
        pass
    try:
        end, observed = timestamp(c['signal_close_at']), timestamp(c['signal_confirmed_at'])
        if end > observed: return result
        one = indexed('bars_1m')
        bars = [one.get(end-dt.timedelta(minutes=i)) for i in range(17,-1,-1)]
        if all(b is not None and available(b) for b in bars):
            numeric=[]
            for b in bars:
                item={'close_at':b['close_at']}
                for k in ('open','high','low','close','volume'):
                    value=float(stop.exact(b[k]))
                    if not math.isfinite(value): raise ValueError('invalid feature number')
                    item[k]=value
                numeric.append(item)
            value=mfi.mfi(numeric,c['signal_confirmed_at'],60)
            if 'direction' in value: result['mfi'] = value['direction'] == 'rising'
    except (ValueError, TypeError, KeyError, OverflowError, UnboundLocalError):
        pass
    return result

def evaluate(c):
    try: return stop.evaluate(c)
    except stop.CandidateError as e: return {'id':c['id'],'status':e.status,'reason':str(e)}
    except (ValueError, TypeError, KeyError, OverflowError) as e: return {'id':c['id'],'status':'invalid','reason':str(e)}

def mean(items): return sum(items,Fraction()) / len(items) if items else None

def bootstrap(values):
    if not values: return None
    rng=random.Random(153); n=len(values)
    draws=sorted(mean([values[rng.randrange(n)] for _ in range(n)]) for _ in range(2000))
    def percentile(p):
        index=Fraction(len(draws)-1)*p; low=index.numerator//index.denominator
        weight=index-low
        return draws[low]*(1-weight)+draws[min(low+1,len(draws)-1)]*weight
    return [stop.number(percentile(Fraction(25,1000))),stop.number(percentile(Fraction(975,1000)))]

def base_r(result):
    try:
        if result['status'] != 'eligible': return None
        arm=result['arms']['fixed']; outcome=arm['base']
        risk=Fraction(arm['risk'])
        if outcome['status'] != 'closed' or risk <= 0: return None
        return Fraction(outcome['pnl'])/risk
    except (KeyError, ValueError, TypeError, ZeroDivisionError): return None

def report_market(rows):
    sessions={}
    for row in rows: sessions.setdefault(row['date'],[]).append(row)
    accepted={}; excluded={}
    for date, items in sorted(sessions.items()):
        reasons=[]
        for item in items:
            for f in FILTERS:
                if item['filters'][f] is None: reasons.append(item['id']+':'+f+'_unknown')
            for scenario in ('base','stress'):
                if base_r(item[scenario]) is None:
                    execution = item[scenario]
                    if execution.get('status') == 'eligible':
                        execution = execution.get('arms', {}).get('fixed', {}).get('base', {})
                    status = str(execution.get('status', 'invalid'))
                    reason = str(execution.get('reason', 'invalid_or_zero_risk'))
                    reasons.append(item['id']+':'+scenario+':'+status+':'+reason)
        if reasons: excluded[date]={'ids':[x['id'] for x in items], 'reasons':reasons}
        else: accepted[date]=items
    summaries={}
    for scenario in ('base','stress'):
        summaries[scenario]={}
        for f in FILTERS:
            deltas=[]; baseline=[]; variant=[]; trades=[]
            for items in accepted.values():
                rs=[base_r(x[scenario]) for x in items]
                contributions=[r if x['filters'][f] else Fraction() for x,r in zip(items,rs)]
                baseline.append(mean(rs)); variant.append(mean(contributions)); deltas.append(mean([v-r for v,r in zip(contributions,rs)]))
                trades.extend(r for x,r in zip(items,rs) if x['filters'][f])
            selected=sum(x['filters'][f] is True for items in accepted.values() for x in items)
            total=sum(len(items) for items in accepted.values())
            summaries[scenario][f]={'baseline_mean_r':stop.number(mean(baseline)) if baseline else None,'variant_mean_contribution':stop.number(mean(variant)) if variant else None,'difference':stop.number(mean(deltas)) if deltas else None,'bootstrap_95':bootstrap(deltas),'selected_count':selected,'suppressed_count':total-selected,'conditional_trade_mean_r':stop.number(mean(trades)) if trades else None,'conditional_trade_count':len(trades)}
    counts={f:{str(v).lower() if v is not None else 'unknown':sum(x['filters'][f] is v for x in rows) for v in (True,False,None)} for f in FILTERS}
    return {'total_candidates':len(rows),'total_sessions':len(sessions),'filter_counts':counts,'included_sessions':{d:[x['id'] for x in items] for d,items in accepted.items()},'excluded_sessions':excluded,'complete_sessions':len(accepted),'sample_sufficiency':'insufficient' if len(accepted)<20 else 'not_established','summaries':summaries}

def compare(obj):
    stop.fields(obj,'version synthetic mode candidates')
    if obj['version']!='a153-v1' or obj['mode']!='independent_filters' or not isinstance(obj['synthetic'],bool): raise ValueError('invalid envelope')
    candidates=obj['candidates']
    if not isinstance(candidates,list) or len(candidates)>1000: raise ValueError('candidate limit')
    validate_values(obj); ids=set(); opportunities=set(); sessions={}; rows=[]
    for c in candidates:
        if not isinstance(c,dict): raise ValueError('invalid candidate')
        ident=stop.nonempty(c.get('id'))
        if ident in ids: raise ValueError('duplicate id')
        ids.add(ident)
        try:
            market,symbol,date=c['market'],c['symbol'],c['date']
            if market not in ('KR','US'): raise ValueError('unsupported market identity')
            stop.nonempty(symbol); stop.nonempty(date)
            key=(market,symbol,date,timestamp(c['signal_close_at']).astimezone(dt.timezone.utc))
            if key in opportunities: raise ValueError('duplicate opportunity')
            opportunities.add(key)
            session=c['session']; skey=(market,date)
            if skey in sessions and sessions[skey]!=session: raise ValueError('inconsistent session')
            sessions[skey]=session
        except (KeyError,TypeError): raise ValueError('missing identity or session')
        selected=features(c); normal=evaluate(c); stressed=copy.deepcopy(c)
        try:
            for k in ('entry_slippage_ticks','exit_slippage_ticks'):
                stressed[k]=stop.integer(stressed[k])*2
            stress=evaluate(stressed)
        except (ValueError,TypeError,KeyError) as e: stress={'id':ident,'status':'invalid','reason':str(e)}
        rows.append({'id':ident,'market':market,'symbol':symbol,'date':date,'filters':selected,'base':normal,'stress':stress})
    return {'version':'a153-v1','synthetic':obj['synthetic'],'provenance':'caller-supplied and unverified','profitability':'inconclusive','production_authorized':False,'limitations':['synthetic or unverified provenance','selection and complete-session conditioning','no authenticated holdout or equity curve','independence not established'],'deferred':['VP','SI','MA20','EMA exits','partial exits','portfolio MDD','MAE/MFE','turnover/exposure','regime inference','authenticated holdout','FX','policy promotion'],'source_sha256':{str(p.relative_to(ROOT)):hashlib.sha256(p.read_bytes()).hexdigest() for p in SOURCES},'candidates':rows,'markets':{m:report_market([x for x in rows if x['market']==m]) for m in ('KR','US')}}

def main():
    parser=argparse.ArgumentParser(description=__doc__); parser.add_argument('input'); parser.add_argument('--output'); args=parser.parse_args()
    try:
        if args.input == '-': data=sys.stdin.buffer.read(16*1024*1024+1)
        else:
            mfi.safe_output(args.input,args.output or str(Path(args.input).absolute().parent/'__a153_unused_output__'))
            with open(args.input,'rb') as f: data=f.read(16*1024*1024+1)
        if len(data)>16*1024*1024: raise ValueError('input exceeds 16 MiB')
        obj=json.loads(data,object_pairs_hook=stop.unique_pairs,parse_constant=lambda _: (_ for _ in ()).throw(ValueError('invalid constant')))
        result=compare(obj); result['input_sha256']=hashlib.sha256(data).hexdigest()
        encoded=(json.dumps(result,sort_keys=True,allow_nan=False)+'\n').encode()
        if args.output:
            if args.input == '-':
                for path in (Path(args.output).absolute(), *Path(args.output).absolute().parents):
                    if path.is_symlink(): raise ValueError('symlink path')
            stop.write_once(args.output,encoded)
        else: sys.stdout.buffer.write(encoded)
    except (ValueError,OSError,TypeError,KeyError,OverflowError,RecursionError) as e:
        print('a153: '+str(e),file=sys.stderr); return 1
    return 0
if __name__=='__main__': raise SystemExit(main())

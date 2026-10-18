"""Qualify caller-supplied profile evidence offline; never grants trading authority."""
import argparse
import datetime as dt
from fractions import Fraction
import hashlib
import json
import os
from pathlib import Path
import re
import sys
from zoneinfo import ZoneInfo

DECIMAL = re.compile(r'^[0-9]+(?:\.[0-9]+)?$')
TIME = re.compile(r'^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(?:\.[0-9]{1,6})?(?:Z|[+-](?:[01][0-9]|2[0-3]):[0-5][0-9])$')
KINDS = {'trades': ('exact_trades', 'executions'), 'bins': ('documented_price_bins', 'executed_price_bins'), 'ohlcv': ('ohlcv_proxy', 'ohlcv'), 'orderbook': ('unavailable', 'orderbook'), 'unavailable': ('unavailable', 'unknown')}

class Invalid(ValueError):
    def __init__(self, reason, status='invalid'):
        self.reason, self.status = reason, status
        super().__init__(reason)

def require(condition, reason, status='invalid'):
    if not condition: raise Invalid(reason, status)

def fields(obj, names):
    require(isinstance(obj, dict) and set(obj) == set(names.split()), 'invalid_fields')

def string(value):
    require(isinstance(value, str) and 0 < len(value) <= 512 and value.strip() == value, 'invalid_string')
    return value

def decimal(value):
    require(isinstance(value, str) and len(value) <= 64 and DECIMAL.fullmatch(value) is not None, 'invalid_decimal')
    return Fraction(value)

def stamp(value):
    try:
        value = string(value)
        require(TIME.fullmatch(value) is not None, 'invalid_timestamp')
        t = dt.datetime.fromisoformat(value.replace('Z', '+00:00'))
        require(t.tzinfo is not None and t.utcoffset() is not None, 'invalid_timestamp')
        return t
    except (ValueError, OverflowError):
        raise Invalid('invalid_timestamp')

def number(value):
    n, d = value.numerator, value.denominator
    whole, rem = divmod(n, d)
    if not rem: return str(whole)
    s = ''
    while rem:
        rem *= 10
        digit, rem = divmod(rem, d)
        s += str(digit)
    return str(whole) + '.' + s

def enums(value, choices):
    require(isinstance(value, str) and value in choices, 'invalid_enum')
    return value

def strings(xs, limit):
    require(isinstance(xs, list) and 0 < len(xs) <= limit, 'invalid_list')
    for x in xs: string(x)

def dataset(d, synthetic):
    classification = 'unavailable'
    try:
        fields(d, 'id kind market symbol currency basis profile_mode session cutoff_at as_of source coverage correction_policy price_increment records bin_method access')
        string(d['id']); string(d['symbol'])
        kind = enums(d['kind'], KINDS); classification, expected = KINDS[kind]
        market = enums(d['market'], ('KR', 'US'))
        require(d['currency'] == {'KR':'KRW', 'US':'USD'}[market] and d['basis'] == 'unadjusted', 'market_currency_or_basis', 'unsupported')
        session = d['session']; fields(session, 'date timezone open_at close_at')
        zone = {'KR':'Asia/Seoul','US':'America/New_York'}[market]
        require(session['timezone'] == zone, 'session_timezone')
        date = dt.date.fromisoformat(string(session['date']))
        op, close, cutoff, asof = stamp(session['open_at']), stamp(session['close_at']), stamp(d['cutoff_at']), stamp(d['as_of'])
        local = ZoneInfo(zone)
        for t in (op, close):
            converted=t.astimezone(local)
            require(converted.date()==date and t.utcoffset()==converted.utcoffset(), 'session_date_or_offset')
        require(op < close and close-op <= dt.timedelta(hours=24), 'invalid_session')
        mode=enums(d['profile_mode'], ('current_prefix','prior_session_fixed'))
        if mode == 'current_prefix': require(op < cutoff <= asof <= close, 'invalid_interval')
        else: require(cutoff == close and cutoff <= asof and date < asof.astimezone(local).date(), 'invalid_interval')
        source=d['source']; fields(source,'name version reference declared_kind')
        for k in ('name','version','reference'): string(source[k])
        enums(source['declared_kind'], ('executions','executed_price_bins','ohlcv','orderbook','unknown'))
        coverage=d['coverage']; fields(coverage,'start_at end_at status gaps sequence_status venues trade_types')
        start, end=stamp(coverage['start_at']),stamp(coverage['end_at'])
        require(start < end and op <= start <= end <= cutoff, 'invalid_coverage')
        enums(coverage['status'], ('complete','partial','unknown')); enums(coverage['sequence_status'], ('complete','gap','unknown'))
        strings(coverage['venues'],32); strings(coverage['trade_types'],32)
        gaps=coverage['gaps']; require(isinstance(gaps,list) and len(gaps)<=100,'invalid_gaps')
        for gap in gaps:
            fields(gap,'start_at end_at'); gs,ge=stamp(gap['start_at']),stamp(gap['end_at']); require(op<=gs<ge<=cutoff,'invalid_gap')
        access=d['access']; fields(access,'license availability cost_note'); string(access['license']); string(access['cost_note']); enums(access['availability'],('provided','restricted','unknown'))
        enums(d['correction_policy'],('none_declared','unresolved','unknown'))
        increment=decimal(d['price_increment']); require(increment>0,'invalid_increment')
        records=d['records']; require(isinstance(records,list),'invalid_records')
        method=d['bin_method']
        if kind=='bins':
            fields(method,'name volume_basis unit boundary width price_min price_max zero_bins_explicit built_at received_at')
            for k in ('name','volume_basis','unit','boundary','width','price_min','price_max','built_at','received_at'): string(method[k])
            require(type(method['zero_bins_explicit']) is bool,'invalid_bin_method')
        else: require(method is None,'invalid_bin_method')
        base={'id':d['id'],'classification':classification,'market':market,'symbol':d['symbol'],'currency':d['currency'],'basis':d['basis'],'session':session,'mode':mode,'requested_interval':{'start_at':session['open_at'],'end_at':d['cutoff_at']},'as_of':d['as_of'],'rows':len(records),'source':source,'coverage':coverage,'coverage_provenance':'caller_asserted_unverified','access':access,'correction_policy':d['correction_policy'],'price_increment':d['price_increment'],'bin_method':method}
        require(source['declared_kind']==expected,'source_kind_mismatch','unsupported')
        require(access['availability']=='provided','access_not_provided','unsupported')
        if kind in ('ohlcv','orderbook','unavailable'):
            require(not records and d['bin_method'] is None,'invalid_proxy_records')
            raise Invalid({'ohlcv':'ohlcv_is_not_execution_profile','orderbook':'book_size_is_not_executed_volume','unavailable':'no_profile_evidence'}[kind], 'unsupported')
        require(d['correction_policy']=='none_declared','corrections_unresolved','unsupported')
        require(start==op and end==cutoff and coverage['status']=='complete' and coverage['sequence_status']=='complete' and not gaps,'coverage_incomplete','not_ready')
        if kind=='trades':
            require(d['bin_method'] is None,'invalid_bin_method')
            seen=set(); last=None; total=Fraction()
            for record in records:
                fields(record,'event_id action reference_id executed_at received_at price volume')
                event=string(record['event_id']); require(event not in seen,'duplicate_event_id'); seen.add(event)
                action=enums(record['action'],('trade','correct','cancel'))
                if action != 'trade':
                    string(record['reference_id']); raise Invalid('correction_event_unsupported','unsupported')
                require(record['reference_id'] is None,'unexpected_reference')
                executed,received=stamp(record['executed_at']),stamp(record['received_at'])
                require(op<=executed<cutoff and executed<=received<=asof and (last is None or last<=received),'invalid_trade_time'); last=received
                price,volume=decimal(record['price']),decimal(record['volume'])
                require(price>0 and volume>0 and price%increment==0,'invalid_trade_value'); total+=volume
        else:
            require(method['volume_basis']=='executed' and method['unit']=='shares' and method['boundary']=='[lower,upper)' and method['zero_bins_explicit'] is True,'unsupported_bin_method','unsupported')
            width,minimum,maximum=(decimal(method[k]) for k in ('width','price_min','price_max'))
            require(width>0 and minimum>0 and maximum>minimum and width%increment==0 and minimum%increment==0 and maximum%increment==0,'invalid_bin_grid')
            built,received=stamp(method['built_at']),stamp(method['received_at']); require(cutoff<=built<=received<=asof,'invalid_bin_time')
            cursor=minimum; total=Fraction()
            for record in records:
                fields(record,'lower upper volume'); lo,hi,volume=(decimal(record[k]) for k in ('lower','upper','volume'))
                require(lo==cursor and hi-lo==width and hi<=maximum and lo%increment==0 and hi%increment==0,'invalid_bin_range'); cursor=hi; total+=volume
            if records: require(cursor==maximum,'invalid_bin_range')
        require(bool(records),'empty_evidence','not_ready')
        require(total>0,'no_executed_volume','not_ready')
        base.update(status='eligible_unverified',reasons=[],total_volume=number(total),vp_input_eligible=not synthetic,vp_exclusion_reasons=['synthetic_only'] if synthetic else [])
        return base
    except (Invalid, ValueError, TypeError, OverflowError, KeyError) as err:
        reason=err.reason if isinstance(err,Invalid) else 'invalid_dataset'
        status=err.status if isinstance(err,Invalid) else 'invalid'
        # 오류는 원시 record를 포함하지 않으며, 검증된 메타데이터만 보존함.
        out=locals().get('base',{'id':d.get('id') if isinstance(d,dict) and isinstance(d.get('id'),str) and len(d['id'])<=512 else None,'classification':classification})
        out.update(status=status,reasons=[reason],vp_input_eligible=False,vp_exclusion_reasons=[reason])
        return out

def qualify(manifest):
    fields(manifest,'version synthetic datasets')
    require(manifest['version']=='a152-v1' and type(manifest['synthetic']) is bool and isinstance(manifest['datasets'],list) and len(manifest['datasets'])<=100,'invalid_envelope')
    datasets=manifest['datasets']; ids=set(); count=0
    for d in datasets:
        if isinstance(d,dict):
            ident=d.get('id')
            if isinstance(ident,str): require(ident not in ids,'duplicate_dataset_id'); ids.add(ident)
            rows=d.get('records')
            if isinstance(rows,list): count+=len(rows)
    require(count<=100000,'record_limit')
    return {'version':'a152-v1','synthetic':manifest['synthetic'],'provenance':'caller-supplied, unverified','authenticity_verified':False,'coverage_verified':False,'profitability':'inconclusive','production_authorized':False,'non_profile_research_blocked':False,'datasets':[dataset(d,manifest['synthetic']) for d in datasets]}

def unique(pairs):
    obj={}
    for k,v in pairs:
        if k in obj: raise ValueError('duplicate_json_key')
        obj[k]=v
    return obj

def write_once(path,data):
    path=Path(path).absolute(); directory=os.open('/',os.O_RDONLY|os.O_DIRECTORY)
    try:
        for part in path.parts[1:-1]:
            nxt=os.open(part,os.O_RDONLY|os.O_DIRECTORY|os.O_NOFOLLOW,dir_fd=directory); os.close(directory); directory=nxt
        fd=os.open(path.name,os.O_WRONLY|os.O_CREAT|os.O_EXCL|os.O_NOFOLLOW,0o600,dir_fd=directory)
        with os.fdopen(fd,'wb') as f: f.write(data)
    finally: os.close(directory)

def main():
    parser=argparse.ArgumentParser(description=__doc__); parser.add_argument('input'); parser.add_argument('--output'); args=parser.parse_args()
    try:
        with open(args.input,'rb') as f: data=f.read(16*1024*1024+1)
        if len(data)>16*1024*1024: raise ValueError('input_limit')
        obj=json.loads(data,object_pairs_hook=unique,parse_constant=lambda _: (_ for _ in ()).throw(ValueError('invalid_constant')))
        out=qualify(obj); out['input_sha256']=hashlib.sha256(data).hexdigest()
        encoded=(json.dumps(out,sort_keys=True,ensure_ascii=False,allow_nan=False)+'\n').encode()
        if args.output: write_once(args.output,encoded)
        else: sys.stdout.buffer.write(encoded)
        return 0
    except (ValueError, OSError, TypeError, OverflowError, RecursionError):
        print('a152: invalid input or output',file=sys.stderr); return 1
if __name__=='__main__': raise SystemExit(main())

"""Offline, caller-supplied stop comparison. No trading authority or data collection."""
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

D = re.compile(r'^[0-9]+(?:\.[0-9]+)?$')
ARMS = ('fixed', 'structural', 'atr')

class ManifestError(ValueError):
    pass

class CandidateError(ValueError):
    def __init__(self, reason, status='excluded'):
        super().__init__(reason); self.status = status

def exact(x):
    if not isinstance(x, str) or len(x) > 64 or not D.fullmatch(x):
        raise ValueError('invalid plain decimal')
    return Fraction(x)

def integer(x):
    if isinstance(x, bool) or not isinstance(x, int) or x < 0 or x > 1000000:
        raise ValueError('invalid nonnegative integer')
    return x

def timestamp(s):
    if not isinstance(s, str): raise ValueError('timestamp required')
    t = dt.datetime.fromisoformat(s.replace('Z','+00:00'))
    if t.tzinfo is None or t.utcoffset() is None: raise ValueError('timestamp offset required')
    return t

def fields(x, keys):
    if not isinstance(x, dict) or set(x) != set(keys.split()):
        raise ManifestError('missing or unknown fields')

def nonempty(x):
    if not isinstance(x, str) or not x.strip() or x != x.strip(): raise ValueError('missing provenance')
    return x

def number(q):
    q = Fraction(q)
    d = q.denominator
    for n in (2, 5):
        while d % n == 0: d //= n
    if d != 1: return f'{q.numerator}/{q.denominator}'
    n, d = q.numerator, q.denominator
    sign = '-' if n < 0 else ''; n = abs(n)
    whole, rem = divmod(n, d)
    if not rem: return sign + str(whole)
    digits = ''
    while rem:
        rem *= 10; digit, rem = divmod(rem, d); digits += str(digit)
    return sign + str(whole) + '.' + digits

def floor_grid(x, tick): return (x // tick) * tick
def ceil_grid(x, tick): return -((-x) // tick) * tick

def atr14(bars):
    if len(bars) < 15: raise CandidateError('insufficient_atr')
    tr = [max(h-l, abs(h-bars[i-1][2]), abs(l-bars[i-1][2])) for i,(h,l,c) in enumerate(bars) if i]
    value = sum(tr[:14], Fraction()) / 14
    for t in tr[14:]: value = (value*13+t)/14
    return value

def geometry(entry, low, atr, tick, budget, max_qty, max_notional):
    target = ceil_grid(entry * Fraction(1021,1000), tick)
    stops = (floor_grid(entry * Fraction(993,1000), tick), floor_grid(low,tick)-tick, floor_grid(entry-Fraction(3,2)*atr,tick))
    result = {}
    for name, stop in zip(ARMS, stops):
        if not 0 < stop < entry < target: raise CandidateError('invalid_stop')
        span = entry-stop
        quantity = min(budget//span, max_qty, max_notional//entry)
        if quantity <= 0: raise CandidateError('zero_quantity')
        result[name] = {'stop':number(stop), 'quantity':int(quantity), 'target':number(target), 'span':number(span), 'risk':number(span*quantity), 'risk_usage':number(span*quantity/budget)}
    return result

def costs(entry, exit_price, quantity, cost):
    buy, sell = entry*quantity, exit_price*quantity
    quantum = exact(cost['quantum'])
    return (ceil_grid(buy*exact(cost['buy_rate'])+exact(cost['buy_fixed']),quantum),
            ceil_grid(sell*exact(cost['sell_rate'])+exact(cost['sell_fixed']),quantum),
            ceil_grid(sell*exact(cost['tax_rate']),quantum))

def pnl(entry, exit_price, quantity, cost):
    return (exit_price-entry)*quantity-sum(costs(entry,exit_price,quantity,cost),Fraction())

def provenance(x, signal):
    for k in ('source','version','digest'): nonempty(x[k])
    if timestamp(x['as_of']) > signal: raise CandidateError('future_snapshot')

def parse_bars(items, interval, session_open, session_close, tick, minimum, maximum):
    if not isinstance(items, list) or len(items) > (390 if interval == 1 else 78): raise ManifestError('bar count')
    result=[]; previous=None
    for b in items:
        fields(b, 'close_at received_at open high low close volume')
        closed, received = timestamp(b['close_at']), timestamp(b['received_at'])
        if received < closed: raise CandidateError('received_before_close')
        seconds=(closed-session_open).total_seconds()
        if closed <= session_open or closed > session_close or seconds % (interval*60): raise CandidateError('bar_outside_session')
        if previous is not None and closed <= previous: raise CandidateError('duplicate_or_reversed_bars')
        vals=[exact(b[k]) for k in ('open','high','low','close','volume')]
        o,h,l,c,v=vals
        if l <= 0 or h < l or not l <= o <= h or not l <= c <= h: raise CandidateError('invalid_ohlc')
        if any(p < minimum or p > maximum or p % tick for p in (o,h,l,c)): raise CandidateError('tick_range_or_grid','unsupported')
        result.append((closed,received,o,h,l,c,v)); previous=closed
    return result

def evaluate(candidate):
    fields(candidate, 'id market currency account_currency symbol date basis data_source data_version signal_confirmed_at signal_close_at session tick cost limits entry_slippage_ticks exit_slippage_ticks bars_1m bars_5m')
    for k in ('id','symbol','data_source','data_version'): nonempty(candidate[k])
    market=candidate['market']; curr={'KR':'KRW','US':'USD'}.get(market)
    if curr is None or candidate['currency'] != curr or candidate['account_currency'] != curr or candidate['basis'] != 'unadjusted': raise CandidateError('market_currency_or_basis','unsupported')
    signal=timestamp(candidate['signal_confirmed_at']); signal_close=timestamp(candidate['signal_close_at'])
    session=candidate['session']; tick_data=candidate['tick']; cost=candidate['cost']; limits=candidate['limits']
    fields(session,'source version digest as_of timezone kind open_at close_at')
    fields(tick_data,'source version digest as_of market symbol valid_from valid_until value min_price max_price')
    fields(cost,'source version digest as_of currency buy_rate sell_rate tax_rate buy_fixed sell_fixed quantum')
    fields(limits,'source version digest as_of currency risk_budget max_quantity max_notional')
    for x in (session,tick_data,cost,limits): provenance(x,signal)
    zone_name={'KR':'Asia/Seoul','US':'America/New_York'}[market]
    if session['timezone'] != zone_name: raise CandidateError('timezone_mismatch')
    zone=ZoneInfo(zone_name); op=timestamp(session['open_at']); cl=timestamp(session['close_at'])
    for t in (op,cl):
        local=t.astimezone(zone)
        if local.date().isoformat()!=candidate['date'] or t.utcoffset()!=local.utcoffset(): raise CandidateError('session_date_or_offset')
    local_op=op.astimezone(zone); local_cl=cl.astimezone(zone)
    if session['kind']=='holiday': raise CandidateError('holiday')
    expected_open=dt.time(9,0) if market=='KR' else dt.time(9,30)
    expected_close=dt.time(15,30) if market=='KR' else dt.time(16,0)
    length=(cl-op).total_seconds()
    if local_op.time()!=expected_open or length<=0 or length>23400 or length%300 or session['kind'] not in ('regular','early_close') or (session['kind']=='regular' and local_cl.time()!=expected_close) or (session['kind']=='early_close' and local_cl.time()>=expected_close): raise CandidateError('invalid_session')
    if tick_data['market'] != market or tick_data['symbol'] != candidate['symbol'] or not timestamp(tick_data['valid_from']) <= op <= signal < timestamp(tick_data['valid_until']) or timestamp(tick_data['valid_until']) < cl: raise CandidateError('tick_scope_or_validity','unsupported')
    tick=exact(tick_data['value']); lo=exact(tick_data['min_price']); hi=exact(tick_data['max_price'])
    if tick <= 0 or lo <= 0 or hi < lo: raise CandidateError('invalid_tick')
    if cost['currency']!=curr or limits['currency']!=curr: raise CandidateError('currency_mismatch','unsupported')
    if any(exact(cost[k])>1 for k in ('buy_rate','sell_rate','tax_rate')) or exact(cost['quantum'])<=0: raise CandidateError('invalid_cost')
    for k in ('buy_fixed','sell_fixed'): exact(cost[k])
    budget=exact(limits['risk_budget']); notional=exact(limits['max_notional']); max_qty=exact(limits['max_quantity'])
    if max_qty.denominator != 1: raise CandidateError('fractional_quantity_cap')
    entry_slip=integer(candidate['entry_slippage_ticks']); exit_slip=integer(candidate['exit_slippage_ticks'])
    one=parse_bars(candidate['bars_1m'],1,op,cl,tick,lo,hi); five=parse_bars(candidate['bars_5m'],5,op,cl,tick,lo,hi)
    by1={b[0]:b for b in one}; by5={b[0]:b for b in five}
    # 모든 제공 5분봉은 동일 구간의 1분봉과 일치해야 함.
    unverified=[]
    for b in five:
        group=[by1.get(b[0]-dt.timedelta(minutes=j)) for j in range(4,-1,-1)]
        if any(x is None for x in group):
            if b[0] <= signal: raise CandidateError('missing_1m_aggregate')
            unverified.append({'close_at':b[0].isoformat(),'reason':'unverified_missing_constituents'})
            continue
        if (b[2],b[3],b[4],b[5],b[6]) != (group[0][2],max(x[3] for x in group),min(x[4] for x in group),group[-1][5],sum((x[6] for x in group),Fraction())): raise CandidateError('bar_aggregation_mismatch')
    if signal_close > signal or signal_close not in by5: raise CandidateError('missing_signal_bar')
    if signal < op or signal >= cl: raise CandidateError('signal_outside_session')
    n=int((signal_close-op).total_seconds()//300)
    prefix=[]
    for i in range(1,n+1):
        end=op+dt.timedelta(minutes=5*i); b=by5.get(end)
        if not b or b[1] > signal: raise CandidateError('missing_or_late_warmup')
        group=[by1.get(end-dt.timedelta(minutes=j)) for j in range(4,-1,-1)]
        if any(x is None or x[1]>signal for x in group): raise CandidateError('missing_or_late_warmup')
        if (b[2],b[3],b[4],b[5],b[6]) != (group[0][2],max(x[3] for x in group),min(x[4] for x in group),group[-1][5],sum((x[6] for x in group),Fraction())): raise CandidateError('bar_aggregation_mismatch')
        prefix.append(b)
    atr=atr14([(b[3],b[4],b[5]) for b in prefix]); low=prefix[-1][4]
    minute=int((signal-op).total_seconds()//60)+1
    entry_open=op+dt.timedelta(minutes=minute); end=entry_open+dt.timedelta(minutes=1)
    bar=by1.get(end)
    if entry_open >= cl or bar is None: raise CandidateError('missing_next_bar','unfilled')
    entry=ceil_grid(bar[2]+entry_slip*tick,tick)
    arms=geometry(entry,low,atr,tick,budget,int(max_qty),notional)
    target=exact(arms['fixed']['target'])
    if not lo<=entry<=hi or not lo<=target<=hi or any(not lo<=exact(a['stop'])<=hi for a in arms.values()): raise CandidateError('computed_price_outside_tick_range','unsupported')
    for name in ARMS:
        arm=arms[name]; stop=exact(arm['stop']); qty=arm['quantity']; last=None; outcomes={}
        for order in ('stop_first','target_first'):
            chosen=None; ambiguous=False; mark=None
            for t in range(minute+1,int(length//60)+1):
                current=by1.get(op+dt.timedelta(minutes=t))
                if current is None: break
                closed,received,o,h,l,c,v=current; mark=(closed,c)
                if o<=stop: fill=floor_grid(o-exit_slip*tick,tick); why='stop_gap'
                elif o>=target: fill=target; why='target_gap'
                else:
                    st=l<=stop; tg=h>=target
                    if st and tg: ambiguous=True
                    if st and (not tg or order=='stop_first'): fill=floor_grid(stop-exit_slip*tick,tick); why='stop_touch'
                    elif tg: fill=target; why='target_touch'
                    else: continue
                if fill<=0 or not lo<=fill<=hi:
                    chosen={'status':'invalid','reason':'invalid_exit_price','at':closed.isoformat()}; break
                fees=costs(entry,fill,qty,cost); net=pnl(entry,fill,qty,cost); chosen={'buy_fee':number(fees[0]),'sell_fee':number(fees[1]),'sell_tax':number(fees[2]),'status':'closed','reason':why,'exit':number(fill),'at':closed.isoformat(),'pnl':number(net),'simulated_r':number(net/(qty*(entry-stop)))}; break
            if chosen is None:
                chosen={'status':'censored','reason':'missing_path_or_session_end','mark_at':mark[0].isoformat() if mark else None,'mark_price':number(mark[1]) if mark else None,'gross_mark_pnl':number((mark[1]-entry)*qty) if mark else None}
            chosen['ambiguous']=ambiguous; outcomes[order]=chosen
        arm['base']=outcomes['stop_first']; arm['sensitivity']=outcomes['target_first']
    return {'id':candidate['id'],'market':market,'status':'eligible','symbol':candidate['symbol'],'date':candidate['date'],'currency':curr,'entry':number(entry),'observed_open':number(bar[2]),'entry_slippage':number(entry-bar[2]),'entry_open_at':entry_open.isoformat(),'atr':number(atr),'unverified_future_5m':len(unverified),'unverified_5m':unverified,'model':{'session':session,'tick':tick_data,'cost':cost,'limits':limits,'signal_confirmed_at':candidate['signal_confirmed_at'],'signal_close_at':candidate['signal_close_at'],'data_source':candidate['data_source'],'data_version':candidate['data_version'],'basis':candidate['basis'],'entry_slippage_ticks':entry_slip,'exit_slippage_ticks':exit_slip},'arms':arms}

def compare(manifest):
    fields(manifest,'version synthetic mode candidates')
    if manifest['version']!='a151-v1' or not isinstance(manifest['synthetic'],bool): raise ValueError('invalid manifest version/synthetic')
    if manifest['mode']!='initial_stops': raise ValueError('unsupported mode')
    candidates=manifest['candidates']
    if not isinstance(candidates,list) or len(candidates)>1000: raise ValueError('candidate limit')
    ids=set(); results=[]
    for c in candidates:
        if not isinstance(c,dict): raise ValueError('candidate structure')
        ident=nonempty(c.get('id'))
        if ident in ids: raise ValueError('duplicate candidate id')
        ids.add(ident)
        try: results.append(evaluate(c))
        except ManifestError: raise
        except CandidateError as e: results.append({'id':ident,'market':c.get('market'), 'status':e.status,'reason':str(e)})
        except (ValueError,KeyError,TypeError,OverflowError) as e: results.append({'id':ident,'market':c.get('market'),'status':'invalid','reason':str(e)})
    stats={}
    for market in ('KR','US'):
        rows=[c for c in results if c['market']==market]; eligible=[c for c in rows if c['status']=='eligible']; complete=[c for c in eligible if all(c['arms'][x]['base']['status']=='closed' for x in ARMS)]
        pairs={}
        for left,right in (('structural','fixed'),('atr','fixed')):
            vals=[Fraction(c['arms'][left]['base']['simulated_r'])-Fraction(c['arms'][right]['base']['simulated_r']) for c in complete]
            pairs[left+'_minus_'+right]=number(sum(vals,Fraction())/len(vals)) if vals else None
        stats[market]={'total':len(rows),'eligible':len(eligible),'complete':len(complete),'paired_mean_simulated_r':pairs,'excluded':{reason:sum(1 for c in rows if c.get('reason')==reason) for reason in sorted({c['reason'] for c in rows if 'reason' in c})},'arms':{x:{status:sum(1 for c in eligible if c['arms'][x]['base']['status']==status) for status in ('closed','censored','invalid')} | {'unfilled':sum(1 for c in rows if c['status']=='unfilled'),'ambiguous':sum(1 for c in eligible if c['arms'][x]['base']['ambiguous'])} for x in ARMS}}
    return {'version':'a151-v1','synthetic':manifest['synthetic'],'provenance':'caller-supplied, unverified','profitability':'inconclusive','candidates':results,'markets':stats}

def unique_pairs(pairs):
    result={}
    for k,v in pairs:
        if k in result: raise ValueError('duplicate JSON key')
        result[k]=v
    return result

def write_once(path,data):
    absolute=Path(path).absolute(); directory=os.open('/',os.O_RDONLY|os.O_DIRECTORY)
    try:
        for part in absolute.parts[1:-1]:
            nxt=os.open(part,os.O_RDONLY|os.O_DIRECTORY|os.O_NOFOLLOW,dir_fd=directory); os.close(directory); directory=nxt
        fd=os.open(absolute.name,os.O_WRONLY|os.O_CREAT|os.O_EXCL|os.O_NOFOLLOW,0o600,dir_fd=directory)
        with os.fdopen(fd,'wb') as f: f.write(data)
    finally: os.close(directory)

def main():
    parser=argparse.ArgumentParser(description=__doc__); parser.add_argument('input'); parser.add_argument('--output'); args=parser.parse_args()
    try:
        with open(args.input,'rb') as f: data=f.read(16*1024*1024+1)
        if len(data)>16*1024*1024: raise ValueError('input exceeds 16 MiB')
        manifest=json.loads(data,object_pairs_hook=unique_pairs,parse_constant=lambda x: (_ for _ in ()).throw(ValueError('invalid constant')))
        out=compare(manifest); out['input_sha256']=hashlib.sha256(data).hexdigest()
        encoded=(json.dumps(out,sort_keys=True,ensure_ascii=False)+'\n').encode()
        if args.output: write_once(args.output,encoded)
        else: sys.stdout.buffer.write(encoded)
    except (ValueError,OSError,KeyError,TypeError,OverflowError) as e:
        print('a151: '+str(e),file=sys.stderr); return 1
    return 0
if __name__=='__main__': raise SystemExit(main())

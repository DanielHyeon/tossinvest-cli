"""Offline primitives for prospective MFI evaluation; no collection or network access."""
import datetime as dt
import json
import math
import os
import stat
import importlib.util
from pathlib import Path


_reader_spec = importlib.util.spec_from_file_location('a145_reader', Path(__file__).resolve().parents[1] / 'a145-jev-offline-analysis' / 'analysis.py')
_reader = importlib.util.module_from_spec(_reader_spec)
_reader_spec.loader.exec_module(_reader)
timestamp = _reader.timestamp


def mfi(bars, observed, max_age):
    try:
        observed = timestamp(observed)
        if not isinstance(max_age, (int, float)) or isinstance(max_age, bool) or not math.isfinite(max_age) or max_age < 0:
            raise ValueError('invalid age')
    except (ValueError, TypeError, AttributeError):
        return {'reason': 'invalid_observation'}
    if len(bars) < 18:
        return {'reason': 'insufficient'}
    items = bars[-18:]
    times, flows = [], []
    try:
        for bar in items:
            moment = timestamp(bar['close_at'])
            values = [bar[x] for x in ('open', 'high', 'low', 'close', 'volume')]
            if any(isinstance(v, bool) or not isinstance(v, (int, float)) or not math.isfinite(v) for v in values):
                raise ValueError('invalid number')
            opening, high, low, close, volume = values
            if low <= 0 or high < low or not low <= opening <= high or not low <= close <= high or volume < 0:
                raise ValueError('invalid OHLCV')
            typical = (high + low + close) / 3
            if not math.isfinite(typical) or not math.isfinite(typical * volume):
                raise ValueError('overflow')
            times.append(moment)
            flows.append((typical, volume))
    except (KeyError, ValueError, TypeError, AttributeError, OverflowError):
        return {'reason': 'invalid_bar'}
    if any(t > observed for t in times):
        return {'reason': 'future'}
    if any(b - a != dt.timedelta(minutes=1) for a, b in zip(times, times[1:])):
        return {'reason': 'nonconsecutive'}
    if (observed - times[-1]).total_seconds() > max_age:
        return {'reason': 'stale'}

    def value(end):
        positive = negative = 0.0
        for i in range(end - 13, end + 1):
            price, volume = flows[i]
            money = price * volume
            if price > flows[i-1][0]:
                positive += money
            elif price < flows[i-1][0]:
                negative += money
        if positive == negative == 0:
            return None
        total = positive + negative
        if not math.isfinite(total) or not math.isfinite(positive) or not math.isfinite(negative):
            return None
        return 100 * (positive / total)

    current, previous = value(17), value(14)
    if current is None or previous is None:
        return {'reason': 'undefined_flow'}
    delta = current - previous
    return {'mfi': current, 'delta': delta, 'direction': delta_direction(delta)}


def delta_direction(delta):
    return 'unchanged' if abs(delta) <= 1e-9 else 'rising' if delta > 0 else 'falling'


def safe_output(source, output):
    source, output = Path(source), Path(output)
    for path in (source, output):
        absolute = path.absolute()
        for parent in (absolute, *absolute.parents):
            if parent.is_symlink():
                raise ValueError('symlink path')
    source, output = source.resolve(), output.resolve()
    if source == output or source in output.parents or output in source.parents:
        raise ValueError('overlapping input/output')


def write_once(path, obj):
    path = Path(path)
    # O_EXCL makes replacement impossible even if a symlink appears at the leaf.
    data = (json.dumps(obj, sort_keys=True, allow_nan=False) + '\n').encode()
    absolute = path.absolute()
    directory = os.open('/', os.O_RDONLY | os.O_DIRECTORY)
    try:
        for part in absolute.parts[1:-1]:
            next_directory = os.open(part, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW, dir_fd=directory)
            os.close(directory)
            directory = next_directory
        fd = os.open(absolute.name, os.O_WRONLY | os.O_CREAT | os.O_EXCL | os.O_NOFOLLOW, 0o600, dir_fd=directory)
        with os.fdopen(fd, 'wb') as stream:
            stream.write(data)
    finally:
        os.close(directory)


def validate_plan(plan):
    for key in ('max_bar_age_seconds', 'max_label_delay_seconds'):
        value = plan.get(key)
        if isinstance(value, bool) or not isinstance(value, (int, float)) or not math.isfinite(value) or value < 0:
            raise ValueError(f'{key} required')
    from zoneinfo import ZoneInfo
    sessions = plan.get('session_plan')
    if not isinstance(sessions, dict) or set(sessions) != {'KR', 'US'}:
        raise ValueError('explicit KR and US sessions required')
    for market, windows in sessions.items():
        if not isinstance(windows, list) or not windows:
            raise ValueError('session windows required')
        seen = set()
        for window in windows:
            day = window['date']
            start, end = timestamp(window['open_at']), timestamp(window['close_at'])
            cadence, skew = window['cadence_seconds'], window['max_cycle_skew_seconds']
            zone = ZoneInfo('Asia/Seoul' if market == 'KR' else 'America/New_York')
            if day in seen or start.astimezone(zone).date().isoformat() != day or start >= end:
                raise ValueError('invalid session window')
            if any(isinstance(x, bool) or not isinstance(x, int) for x in (cadence, skew)) or cadence <= 0 or skew < 0 or skew >= cadence / 2:
                raise ValueError('invalid cycle cadence')
            seen.add(day)
    required = ('completeness_rules', 'candidate_rule', 'cycle_schedule',
                'interruption_rules',
                'replacement_rules', 'interim_reporting_rules', 'stop_rules',
                'analysis_version')
    for key in required:
        if not isinstance(plan.get(key), str) or not plan[key].strip():
            raise ValueError(f'{key} required')
    thresholds = plan.get('coverage_thresholds')
    if not isinstance(thresholds, dict):
        raise ValueError('coverage thresholds required')
    for field in ('min_cycle_fraction', 'min_mfi_fraction', 'min_label_fraction'):
        value = thresholds.get(field)
        if isinstance(value, bool) or not isinstance(value, (int,float)) or not math.isfinite(value) or not 0 <= value <= 1:
            raise ValueError('invalid coverage threshold')
    sources = plan.get('sources')
    if not isinstance(sources, dict) or set(sources) != {'KR','US'}:
        raise ValueError('market-specific sources required')
    for source in sources.values():
        if not isinstance(source, dict) or not source.get('name') or not source.get('timezone'):
            raise ValueError('source name and timezone required')
        ZoneInfo(source['timezone'])
    if plan.get('horizon_minutes') != 60:
        raise ValueError('60-minute horizon required')


def pair_candidates(judgments):
    """Keep candidate denominator even when question pair/selection is ambiguous."""
    import collections
    groups = collections.defaultdict(list)
    digests = collections.defaultdict(set)
    for row in judgments:
        identity = tuple(row.get(k) for k in ('state_rev','market','at','symbol'))
        digests[identity].add(row.get('state_digest'))
    for index, row in enumerate(judgments):
        key = (row.get('state_rev'), row.get('market'), row.get('at'), row.get('symbol'))
        # Without identity, count each unjoinable row, never merge distinct candidates.
        if any(not isinstance(part, str) or not part for part in key):
            key = (*key, index)
        groups[key].append(row)
    result = []
    for key, rows in groups.items():
        by_question = collections.defaultdict(list)
        for row in rows:
            by_question[row.get('question_id')].append(row)
        valid_pair = len(digests[key[:4]]) == 1 and len(rows) == 2 and all(len(by_question[q]) == 1 for q in ('j1_up_within_horizon', 'j2_hold_off_red_flag'))
        selection = 'unknown'
        if valid_pair:
            one, two = by_question['j1_up_within_horizon'][0], by_question['j2_hold_off_red_flag'][0]
            screen = one.get('screen')
            valid_pair = all(one.get(field) == two.get(field) for field in
                             ('screen', 'selection', 'model_answered', 'model_version',
                              'question_rev', 'label_due_at', 'horizon_minutes', 'base_price'))
            if valid_pair and isinstance(screen, dict) and isinstance(screen.get('finalist'), bool):
                selection = 'selected' if screen['finalist'] else 'nonselected'
        result.append({'key': key, 'selection': selection, 'pair_valid': valid_pair, 'rows': rows})
    return result


def label_return(judgment, label, max_delay):
    if label is None:
        return None, 'missing_label'
    try:
        observed = timestamp(judgment['at'])
        due = observed + dt.timedelta(minutes=60)
        if judgment['horizon_minutes'] != 60 or timestamp(judgment['label_due_at']) != due or timestamp(label['due_at']) != due:
            raise ValueError('bad horizon')
        delay = (timestamp(label['at']) - due).total_seconds()
        if delay < 0:
            raise ValueError('early')
        if delay > max_delay:
            return None, 'late_label'
        if label.get('label_gap'):
            if label.get('realized_price') is not None or label.get('realized_return') is not None:
                raise ValueError('gap with value')
            return None, 'label_gap'
        base, price, result = judgment['base_price'], label['realized_price'], label['realized_return']
        if any(isinstance(x, bool) or not isinstance(x, (float, int)) or not math.isfinite(x) for x in (base, price, result)) or base <= 0 or price <= 0:
            raise ValueError('invalid numeric label')
        if abs(result - (price-base)/base) > 5.1e-9:
            raise ValueError('return mismatch')
        return result, None
    except (KeyError, TypeError, ValueError, AttributeError, OverflowError):
        return None, 'invalid_label'


def summarize(records):
    """Group candidate outcomes; no account return or same-K strategy inference."""
    import collections
    import statistics
    groups = collections.defaultdict(list)
    cycles = collections.defaultdict(list)
    for row in records:
        key = (row['market'], row['session'], row['selection'])
        if key[2] not in ('selected', 'nonselected', 'unknown'):
            raise ValueError('invalid selection category')
        groups[key].append(row)
        cycles[(*key, row['cycle'])].append(row)
    output = []
    for key, rows in sorted(groups.items()):
        reasons = collections.Counter(row.get('mfi_reason') for row in rows if row.get('mfi_reason'))
        label_reasons = collections.Counter(row.get('label_reason') for row in rows if row.get('label_reason'))
        directions = {}
        for direction in ('rising', 'falling', 'unchanged'):
            subset = [row for row in rows if row.get('direction') == direction]
            values = [row['return'] for row in subset if row.get('return') is not None]
            directions[direction] = {'candidates': len(subset), 'valid_labels': len(values),
              'mean_return_pct': statistics.mean(values)*100 if values else None,
              'positive_fraction': sum(value > 0 for value in values)/len(values) if values else None}
        differences = []
        for cycle_key, subset in cycles.items():
            if cycle_key[:3] != key:
                continue
            rise = [r['return'] for r in subset if r.get('direction') == 'rising' and r.get('return') is not None]
            fall = [r['return'] for r in subset if r.get('direction') == 'falling' and r.get('return') is not None]
            if rise and fall:
                differences.append((statistics.mean(rise)-statistics.mean(fall))*100)
        output.append({'market':key[0], 'session':key[1], 'selection':key[2],
            'candidates':len(rows), 'mfi_valid':sum(bool(r.get('direction')) for r in rows),
            'mfi_excluded':dict(reasons), 'label_valid':sum(r.get('return') is not None for r in rows),
            'label_excluded':dict(label_reasons), 'directions':directions,
            'matched_cycles':len(differences),
            'equal_cycle_rising_minus_falling_pp':statistics.mean(differences) if differences else None})
    return output


def source_bars(state, evidence, observed, digest, source, window):
    """Accept only capture content matching the already-fixed r1 input and receipt time.

    File provenance and receipt metadata are supplied evidence, not cryptographic proof
    of the remote server's identity or the recording clock.
    """
    try:
        if evidence.get('source') != source['name'] or evidence.get('source_timezone') != source['timezone']:
            return None, 'source_mismatch'
        if (evidence['state_digest'] != digest or evidence['market'] != state['market'] or
            evidence['symbol'] != state['symbol'] or timestamp(evidence['observed_at']) != timestamp(observed)):
            return None, 'source_mismatch'
        received = timestamp(evidence['received_at'])
        if received > timestamp(observed) or not evidence['source']:
            return None, 'unproven_asof'
        source = evidence['bars']
        original = state['intraday']['bars_1m']
        if len(source) < 18 or len(original) < 18:
            return None, 'insufficient'
        source, original = source[-18:], original[-18:]
        for bar, previous in zip(source, original):
            close = timestamp(bar['close_at'])
            from zoneinfo import ZoneInfo
            zone = ZoneInfo(evidence['source_timezone'])
            if close.utcoffset() != close.astimezone(zone).utcoffset():
                return None, 'ambiguous_timezone'
            if close > received:
                return None, 'unproven_asof'
            if close <= timestamp(window['open_at']) or close > timestamp(window['close_at']):
                return None, 'outside_session'
            if close.strftime('%H:%M') != previous['t'] or any(bar[a] != previous[b] for a,b in (('open','o'),('high','h'),('low','l'),('close','c'),('volume','v'))):
                return None, 'source_mismatch'
        return source, None
    except (KeyError, TypeError, ValueError, AttributeError):
        return None, 'unproven_asof'


def read_evidence(path):
    path = Path(path).absolute()
    directory = os.open('/', os.O_RDONLY | os.O_DIRECTORY)
    try:
        for part in path.parts[1:-1]:
            nxt = os.open(part, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW, dir_fd=directory)
            os.close(directory)
            directory = nxt
        fd = os.open(path.name, os.O_RDONLY | os.O_NOFOLLOW, dir_fd=directory)
        with os.fdopen(fd, 'rb') as stream:
            return stream.read()
    finally:
        os.close(directory)


def cycle_slot(observed, window):
    start, end = timestamp(window['open_at']), timestamp(window['close_at'])
    elapsed = (observed-start).total_seconds()
    cadence = window['cadence_seconds']
    slot = round(elapsed/cadence)
    if observed < start or observed >= end or abs(elapsed-slot*cadence) > window['max_cycle_skew_seconds']:
        raise ValueError('observation outside preregistered window or cadence')
    moment = start + dt.timedelta(seconds=slot*cadence)
    if moment >= end:
        raise ValueError('cycle slot falls outside preregistered window')
    return moment.isoformat()


def deduplicate_records(records):
    import collections
    by_identity = collections.defaultdict(list)
    for record in records:
        by_identity[(record['market'],record['session'],record['cycle'],record['symbol'])].append(record)
    records = []
    conflicts = 0
    for key, items in by_identity.items():
        if len(items) > 1:
            conflicts += 1
            records.append({'market':key[0], 'session':key[1], 'cycle':key[2], 'symbol':key[3],
                'selection':'unknown', 'mfi_reason':'duplicate_cycle_identity', 'label_reason':'duplicate_cycle_identity'})
        else:
            records.append(items[0])
    return records, conflicts


def evaluate(snapshot, plan, sidecar):
    import collections
    import hashlib
    from zoneinfo import ZoneInfo
    validate_plan(plan)
    registered = timestamp(plan['registered_at'])
    snapshot = Path(snapshot)
    raw = read_evidence(snapshot / 'ledger.jsonl')
    judgments, labels, identifiers = [], {}, set()
    for line in raw.splitlines():
        row = json.loads(line)
        if row.get('kind') == 'judgment':
            if row['id'] in identifiers:
                raise ValueError('duplicate judgment ID')
            identifiers.add(row['id'])
            if timestamp(row['at']) <= registered or timestamp(row['at']) < timestamp(plan['first_observation_not_before']):
                raise ValueError('observation predates registration')
            judgments.append(row)
        elif row.get('kind') == 'label':
            if row['judgment_id'] in labels:
                raise ValueError('duplicate label ID')
            labels[row['judgment_id']] = row
    if set(labels) - identifiers:
        raise ValueError('orphan label')
    if not isinstance(sidecar, dict) or not isinstance(sidecar.get('captures'), dict) or not isinstance(sidecar.get('sessions'), dict):
        raise ValueError('sidecar captures and sessions maps required')
    records = []
    for candidate in pair_candidates(judgments):
        rows = candidate['rows']
        one = next((x for x in rows if x.get('question_id') == _reader.J1), None)
        first = one or rows[0]
        market = first.get('market')
        if market not in ('KR','US'):
            raise ValueError('invalid market')
        observed = timestamp(first['at'])
        day = observed.astimezone(ZoneInfo('Asia/Seoul' if market == 'KR' else 'America/New_York')).date().isoformat()
        windows = [w for w in plan['session_plan'][market] if w['date'] == day]
        if not windows:
            raise ValueError('observation belongs to an unregistered session')
        window = windows[0]
        cycle = cycle_slot(observed, window)
        record = {'market':market, 'session':day, 'cycle':cycle, 'symbol':first.get('symbol'), 'selection':candidate['selection'],
                  'mfi_reason':'unproven_asof', 'label_reason':'invalid_pair' if not candidate['pair_valid'] else 'missing_label'}
        if not candidate['pair_valid']:
            record['mfi_reason'] = 'invalid_pair'
            records.append(record)
            continue
        digest = first.get('state_digest', '').removeprefix('sha256:')
        if len(digest) != 64 or any(ch not in '0123456789abcdef' for ch in digest):
            record.update(selection='unknown', mfi_reason='invalid_identity', label_reason='invalid_identity')
            records.append(record)
            continue
        try:
            state_raw = read_evidence(snapshot / 'states' / (digest + '.json'))
        except FileNotFoundError:
            record.update(selection='unknown', mfi_reason='missing_state', label_reason='missing_state')
            records.append(record)
            continue
        if not state_raw.endswith(b'\n') or hashlib.sha256(state_raw[:-1]).hexdigest() != digest:
            raise ValueError('state digest mismatch')
        state = json.loads(state_raw)
        identity_matches = (state.get('market') == market and state.get('symbol') == first.get('symbol') and
                            timestamp(state['observed_at']) == observed and state.get('base_price') == first.get('base_price'))
        if not identity_matches:
            record.update(selection='unknown', mfi_reason='invalid_identity', label_reason='invalid_identity')
            records.append(record)
            continue
        capture_digest = sidecar['captures'].get(digest)
        if isinstance(capture_digest, str) and len(capture_digest) == 64 and all(ch in '0123456789abcdef' for ch in capture_digest):
            try:
                data = read_evidence(snapshot / 'captures' / (capture_digest + '.json'))
            except FileNotFoundError:
                data = None
            if data is None:
                bars, reason = None, 'missing_capture'
            elif hashlib.sha256(data).hexdigest() != capture_digest:
                bars, reason = None, 'capture_digest_mismatch'
            else:
                try:
                    bars, reason = source_bars(state, json.loads(data), first['at'], digest, plan['sources'][market], window)
                except (ValueError, TypeError):
                    bars, reason = None, 'invalid_capture'
            if reason:
                record['mfi_reason'] = reason
            else:
                measure = mfi(bars, first['at'], plan['max_bar_age_seconds'])
                record['mfi_reason'] = measure.get('reason')
                record['direction'] = measure.get('direction')
        if candidate['pair_valid'] and one:
            two = next(x for x in rows if x.get('question_id') == _reader.J2)
            left, right = labels.get(one['id']), labels.get(two['id'])
            if left and right and all(left.get(k) == right.get(k) for k in ('realized_price','realized_return','label_gap')):
                value, reason = label_return(one, left, plan['max_label_delay_seconds'])
                _, other_reason = label_return(two, right, plan['max_label_delay_seconds'])
                record['return'] = value if not other_reason else None
                record['label_reason'] = reason or other_reason
            elif left and right:
                record['label_reason'] = 'label_disagreement'
        records.append(record)
    records, conflicts = deduplicate_records(records)
    session_status = []
    allowed_sessions = {market + ':' + w['date'] for market, windows in plan['session_plan'].items() for w in windows}
    if set(sidecar['sessions']) - allowed_sessions:
        raise ValueError('unregistered session status')
    for market, windows in plan['session_plan'].items():
        for window in windows:
            day = window['date']
            item = sidecar['sessions'].get(market + ':' + day, {'status':'unknown'})
            if not isinstance(item, dict) or item.get('status') not in ('complete','partial','failed','unknown'):
                raise ValueError('invalid session status')
            if item['status'] in ('partial','failed') and not item.get('reason'):
                raise ValueError('partial/failed session requires a reason')
            subset = [r for r in records if r['market'] == market and r['session'] == day]
            cycles = {r['cycle'] for r in subset}
            seconds = (timestamp(window['close_at']) - timestamp(window['open_at'])).total_seconds()
            planned_cycles = math.ceil(seconds / window['cadence_seconds'])
            cycle_fraction = len(cycles) / planned_cycles
            mfi_fraction = sum(bool(r.get('direction')) for r in subset) / len(subset) if subset else 0
            label_fraction = sum(r.get('return') is not None for r in subset) / len(subset) if subset else 0
            minimums = plan['coverage_thresholds']
            meets = (cycle_fraction >= minimums['min_cycle_fraction'] and mfi_fraction >= minimums['min_mfi_fraction'] and label_fraction >= minimums['min_label_fraction'])
            status = 'eligible_for_review' if item['status'] == 'complete' and meets and subset and len(cycles) == planned_cycles else 'incomplete'
            session_status.append({'market':market, 'session':day, 'planned_cycles':planned_cycles,
                'observed_cycles':len(cycles), 'cycle_fraction':cycle_fraction,
                'mfi_fraction':mfi_fraction, 'label_fraction':label_fraction,
                'supplied_status':item['status'], 'status':status, 'reason':item.get('reason'),
                'status_source':'supplied status plus calculated coverage; not proof of complete capture'})
    return {'status':'inconclusive', 'sessions':session_status, 'groups':summarize(records),
            'warnings':['Observational only; repeated symbols and overlapping horizons are not independent.',
                        'Capture source and receipt time are supplied provenance, not proof of their authenticity.',
                        'Manifest registration time is supplied evidence, not independent proof that the file was fixed before observation.',
                        'No account return, causality or same-K strategy effect is established.'],
            'raw_judgment_rows':len(judgments), 'candidate_records':len(records), 'duplicate_cycle_identities':conflicts}


def read_sidecar(path, manifest_digest):
    captures, sessions = {}, {}
    lines = read_evidence(path).splitlines()
    if not lines or json.loads(lines[0]) != {'kind':'manifest', 'sha256':manifest_digest}:
        raise ValueError('sidecar manifest hash mismatch')
    for line in lines[1:]:
        event = json.loads(line)
        if event.get('kind') == 'capture':
            key, value = event['state_digest'], event['capture_digest']
            if key in captures:
                raise ValueError('duplicate sidecar identity')
            captures[key] = value
        elif event.get('kind') == 'session':
            key = event['market'] + ':' + event['date']
            if key in sessions:
                raise ValueError('duplicate session record')
            sessions[key] = {'status':event['status'], 'reason':event.get('reason')}
        else:
            raise ValueError('unknown sidecar event')
    return {'captures':captures, 'sessions':sessions}


def append_event(path, event, manifest_digest, protected=()):
    import fcntl
    path = Path(path).absolute()
    data = (json.dumps(event, sort_keys=True, allow_nan=False) + '\n').encode()
    directory = os.open('/', os.O_RDONLY | os.O_DIRECTORY)
    try:
        for part in path.parts[1:-1]:
            nxt = os.open(part, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW, dir_fd=directory)
            os.close(directory)
            directory = nxt
        fd = os.open(path.name, os.O_RDWR | os.O_APPEND | os.O_CREAT | os.O_NOFOLLOW, 0o600, dir_fd=directory)
        with os.fdopen(fd, 'r+b') as stream:
            metadata = os.fstat(stream.fileno())
            if not stat.S_ISREG(metadata.st_mode) or metadata.st_nlink != 1:
                raise ValueError('sidecar must be a regular file without hard links')
            for source in protected:
                info = os.stat(source, follow_symlinks=False)
                if (metadata.st_dev, metadata.st_ino) == (info.st_dev, info.st_ino):
                    raise ValueError('sidecar aliases protected input')
            fcntl.flock(stream, fcntl.LOCK_EX)
            existing = stream.read()
            if existing and not existing.endswith(b'\n'):
                raise ValueError('incomplete sidecar')
            entries = [json.loads(line) for line in existing.splitlines()]
            header = {'kind':'manifest', 'sha256':manifest_digest}
            if entries and entries[0] != header:
                raise ValueError('sidecar manifest hash mismatch')
            if event.get('kind') == 'capture':
                key = event.get('state_digest')
                if not isinstance(key, str) or len(key) != 64 or any(x not in '0123456789abcdef' for x in key):
                    raise ValueError('invalid state digest')
                if any(e.get('kind') == 'capture' and e.get('state_digest') == key for e in entries):
                    raise ValueError('duplicate sidecar identity')
            elif event.get('kind') == 'session':
                if event.get('market') not in ('KR','US') or not event.get('date') or event.get('status') not in ('complete','partial','failed','unknown'):
                    raise ValueError('invalid session event')
                if event['status'] in ('partial','failed') and not event.get('reason'):
                    raise ValueError('reason required')
                if any(e.get('kind') == 'session' and e.get('market') == event['market'] and e.get('date') == event['date'] for e in entries):
                    raise ValueError('duplicate session record')
            else:
                raise ValueError('invalid event')
            if not entries:
                stream.write((json.dumps(header, sort_keys=True) + '\n').encode())
            stream.write(data)
            stream.flush()
            os.fsync(stream.fileno())
    finally:
        os.close(directory)


def main():
    import argparse
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('snapshot', type=Path)
    parser.add_argument('manifest', type=Path)
    parser.add_argument('sidecar', type=Path)
    parser.add_argument('output', type=Path)
    import sys
    if len(sys.argv) > 1 and sys.argv[1] == 'record':
        record = argparse.ArgumentParser(description='Append one immutable sidecar record')
        for name in ('snapshot', 'manifest', 'event', 'sidecar'):
            record.add_argument(name, type=Path)
        args = record.parse_args(sys.argv[2:])
        for source in (args.snapshot, args.manifest, args.event, Path(__file__).resolve().parents[2]):
            safe_output(source, args.sidecar)
        event = json.loads(read_evidence(args.event))
        import hashlib
        manifest_raw = read_evidence(args.manifest)
        plan = json.loads(manifest_raw)
        validate_plan(plan)
        if event.get('kind') == 'capture':
            digest, capture = event.get('state_digest'), event.get('capture_digest')
            if any(not isinstance(v,str) or len(v) != 64 or any(c not in '0123456789abcdef' for c in v) for v in (digest,capture)):
                raise ValueError('invalid evidence digest')
            state_data = read_evidence(args.snapshot / 'states' / (digest + '.json'))
            capture_data = read_evidence(args.snapshot / 'captures' / (capture + '.json'))
            if not state_data.endswith(b'\n') or hashlib.sha256(state_data[:-1]).hexdigest() != digest or hashlib.sha256(capture_data).hexdigest() != capture:
                raise ValueError('evidence digest mismatch')
            state = json.loads(state_data)
            from zoneinfo import ZoneInfo
            market = state['market']
            observed = timestamp(state['observed_at'])
            if observed <= timestamp(plan['registered_at']) or observed < timestamp(plan['first_observation_not_before']):
                raise ValueError('observation predates registration')
            local_day = timestamp(state['observed_at']).astimezone(ZoneInfo('Asia/Seoul' if market == 'KR' else 'America/New_York')).date().isoformat()
            windows = [w for w in plan['session_plan'][market] if w['date'] == local_day]
            if not windows:
                raise ValueError('unregistered session')
            cycle_slot(timestamp(state['observed_at']), windows[0])
            bars, reason = source_bars(state, json.loads(capture_data), state['observed_at'], digest, plan['sources'][market], windows[0])
            if reason:
                raise ValueError('capture not eligible: ' + reason)
        elif event.get('kind') == 'session':
            if not any(w['date'] == event.get('date') for w in plan['session_plan'].get(event.get('market'), [])):
                raise ValueError('unregistered session')
        append_event(args.sidecar, event, hashlib.sha256(manifest_raw).hexdigest(),
                     (args.snapshot / 'ledger.jsonl', args.manifest, args.event))
        print('Sidecar event appended; collection is not started')
        return
    if len(sys.argv) > 1 and sys.argv[1] == 'register':
        registration = argparse.ArgumentParser(description='Create a new prospective manifest')
        registration.add_argument('draft', type=Path)
        registration.add_argument('output', type=Path)
        args = registration.parse_args(sys.argv[2:])
        for source in (args.draft, Path(__file__).resolve().parents[2]):
            safe_output(source, args.output)
        plan = json.loads(read_evidence(args.draft))
        validate_plan(plan)
        now = dt.datetime.now(dt.timezone.utc)
        if timestamp(plan['first_observation_not_before']) <= now:
            raise ValueError('first observation must be in the future')
        if 'registered_at' in plan:
            raise ValueError('registration time is assigned by this tool')
        plan['registered_at'] = now.isoformat()
        write_once(args.output, plan)
        print('Prospective manifest created; collection is not started')
        return
    args = parser.parse_args()
    for source in (args.snapshot, args.manifest, args.sidecar, Path(__file__).resolve().parents[2]):
        safe_output(source, args.output)
    import hashlib
    manifest_raw = read_evidence(args.manifest)
    plan = json.loads(manifest_raw)
    sidecar = read_sidecar(args.sidecar, hashlib.sha256(manifest_raw).hexdigest())
    result = evaluate(args.snapshot, plan, sidecar)
    write_once(args.output, result)
    print('Offline evaluation written; status: inconclusive')


if __name__ == '__main__':
    main()

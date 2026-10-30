#!/usr/bin/env python3
"""Read-only paired profile evaluation; no preregistered verdict is changed."""
import argparse
import collections
import hashlib
import importlib.util
import json
import os
import pathlib
import statistics

ROOT = pathlib.Path(__file__).resolve().parents[2]
SPEC = importlib.util.spec_from_file_location("prereg_eval", ROOT / "openspec/changes/a128-jev-shadow-judgment-probe/analysis/prereg_eval.py")
prereg = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(prereg)


def evaluate(ledger, states_dir, types=None):
    types = types or {}
    rows, labels, sources = [], {}, set()
    with open(ledger, encoding="utf-8") as f:
        for line in f:
            if not line.strip():
                continue
            r = json.loads(line)
            if r.get("cohort") and r.get("source_state_digest"):
                sources.add(r["source_state_digest"].split(":", 1)[-1])
            if r.get("kind") == "judgment" and r.get("question_id") == prereg.J2 and r.get("input_profile") in ("current", "initial"):
                rows.append(r)
            elif r.get("kind") == "label":
                labels[r["judgment_id"]] = r.get("realized_return")
    loaded_states, unreadable = prereg.load_states(states_dir)
    def valid_digest(value):
        raw_digest = value.removeprefix("sha256:") if isinstance(value, str) and value.startswith("sha256:") else (value if isinstance(value, str) and len(value) == 64 else "")
        if len(raw_digest) != 64 or any(c not in "0123456789abcdef" for c in raw_digest):
            return False
        try:
            data = (pathlib.Path(states_dir) / (raw_digest + ".json")).read_bytes()
        except (FileNotFoundError, OSError):
            return False
        if data.endswith(b"\n"):
            data = data[:-1]
        return hashlib.sha256(data).hexdigest() == raw_digest
    all_states = {k:v for k,v in loaded_states.items() if valid_digest(k)}
    source_states = {k: v for k, v in all_states.items() if k in sources}
    bars = prereg.build_bars(source_states)
    grouped = collections.defaultdict(lambda: collections.defaultdict(list))
    invalid_rows = 0
    for r in rows:
        required = ("cohort", "source_state_digest", "state_digest", "market", "symbol", "at", "question_id", "question_rev", "question_digest", "model_requested", "model_answered")
        if any(not r.get(k) for k in required) or not r.get("horizon_minutes", 0) > 0 or not r.get("base_price", 0) > 0 or not str(r.get("source_state_digest", "")).startswith("sha256:") or not str(r.get("state_digest", "")).startswith("sha256:") or (r.get("market") == "US" and r.get("horizon_minutes") != 60) or not valid_digest(r.get("source_state_digest")) or not valid_digest(r.get("state_digest")):
            invalid_rows += 1
            continue
        key = tuple(r.get(k) for k in ("cohort", "source_state_digest", "market", "symbol", "at", "question_id", "question_rev", "question_digest", "model_requested", "model_answered", "horizon_minutes", "base_price"))
        grouped[key][r["input_profile"]].append(r)
    stockdays = collections.defaultdict(list)
    rejected = 0
    for pair in grouped.values():
        if len(pair["current"]) != 1 or len(pair["initial"]) != 1:
            rejected += 1
            continue
        a, b = pair["current"][0], pair["initial"][0]
        j = {"at": prereg.parse_at(a["at"]), "market": a["market"], "symbol": a["symbol"], "digest": a["state_digest"].split(":", 1)[-1]}
        if prereg.in_window(j, all_states) is not True:
            continue
        stockdays[prereg.stockday(j)].append((j["at"], a, b))
    groups = collections.defaultdict(list)
    missing = collections.Counter()
    us_entry_time_difference_stockdays = 0
    for sd, candidates in sorted(stockdays.items()):
        candidates.sort(key=lambda x: x[0])
        market, date, symbol = sd
        kind = prereg.d2_type(symbol, types)
        if market == "US":
            chosen_times = []
            for index in (1, 2):
                flagged = [candidate for candidate in candidates if float(candidate[index]["p"]) >= prereg.P_STRONG]
                chosen_times.append((flagged[0] if flagged else candidates[0])[0])
            us_entry_time_difference_stockdays += chosen_times[0] != chosen_times[1]
        for profile, index in (("current", 1), ("initial", 2)):
            flagged = [candidate for candidate in candidates if float(candidate[index]["p"]) >= prereg.P_STRONG]
            chosen = flagged[0] if flagged else candidates[0]
            a, b = chosen[1], chosen[2]
            if market == "KR":
                ret, why = prereg.kr_trade(bars.get(sd, {}))
            else:
                ra, rb = labels.get(a["id"]), labels.get(b["id"])
                if ra is None or rb is None or ra != rb:
                    ret, why = None, "shared_label_missing_or_mismatch"
                else:
                    ret, why = float(ra), None
            if ret is None:
                missing[profile + ":" + why] += 1
                continue
            side = "signal" if flagged else "non-signal"
            groups[(market, profile, side, kind, date)].append(ret)
    result = {"us_selection_rule": "first flagged observation per profile, otherwise first paired observation; same-observation labels are shared", "us_entry_time_difference_stockdays": us_entry_time_difference_stockdays, "paired_stockdays": len(stockdays), "rejected_pair_groups": rejected, "invalid_rows": invalid_rows, "missing_source_states": len(sources - set(source_states)), "unreadable_states": unreadable, "price_missing": dict(missing), "groups": {}}
    summaries = collections.defaultdict(list)
    for (market, profile, side, kind, date), values in groups.items():
        summaries[(market, profile, side, kind)].append((date, statistics.fmean(values), len(values)))
    for (market, profile, side, kind), daily in sorted(summaries.items()):
        gross = statistics.fmean(v for _, v, _ in daily)
        result["groups"]["/".join((market, profile, side, kind))] = {"stockdays": sum(n for _, _, n in daily), "dates": len(daily), "gross_date_equal": gross, "net_date_equal": gross - prereg.COST_RT[market], "gross_stockday_equal": sum(v*n for _, v, n in daily) / sum(n for _, _, n in daily), "net_stockday_equal": sum(v*n for _, v, n in daily) / sum(n for _, _, n in daily) - prereg.COST_RT[market]}
    return result


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("ledger")
    p.add_argument("--states", required=True)
    p.add_argument("--types", help="optional security-types.json")
    args = p.parse_args()
    types = {}
    if args.types:
        with open(args.types, encoding="utf-8") as f:
            raw = json.load(f)
        types = raw.get("types", raw)
    print(json.dumps(evaluate(args.ledger, args.states, types), ensure_ascii=False, indent=2, sort_keys=True))

if __name__ == "__main__":
    main()

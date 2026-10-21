#!/usr/bin/env python3
"""a128 사전등록 가설 평가기 — J2 강한 경고(Jev) 와 공식 경고(F) 의 4 팔 + 맞대결.

정본 규칙: prereg-2026-10-22-j2-warning-hypotheses.md (같은 디렉터리).
이 스크립트는 그 문서의 규칙을 기계로 옮긴 것이며 규칙을 바꾸지 않음.

- 표준 라이브러리만 사용, 읽기 전용(원장·state 파일을 쓰거나 옮기지 않음).
- 주문·계좌·네트워크 접근 없음.
- 실행: python3 -I prereg_eval.py --window in-sample | out-of-sample
        python3 -I prereg_eval.py --self-test
        python3 -I prereg_eval.py --calibrate-x   (F 임계값 X 도출 과정 재현 — 감사용)
"""
from __future__ import annotations

import argparse
import datetime as dt
import json
import os
import statistics
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
LEDGER = os.path.join(HERE, "shadow-ledger", "ledger.jsonl")
STATES = os.path.join(HERE, "shadow-ledger", "states")

J2 = "j2_hold_off_red_flag"
P_STRONG = 0.6  # Jev 강한 경고 문턱(p >= 0.6)

# F(공식) 강한 경고 문턱 — 표본 내 빈도 맞춤으로 2026-10-22 에 고정한 값(사전등록 문서 4절).
# 이후 변경 금지. --calibrate-x 는 도출 재현용이며 평가에 쓰이지 않음.
X_FROZEN = {"KR": -2.69, "US": -7.15}

# 왕복 비용 — 엔진이 쓰는 공유 비용 모델 costs.DefaultModel() 의 값(사전등록 문서 3절).
# KR: 매수 수수료 0.10% + 매도 수수료 0.10% + 매도 거래세 0.30% = 0.50%
# US: 매수 0.50% + 매도 0.50% + 매도 규제 수수료 0.10% + 환전 0.50% x 2 다리 = 2.10%
COST_RT = {"KR": 0.0050, "US": 0.0210}

N_MIN = {"KR": 20, "US": 15}  # 통과 조건 (3) 최소 신호 거래 수

IN_SAMPLE_END = "2026-10-22"  # 이 날짜 미만 세션 = 표본 내
OOS_NOMINAL = ["2026-10-22", "2026-10-23", "2026-10-26", "2026-10-27", "2026-10-28"]
OOS_EXTENSION = ["2026-10-29", "2026-10-30"]  # 결측 세션 대체 상한 2
KST = dt.timezone(dt.timedelta(hours=9))


# ---------------------------------------------------------------- 적재(읽기 전용)

def parse_at(s: str) -> dt.datetime:
    """원장 시각(UTC, Z 접미) 을 aware datetime 으로 변환함."""
    return dt.datetime.fromisoformat(s.replace("Z", "+00:00"))


def load_ledger(path: str):
    """J2 판단 행과 라벨 행을 읽음.

    프로브가 지금도 덧붙이는 중이므로 마지막 줄이 잘려 있을 수 있음 — 마지막 줄의
    해석 실패만 건너뛰고 그 수를 셈. 중간 줄의 해석 실패는 원장 손상이므로 중단함.
    """
    with open(path, "rb") as fh:
        lines = fh.read().decode("utf-8").split("\n")
    j2, labels, skipped = [], {}, 0
    last = max((i for i, ln in enumerate(lines) if ln.strip()), default=-1)
    for i, ln in enumerate(lines):
        if not ln.strip():
            continue
        try:
            r = json.loads(ln)
        except json.JSONDecodeError:
            if i == last:
                skipped += 1
                continue
            raise
        if r.get("kind") == "judgment" and r.get("question_id") == J2:
            j2.append({
                "id": r["id"], "at": parse_at(r["at"]), "market": r["market"],
                "symbol": r["symbol"], "p": float(r["p"]),
                "base_price": r.get("base_price"),
                "digest": r["state_digest"].split(":", 1)[1],
            })
        elif r.get("kind") == "label":
            labels[r["judgment_id"]] = r.get("realized_return")
    return j2, labels, skipped


def load_states(dirpath: str):
    """state 파일 전부를 digest(hex) → 필요한 필드로 읽음. 쓰는 중인 파일은 건너뛰고 셈."""
    states, unreadable = {}, 0
    for name in sorted(os.listdir(dirpath)):
        if not name.endswith(".json"):
            continue
        try:
            with open(os.path.join(dirpath, name), "rb") as fh:
                d = json.loads(fh.read().decode("utf-8"))
        except (json.JSONDecodeError, UnicodeDecodeError):
            unreadable += 1
            continue
        intraday = d.get("intraday") or {}
        states[name[:-5]] = {
            "market": d.get("market"), "symbol": d.get("symbol"),
            "observed_at": parse_at(d["observed_at"]) if d.get("observed_at") else None,
            "minutes_since_open": d.get("minutes_since_open"),
            "change_pct": d.get("change_pct_vs_reference"),
            "bars": intraday.get("bars_1m") or [],
        }
    return states, unreadable


# ---------------------------------------------------------------- 규칙

def session_date(j) -> str:
    """세션 날짜 = 판단 시각의 UTC 날짜. KR 장(00:00~06:30Z)은 KST 날짜와 같고,
    US 장(13:30~20:00Z)은 개장일과 같음."""
    return j["at"].date().isoformat()


def stockday(j):
    return (j["market"], session_date(j), j["symbol"])


def in_window(j, states):
    """신호 창 판정. KR: 판단 시각 09:00~09:59 KST. US: state 의 minutes_since_open 30~60.
    US 에서 state 가 없으면 None(판정 불가 — 호출자가 셈)."""
    if j["market"] == "KR":
        return j["at"].astimezone(KST).hour == 9
    st = states.get(j["digest"])
    if st is None or st["minutes_since_open"] is None:
        return None
    return 30 <= st["minutes_since_open"] <= 60


def build_bars(states):
    """주식-일 (market, UTC 날짜, symbol) 별 1분봉 합집합.
    같은 라벨의 봉이 여러 state 에 있으면 가장 늦게 관측한 state 의 값을 씀
    (앞선 state 의 마지막 봉은 아직 형성 중일 수 있으므로)."""
    out = {}
    for st in states.values():
        if st["observed_at"] is None:
            continue
        key = (st["market"], st["observed_at"].date().isoformat(), st["symbol"])
        book = out.setdefault(key, {})
        for b in st["bars"]:
            prev = book.get(b["t"])
            if prev is None or st["observed_at"] > prev[0]:
                book[b["t"]] = (st["observed_at"], float(b["c"]))
    return out


def pick_close(book, target: str, lo: str):
    """라벨 target 봉의 종가, 없으면 [lo, target] 안의 가장 가까운 이전 봉 종가, 없으면 None.
    KR 봉 라벨은 KST "HH:MM" 이고 자정을 넘지 않으므로 문자열 비교로 충분함."""
    cands = [t for t in book if lo <= t <= target]
    return book[max(cands)][1] if cands else None


def kr_trade(book):
    """KR 거래: 10:00 봉 종가 진입 → 14:10 봉 종가 청산. (총수익률, 결측 사유)."""
    entry = pick_close(book, "10:00", "09:55")
    if entry is None:
        return None, "entry_missing"
    exit_ = pick_close(book, "14:10", "14:05")
    if exit_ is None:
        return None, "exit_missing"
    return exit_ / entry - 1.0, None


def jev_flag(j, states):
    return j["p"] >= P_STRONG


def make_f_flag(x):
    """F 경고: 판단 시점 state 의 change_pct_vs_reference <= X. state·값이 없으면 경고 아님."""
    def flag(j, states):
        st = states.get(j["digest"])
        if st is None or st["change_pct"] is None:
            return False
        return float(st["change_pct"]) <= x
    return flag


def run_arm(market, window_js, flag, bars, labels, states):
    """한 팔(신호 대 기준선)의 거래 목록을 만듦.

    window_js: 해당 시장·창 안의 J2 판단(시각순).
    KR: 창 안에서 한 번이라도 경고된 주식-일 = 신호, 나머지 주식-일 = 기준선. 거래는 주식-일당 하나.
    US: 경고된 첫 판단 = 신호(base_price 진입, 그 판단의 60분 라벨 청산).
        경고가 하나도 없는 주식-일의 첫 판단 = 기준선.
    반환: {"signal": {sd: ret}, "base": {sd: ret}, "missing": {사유: 수}}
    """
    groups = {}
    for j in window_js:
        groups.setdefault(stockday(j), []).append(j)
    res = {"signal": {}, "base": {}, "missing": {}}

    def miss(reason):
        res["missing"][reason] = res["missing"].get(reason, 0) + 1

    for sd, js in sorted(groups.items()):
        js.sort(key=lambda j: j["at"])
        flagged = [j for j in js if flag(j, states)]
        side = "signal" if flagged else "base"
        if market == "KR":
            ret, why = kr_trade(bars.get(sd, {}))
        else:
            pick = flagged[0] if flagged else js[0]
            if pick["id"] not in labels:
                ret, why = None, "label_pending"
            elif labels[pick["id"]] is None:
                ret, why = None, "label_null"
            else:
                ret, why = float(labels[pick["id"]]), None
        if ret is None:
            miss(f"{side}:{why}")
            continue
        res[side][sd] = ret
    return res


def mean(xs):
    return statistics.fmean(xs) if xs else None


def verdict(market, sig, base):
    """통과 규칙: (1) 신호 순평균 > 0, (2) 신호 순평균 > 기준선 순평균, (3) n >= N_MIN."""
    if len(sig) < N_MIN[market]:
        return "INSUFFICIENT"
    if not base:
        return "INSUFFICIENT"  # 기준선 0 이면 (2) 판정 불가
    ms, mb = mean(sig), mean(base)
    return "PASS" if ms > 0 and ms > mb else "FAIL"


# ---------------------------------------------------------------- 평가

def select_sessions(window, market, js_by_date):
    """평가 세션 목록과 상태 문구를 정함.
    표본 내: 2026-10-22 미만 전부.
    표본 외: 명목 5 세션 + 대체 2 세션 후보 중, 신호 창 안 J2 판단이 있는 날짜의 앞 5개.
    (휴장·프로브 실패로 빈 명목 세션은 자동으로 다음 후보로 대체됨. 미래 세션은 아직 빈 것.)"""
    if window == "in-sample":
        ds = sorted(d for d in js_by_date if d < IN_SAMPLE_END)
        return ds, f"표본 내 세션 {len(ds)}개"
    cands = OOS_NOMINAL + OOS_EXTENSION
    present = [d for d in cands if js_by_date.get(d)][:5]
    missing = [d for d in OOS_NOMINAL if d not in present]
    ext = [d for d in present if d in OOS_EXTENSION]
    status = "COMPLETE" if len(present) == 5 else "INCOMPLETE(판정 잠정 — 세션 미충족)"
    return present, (f"표본 외 세션 {len(present)}/5 {status}; 명목 결측={missing or '없음'}; "
                     f"대체 사용={ext or '없음'}")


def evaluate(j2, labels, states, window, out=print):
    """전체 평가. 출력 문자열을 out 으로 보내고 결과 dict 를 반환함(자가 시험용)."""
    bars = build_bars(states)
    results = {}
    for market in ("KR", "US"):
        no_state = 0
        by_date = {}
        for j in j2:
            if j["market"] != market:
                continue
            w = in_window(j, states)
            if w is None:
                no_state += 1
                continue
            if w:
                by_date.setdefault(session_date(j), []).append(j)
        sessions, status = select_sessions(window, market, by_date)
        window_js = [j for d in sessions for j in by_date[d]]
        arms = {
            "Jev": run_arm(market, window_js, jev_flag, bars, labels, states),
            "F": run_arm(market, window_js, make_f_flag(X_FROZEN[market]), bars, labels, states),
        }
        results[market] = {"sessions": sessions, "arms": arms}
        out(f"\n=== {market} — {status}; 창 판정 불가(state 없음) {no_state}건; "
            f"왕복 비용 {COST_RT[market]*100:.2f}% ===")
        for name, arm in arms.items():
            label = {"Jev": f"H-{market} (Jev J2 p>={P_STRONG})",
                     "F": f"F-{market} (change_pct <= {X_FROZEN[market]:.2f})"}[name]
            out(f"\n-- {label}")
            out(f"{'session':<11} {'n_sig':>5} {'sig_net%':>9} {'n_base':>6} {'base_net%':>9}")
            pos = 0
            for d in sessions:
                s = [r - COST_RT[market] for sd, r in arm["signal"].items() if sd[1] == d]
                b = [r - COST_RT[market] for sd, r in arm["base"].items() if sd[1] == d]
                pos += 1 if s and mean(s) > 0 else 0
                out(f"{d:<11} {len(s):>5} {fmt(mean(s)):>9} {len(b):>6} {fmt(mean(b)):>9}")
            sig = [r - COST_RT[market] for r in arm["signal"].values()]
            base = [r - COST_RT[market] for r in arm["base"].values()]
            gross = [r for r in arm["signal"].values()]
            up = sum(1 for r in gross if r > 0)
            v = verdict(market, sig, base)
            arm["verdict"], arm["sig_net"], arm["base_net"] = v, mean(sig), mean(base)
            out(f"{'TOTAL':<11} {len(sig):>5} {fmt(mean(sig)):>9} {len(base):>6} {fmt(mean(base)):>9}")
            out(f"신호 총(gross) 평균 {fmt(mean(gross))}%, 상승 비율 {up}/{len(gross)}; "
                f"신호 순평균>0 세션 {pos}/{len(sessions)}(보고용); 결측 {arm['missing'] or '없음'}")
            out(f"판정: {v} (n 최소 {N_MIN[market]})")
        results[market]["h2h"] = head_to_head(market, arms, out)
    out("\n주의: PASS 는 추가 그림자 시험을 할 근거일 뿐 수익성 결론이 아님. 어떤 거래도 하지 않음.")
    return results


def head_to_head(market, arms, out):
    """Jev 팔 대 F 팔 맞대결. 셀: 둘 다 / Jev 만 / F 만 경고된 주식-일.
    둘 다 셀에서 US 는 두 팔이 다른 판단을 고를 수 있어 팔별 평균을 따로 보임."""
    c = COST_RT[market]
    js, fs = arms["Jev"]["signal"], arms["F"]["signal"]
    both = sorted(set(js) & set(fs))
    jonly = [js[k] - c for k in js if k not in fs]
    fonly = [fs[k] - c for k in fs if k not in js]
    mj, mf = mean([r - c for r in js.values()]), mean([r - c for r in fs.values()])
    mjo, mfo = mean(jonly), mean(fonly)
    adds = (mj is not None and mf is not None and mjo is not None and mfo is not None
            and mj > mf and mjo > mfo)
    out(f"\n-- 맞대결 {market}: Jev 순평균 {fmt(mj)}% vs F 순평균 {fmt(mf)}%, 차이 "
        f"{fmt(None if mj is None or mf is None else mj - mf)}%p")
    out(f"   둘 다 n={len(both)} (Jev 거래 {fmt(mean([js[k]-c for k in both]))}% / "
        f"F 거래 {fmt(mean([fs[k]-c for k in both]))}%); Jev 만 n={len(jonly)} {fmt(mjo)}%; "
        f"F 만 n={len(fonly)} {fmt(mfo)}%")
    # 어느 팔이든 신호 거래가 0 이면 비교 대상 자체가 없으므로 판정 불가로 표시함
    # (한 팔은 있고 Jev 만·F 만 셀이 빈 경우는 사전등록 규칙대로 "formula is enough")
    v = ("판정 불가(팔 신호 0)" if mj is None or mf is None
         else "Jev adds value" if adds else "formula is enough")
    out(f"   맞대결 판정: {v}")
    return {"verdict": v, "both": len(both), "jev_only": len(jonly), "f_only": len(fonly)}


def fmt(x):
    return "-" if x is None else f"{x*100:+.2f}"


# ---------------------------------------------------------------- X 도출(감사용 재현)

def calibrate_x(j2, states, out=print):
    """표본 내 빈도 맞춤: 창 안 J2 판단 중 p>=0.6 비율 q 를 구하고,
    change_pct 오름차순 k=round(q*N) 번째 값을 X 로 삼음(<= X 이면 정확히 k 개 이상 경고)."""
    xs = {}
    for market in ("KR", "US"):
        js = [j for j in j2 if j["market"] == market and session_date(j) < IN_SAMPLE_END
              and in_window(j, states)]
        vals = sorted(float(states[j["digest"]]["change_pct"]) for j in js
                      if j["digest"] in states and states[j["digest"]]["change_pct"] is not None)
        q = sum(1 for j in js if j["p"] >= P_STRONG) / len(js)
        k = round(q * len(vals))
        x = vals[k - 1]
        share = sum(1 for v in vals if v <= round(x, 2)) / len(vals)
        xs[market] = round(x, 2)
        out(f"{market}: 창 안 J2 {len(js)}건(change_pct 있음 {len(vals)}), Jev 경고 비율 q={q:.4f}, "
            f"k={k}, X={x:.4f} → 고정값 {round(x, 2):.2f} (F 경고 비율 {share:.4f})")
    return xs


# ---------------------------------------------------------------- 자가 시험

def self_test():
    """작은 합성 픽스처로 규칙 갈래를 단언함(실제 원장 미사용)."""
    utc = dt.timezone.utc

    def at(d, hm):
        h, m = map(int, hm.split(":"))
        return dt.datetime.fromisoformat(d).replace(hour=h, minute=m, tzinfo=utc)

    def bars(*pairs):
        return [{"t": t, "o": c, "h": c, "l": c, "c": c, "v": 1} for t, c in pairs]

    d = "2026-10-22"
    states = {
        # KR A: 09:10 KST(00:10Z) p 0.7, 하락 -3% → Jev·F 모두 경고. 10:00 봉 100, 14:10 봉 110.
        "a1": {"market": "KR", "symbol": "A", "observed_at": at(d, "00:10"), "minutes_since_open": 10,
               "change_pct": -3.0, "bars": bars(("09:10", 99))},
        "a2": {"market": "KR", "symbol": "A", "observed_at": at(d, "01:00"), "minutes_since_open": 60,
               "change_pct": -3.0, "bars": bars(("10:00", 95))},  # 형성 중 값 — 늦은 state 가 덮어씀
        "a3": {"market": "KR", "symbol": "A", "observed_at": at(d, "01:10"), "minutes_since_open": 70,
               "change_pct": -3.0, "bars": bars(("10:00", 100))},
        "a4": {"market": "KR", "symbol": "A", "observed_at": at(d, "05:20"), "minutes_since_open": 320,
               "change_pct": -1.0, "bars": bars(("14:10", 110))},
        # KR B: p 0.3, 하락 -5% → F 만 경고. 10:00 없음 → 09:58 대체, 14:10 없음 → 14:07 대체.
        "b1": {"market": "KR", "symbol": "B", "observed_at": at(d, "00:20"), "minutes_since_open": 20,
               "change_pct": -5.0, "bars": bars(("09:58", 50), ("14:07", 45))},
        # KR C: p 0.2, 상승 → 기준선. 14:10·14:05~14:10 봉 없음 → exit_missing.
        "c1": {"market": "KR", "symbol": "C", "observed_at": at(d, "00:30"), "minutes_since_open": 30,
               "change_pct": 1.0, "bars": bars(("10:00", 10), ("14:04", 11))},
        # KR D: 10:00 KST 판단(창 밖) → 무시.
        "d1": {"market": "KR", "symbol": "D", "observed_at": at(d, "01:00"), "minutes_since_open": 60,
               "change_pct": -9.0, "bars": bars(("10:00", 1), ("14:10", 2))},
        # US E: msince 20(창 밖) p 0.9, msince 40 p 0.7(첫 신호), msince 50 p 0.8.
        "e0": {"market": "US", "symbol": "E", "observed_at": at(d, "13:50"), "minutes_since_open": 20,
               "change_pct": -4.0, "bars": []},
        "e1": {"market": "US", "symbol": "E", "observed_at": at(d, "14:10"), "minutes_since_open": 40,
               "change_pct": 0.5, "bars": []},
        "e2": {"market": "US", "symbol": "E", "observed_at": at(d, "14:20"), "minutes_since_open": 50,
               "change_pct": -4.0, "bars": []},
        # US G: p 0.1 두 번 → 기준선(첫 판단 라벨 null → 결측으로 셈).
        "g1": {"market": "US", "symbol": "G", "observed_at": at(d, "14:00"), "minutes_since_open": 30,
               "change_pct": 0.0, "bars": []},
        "g2": {"market": "US", "symbol": "G", "observed_at": at(d, "14:30"), "minutes_since_open": 60,
               "change_pct": 0.0, "bars": []},
    }

    def jd(i, dig, mkt, sym, hm, p):
        return {"id": i, "at": at(d, hm), "market": mkt, "symbol": sym, "p": p,
                "base_price": 1.0, "digest": dig}

    j2 = [jd("a", "a1", "KR", "A", "00:10", 0.7), jd("b", "b1", "KR", "B", "00:20", 0.3),
          jd("c", "c1", "KR", "C", "00:30", 0.2), jd("d", "d1", "KR", "D", "01:00", 0.9),
          jd("e0", "e0", "US", "E", "13:50", 0.9), jd("e1", "e1", "US", "E", "14:10", 0.7),
          jd("e2", "e2", "US", "E", "14:20", 0.8), jd("g1", "g1", "US", "G", "14:00", 0.1),
          jd("g2", "g2", "US", "G", "14:30", 0.1)]
    labels = {"e0": 0.5, "e1": 0.02, "e2": -0.03, "g1": None, "g2": 0.01}

    saved = dict(X_FROZEN)
    X_FROZEN.update({"KR": -2.0, "US": -2.0})
    try:
        r = evaluate(j2, labels, states, "out-of-sample", out=lambda s: None)
    finally:
        X_FROZEN.clear()
        X_FROZEN.update(saved)

    kr, us = r["KR"]["arms"], r["US"]["arms"]
    A, B, C = ("KR", d, "A"), ("KR", d, "B"), ("KR", d, "C")
    # Jev KR: A 신호, 늦은 state 의 10:00 봉(100) 진입 → 14:10(110) 청산 = +10%
    assert abs(kr["Jev"]["signal"][A] - 0.10) < 1e-12, kr["Jev"]
    # B 는 기준선, 09:58/14:07 대체 봉 사용 = -10%; C 는 청산 결측; D(10:00 판단) 는 창 밖
    assert abs(kr["Jev"]["base"][B] + 0.10) < 1e-12
    assert kr["Jev"]["missing"] == {"base:exit_missing": 1}, kr["Jev"]["missing"]
    assert ("KR", d, "D") not in kr["Jev"]["signal"] and ("KR", d, "D") not in kr["Jev"]["base"]
    # F KR: A(-3) 와 B(-5) 모두 X=-2 이하 → 신호, C 는 기준선(결측)
    assert set(kr["F"]["signal"]) == {A, B} and kr["F"]["base"] == {}
    # US Jev: 창 밖 e0 무시, 첫 신호 e1 라벨 +2%; G 기준선 첫 판단 g1 라벨 null → 결측
    E = ("US", d, "E")
    assert us["Jev"]["signal"] == {E: 0.02}, us["Jev"]
    assert us["Jev"]["missing"] == {"base:label_null": 1}
    # US F: e1(+0.5) 은 경고 아님, 첫 F 경고는 e2(-4) → 라벨 -3%
    assert us["F"]["signal"] == {E: -0.03}, us["F"]
    # 판정: n 부족 → INSUFFICIENT
    assert kr["Jev"]["verdict"] == "INSUFFICIENT" and us["F"]["verdict"] == "INSUFFICIENT"
    # 맞대결 셀: KR 둘 다 A, F 만 B, Jev 만 0 → Jev 만 셀이 비어 "formula is enough"
    h = r["KR"]["h2h"]
    assert (h["both"], h["jev_only"], h["f_only"]) == (1, 0, 1) and h["verdict"] == "formula is enough"
    assert r["US"]["h2h"]["verdict"] == "formula is enough"  # US: 둘 다 E, 단독 셀 0
    assert head_to_head("KR", {"Jev": {"signal": {}}, "F": {"signal": {A: 0.1}}},
                        lambda s: None)["verdict"] == "판정 불가(팔 신호 0)"
    assert head_to_head("KR", {"Jev": {"signal": {A: 0.1, B: 0.05}}, "F": {"signal": {A: 0.1, C: 0.0}}},
                        lambda s: None)["verdict"] == "Jev adds value"
    # 통과 규칙 갈래 직접 확인
    assert verdict("US", [0.01] * 15, [0.0]) == "PASS"
    assert verdict("US", [0.01] * 15, [0.02]) == "FAIL"
    assert verdict("US", [-0.01] * 15, [-0.02]) == "FAIL"
    assert verdict("US", [0.01] * 14, [0.0]) == "INSUFFICIENT"
    assert verdict("KR", [0.01] * 20, []) == "INSUFFICIENT"
    # 표본 외 세션 선택: 명목 22 결측 → 대체 29 사용
    by = {k: [1] for k in ["2026-10-23", "2026-10-26", "2026-10-27", "2026-10-28", "2026-10-29"]}
    s, _ = select_sessions("out-of-sample", "KR", by)
    assert s == ["2026-10-23", "2026-10-26", "2026-10-27", "2026-10-28", "2026-10-29"]
    by["2026-10-14"] = [1]
    assert "2026-10-14" not in select_sessions("out-of-sample", "KR", by)[0]
    assert select_sessions("in-sample", "KR", by)[0] == ["2026-10-14"]
    print("self-test OK")


def main(argv=None):
    ap = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    g = ap.add_mutually_exclusive_group(required=True)
    g.add_argument("--window", choices=["in-sample", "out-of-sample"])
    g.add_argument("--self-test", action="store_true")
    g.add_argument("--calibrate-x", action="store_true")
    a = ap.parse_args(argv)
    if a.self_test:
        self_test()
        return 0
    j2, labels, skipped = load_ledger(LEDGER)
    states, unreadable = load_states(STATES)
    print(f"원장 J2 {len(j2)}건, 라벨 {len(labels)}건, 잘린 마지막 줄 {skipped}, "
          f"state {len(states)}개(읽기 실패 {unreadable})")
    if a.calibrate_x:
        calibrate_x(j2, states)
        return 0
    evaluate(j2, labels, states, a.window)
    return 0


if __name__ == "__main__":
    sys.exit(main())

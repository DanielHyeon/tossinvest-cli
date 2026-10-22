#!/usr/bin/env python3
"""a128 사전등록 가설 평가기 — J2 강한 경고(Jev) 와 공식 경고(F) 의 4 팔 + 맞대결.

정본 규칙: prereg-2026-10-22-j2-warning-hypotheses.md (같은 디렉터리).
이 스크립트는 그 문서의 규칙을 기계로 옮긴 것이며 규칙을 바꾸지 않음.

- 표준 라이브러리만 사용, 읽기 전용(원장·state 파일을 쓰거나 옮기지 않음).
- 주문·계좌·네트워크 접근 없음.
- 실행: python3 -I prereg_eval.py --window in-sample | out-of-sample
        python3 -I prereg_eval.py --self-test
        python3 -I prereg_eval.py --calibrate-x   (F 임계값 X 도출 과정 재현 — 감사용)
        python3 -I prereg_eval.py --d2            (D2 — prereg-2026-10-23-d2-twice-daily-index-correlation.md)
"""
from __future__ import annotations

import argparse
import datetime as dt
import json
import os
import statistics
import sys
from zoneinfo import ZoneInfo

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
            "index_pct": index_pct(d),  # D2 전용 — 기존 네 팔은 읽지 않음
        }
    return states, unreadable


def index_pct(d):
    """D2 지수 등락(%) — KR: indices[symbol=KOSPI].change_today_pct, US: major_indices[code=SPX.CBI].change_pct_vs_base.
    값이 없으면 None(사전등록 D2 4절)."""
    mc = d.get("market_context")
    if not isinstance(mc, dict):
        return None
    if d.get("market") == "KR":
        rows, key, field, want = mc.get("indices") or [], "symbol", "change_today_pct", "KOSPI"
    else:
        rows, key, field, want = mc.get("major_indices") or [], "code", "change_pct_vs_base", "SPX.CBI"
    if not isinstance(rows, list):
        return None
    for r in rows:
        if isinstance(r, dict) and r.get(key) == want and r.get(field) is not None:
            try:
                return float(r[field])
            except (TypeError, ValueError):
                return None  # 숫자가 아닌 값 = 결측(기존 팔 적재를 멈추지 않음)
    return None


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


# ---------------------------------------------------------------- D2(사전등록 2026-10-23)
# 정본: prereg-2026-10-23-d2-twice-daily-index-correlation.md. 기존 네 팔과 독립 — 위 함수들의 동작을 바꾸지 않음.



def d2_tz(market):
    """KR = KST, US = America/New_York(서머타임 자동). ZoneInfo 는 D2 경로에서만 만듦 — import 시 tzdata 의존을 피함."""
    return KST if market == "KR" else ZoneInfo("America/New_York")

D2_START, D2_END = "2026-10-23", "2026-12-24"
# 창 끝 = US 12-24 세션 13:00 ET 진입의 +60분 청산 시각(14:00 EST = 19:00Z). 이 시각 전 실행은 중간 집계(판정 아님).
D2_FINAL_AFTER = dt.datetime(2026, 12, 24, 19, 0, tzinfo=dt.timezone.utc)
D2_ENTRIES = ("10:00", "13:00")
D2_LAG_MAX = dt.timedelta(minutes=10)   # 지수 관측 차이 상한(프로브 회차 간격 600초)
D2_BAR_WINDOW = dt.timedelta(minutes=5)  # [T-5분, T] 안 가장 가까운 이전 봉
D2_DATES_MIN = 5                        # B1·B3 각 날짜 수 하한
D2_BINS = ("B1 <=-3", "B2 (-3,-1]", "B3 (-1,0)", "B4 >=0")
D2_TYPE_GROUP = {"STOCK": "STOCK", "FOREIGN_STOCK": "STOCK", "ETF": "ETF", "FOREIGN_ETF": "ETF", "ETN": "ETN",
                 "DEPOSITARY_RECEIPT": "기타", "INFRASTRUCTURE_FUND": "기타", "REIT": "기타", "STOCK_WARRANTS": "기타"}
D2_RECEIPTS = os.path.join(HERE, "d2-receipts")


def d2_bin(v):
    """지수 등락 v(%) 의 구간. B1: v<=-3, B2: -3<v<=-1, B3: -1<v<0, B4: v>=0."""
    return D2_BINS[0] if v <= -3 else D2_BINS[1] if v <= -1 else D2_BINS[2] if v < 0 else D2_BINS[3]


def d2_type(sym, types):
    """securityType → 갈래. enum 밖·조회 실패·영수증 없음 = 미분류(사전등록 D2 7절)."""
    return D2_TYPE_GROUP.get(types.get(sym), "미분류")


def d2_label_time(market, date, label):
    """state 봉 라벨(KST "HH:MM") → aware 시각. KR 은 세션 날짜 그대로.
    US 세션(UTC 13:30~21:00)은 KST 22:30~06:00 이라 자정을 넘음 — 12시 이후 라벨은 세션 날짜, 그 전은 다음 날."""
    h, m = map(int, label.split(":"))
    day = dt.date.fromisoformat(date)
    if market == "US" and h < 12:
        day += dt.timedelta(days=1)
    return dt.datetime(day.year, day.month, day.day, h, m, tzinfo=KST)


def d2_books(states, receipts):
    """주식-일별 1분봉 책 두 개(state 합집합, 영수증). 키 = 시각, 값 = 종가. state 는 build_bars 관례(늦은 관측 우선)."""
    sbook = {sd: {d2_label_time(sd[0], sd[1], t): c for t, (_, c) in b.items()}
             for sd, b in build_bars(states).items()}
    rbook = {}
    for (market, date), rows in receipts.items():
        for sym, bars in rows.items():
            rbook[(market, date, sym)] = {parse_at(b["time"]): float(b["close"]) for b in bars}
    return sbook, rbook


def d2_pick(sbook, rbook, sd, t):
    """시각 t 의 [t-5분, t] 안 가장 가까운 이전 봉 종가와 출처. state 를 먼저, 없을 때만 영수증. 둘 다 없으면 (None, None)."""
    for src, book in (("state", sbook), ("receipt", rbook)):
        b = book.get(sd) or {}
        cands = [x for x in b if t - D2_BAR_WINDOW <= x <= t]
        if cands:
            return b[max(cands)], src
    return None, None


def d2_index(states):
    """(market, 세션 날짜) → [(관측 시각, 지수 등락)] 시각순. 값 없는 state 는 뺌."""
    out = {}
    for st in states.values():
        if st["observed_at"] is None or st.get("index_pct") is None:
            continue
        out.setdefault((st["market"], st["observed_at"].date().isoformat()), []).append(
            (st["observed_at"], st["index_pct"]))
    for v in out.values():
        v.sort()
    return out


def spearman(xs, ys):
    """스피어만 ρ(동순위 평균 순위). 표본 3 미만·분산 0 이면 None."""
    if len(xs) < 3:
        return None

    def ranks(v):
        order = sorted(range(len(v)), key=lambda i: v[i])
        r = [0.0] * len(v)
        i = 0
        while i < len(order):
            k = i
            while k + 1 < len(order) and v[order[k + 1]] == v[order[i]]:
                k += 1
            for x in order[i:k + 1]:
                r[x] = (i + k) / 2 + 1
            i = k + 1
        return r
    rx, ry = ranks(xs), ranks(ys)
    mx, my = statistics.fmean(rx), statistics.fmean(ry)
    sxy = sum((a - mx) * (b - my) for a, b in zip(rx, ry))
    sxx, syy = sum((a - mx) ** 2 for a in rx), sum((b - my) ** 2 for b in ry)
    return None if sxx == 0 or syy == 0 else sxy / (sxx * syy) ** 0.5


def d2_verdict(day_means_by_bin):
    """가설: B1 날짜 평균 순수익의 평균 > B3 의 것. B1·B3 날짜 수 각 5 이상일 때만."""
    b1, b3 = day_means_by_bin.get(D2_BINS[0], []), day_means_by_bin.get(D2_BINS[2], [])
    if len(b1) < D2_DATES_MIN or len(b3) < D2_DATES_MIN:
        return "INSUFFICIENT"
    return "SUPPORTED" if mean(b1) > mean(b3) else "NOT_SUPPORTED"


def d2_trades(j2, states, receipts, types):
    """D2 가상 거래 목록과 결측 집계. 반환: (trades, missing). trade = dict(market, entry, date, symbol,
    signal, type, net, idx, lag_min, src)."""
    sbook, rbook = d2_books(states, receipts)
    idx = d2_index(states)
    lists = {}  # (market, date) → {symbol: 목록 회차 J2 p}
    ran = set()  # 판단이 하나라도 있는 (market, date) — 목록 없는 날을 세려고
    for j in sorted(j2, key=lambda j: j["at"]):  # 같은 회차 같은 종목 판단이 둘이면 시각 첫 행
        d = session_date(j)
        if not (D2_START <= d <= D2_END):
            continue
        ran.add((j["market"], d))
        loc = j["at"].astimezone(d2_tz(j["market"]))
        if loc.hour == 10 and loc.minute < 10:
            row = lists.setdefault((j["market"], d), {})
            row.setdefault(j["symbol"], j["p"])
    trades, missing = [], {}

    def miss(key):
        missing[key] = missing.get(key, 0) + 1

    for market, d in sorted(ran - set(lists)):
        miss(f"{market}:no_list_dates")  # 프로브는 돌았지만 10:00 회차가 없음 → 그날 두 진입 모두 결측

    for (market, d), syms in sorted(lists.items()):
        tz = d2_tz(market)
        day = dt.date.fromisoformat(d)
        for entry in D2_ENTRIES:
            h, m = map(int, entry.split(":"))
            t_in = dt.datetime(day.year, day.month, day.day, h, m, tzinfo=tz)
            t_out = (dt.datetime(day.year, day.month, day.day, 14, 10, tzinfo=tz) if market == "KR"
                     else t_in + dt.timedelta(minutes=60))
            obs = [(o, v) for o, v in idx.get((market, d), []) if o <= t_in]
            if not obs or t_in - obs[-1][0] > D2_LAG_MAX:
                miss(f"{market}-{entry}:index_missing_dates")
                continue
            o, v = obs[-1]
            for sym, p in sorted(syms.items()):
                sd = (market, d, sym)
                pin, sin = d2_pick(sbook, rbook, sd, t_in)
                if pin is None:
                    miss(f"{market}-{entry}:entry_missing")
                    continue
                pout, sout = d2_pick(sbook, rbook, sd, t_out)
                if pout is None:
                    miss(f"{market}-{entry}:exit_missing")
                    continue
                trades.append({"market": market, "entry": entry, "date": d, "symbol": sym,
                               "signal": p >= P_STRONG, "type": d2_type(sym, types),
                               "net": pout / pin - 1.0 - COST_RT[market], "idx": v,
                               "lag_min": (t_in - o).total_seconds() / 60, "src": (sin, sout)})
    return trades, missing


def d2_load_receipts(root=D2_RECEIPTS):
    """영수증(읽기 전용). minute-bars/<MARKET>-<YYYY-MM-DD>.json = {"bars": {sym: [{"time": ISO, "close": n}]}, ...},
    security-types.json = {"types": {sym: securityType}, ...}. 없으면 빈 값 — 유형은 전부 미분류, 보충 없음."""
    receipts, types = {}, {}
    mdir = os.path.join(root, "minute-bars")
    if os.path.isdir(mdir):
        for name in sorted(os.listdir(mdir)):
            if name.endswith(".json"):
                market, date = name[:-5].split("-", 1)
                with open(os.path.join(mdir, name), "rb") as fh:
                    receipts[(market, date)] = json.loads(fh.read().decode("utf-8")).get("bars") or {}
    tpath = os.path.join(root, "security-types.json")
    if os.path.isfile(tpath):
        with open(tpath, "rb") as fh:
            raw = json.loads(fh.read().decode("utf-8"))
        # 문서(7절) 형식 = 평면 {"<symbol>": "<securityType>"}(+ 조회 시각·호출 수 칸). {"types": {...}} 래퍼도 받음.
        flat = raw.get("types") if isinstance(raw.get("types"), dict) else raw
        types = {k: v for k, v in flat.items() if isinstance(v, str)}
    return receipts, types


def evaluate_d2(j2, states, receipts, types, now, out=print):
    """D2 평가. 칸(시장×진입 시각)마다 날짜 단위 주 분석 + 종목 단위 보조. 반환: {칸: {verdict, rho, bins, dates}}."""
    trades, missing = d2_trades(j2, states, receipts, types)
    final = now >= D2_FINAL_AFTER
    out(f"\n=== D2 — 창 {D2_START}~{D2_END}; "
        f"{'최종 판정' if final else '중간 집계(서술용, 판정 아님)'}; 거래 {len(trades)}건; 유형 영수증 "
        f"{'있음' if types else '없음(전부 미분류)'} ===")
    srcs = {}
    for t in trades:
        for s in t["src"]:
            srcs[s] = srcs.get(s, 0) + 1
    out(f"봉 출처(진입+청산) {srcs or '없음'}; 결측 {missing or '없음'}")
    res = {}
    for market in ("KR", "US"):
        for entry in D2_ENTRIES:
            cell = [t for t in trades if t["market"] == market and t["entry"] == entry]
            by_day = {}
            for t in cell:
                by_day.setdefault(t["date"], []).append(t)
            days = sorted(by_day)
            dv = [by_day[d][0]["idx"] for d in days]
            dm = [mean([t["net"] for t in by_day[d]]) for d in days]
            bins = {}
            for v, m in zip(dv, dm):
                bins.setdefault(d2_bin(v), []).append(m)
            rho = spearman(dv, dm)
            v = d2_verdict(bins)
            key = f"{market}-{entry}"
            res[key] = {"verdict": v if final else None, "interim": None if final else v, "rho": rho,
                         "dates": len(days), "bins": {b: len(x) for b, x in bins.items()}}
            out(f"\n-- {key} ({'KST' if market == 'KR' else 'ET'}; 비용 {COST_RT[market]*100:.2f}%) — 날짜 {len(days)}개, "
                f"스피어만 ρ(날짜) {'-' if rho is None else f'{rho:+.3f}'}")
            for b in D2_BINS:
                xs = bins.get(b, [])
                out(f"   {b:<11} 날짜 {len(xs):>3}  날짜 평균 순수익의 평균 {fmt(mean(xs)):>7}%")
            for d, v_, m_ in zip(days, dv, dm):
                lags = sorted({t["lag_min"] for t in by_day[d]})
                out(f"   {d} 지수 {v_:+.2f}% 차이 {'/'.join(f'{x:g}' for x in lags)}분 n={len(by_day[d])} 날짜 평균 {fmt(m_)}%")
            out(f"   가설(B1 > B3, 각 날짜 >= {D2_DATES_MIN}): "
                f"{v if final else v + ' — 중간 집계(판정 아님)'}")
            # 보조(서술용): 종목 단위 ρ 와 갈래별 평균
            srho = spearman([t["idx"] for t in cell], [t["net"] for t in cell])
            out(f"   보조 종목 단위: n={len(cell)} ρ {'-' if srho is None else f'{srho:+.3f}'}")
            groups = {}
            for t in cell:
                groups.setdefault((("신호" if t["signal"] else "비신호"), t["type"]), []).append(t["net"])
            for g in sorted(groups):
                out(f"     {g[0]}/{g[1]}: n={len(groups[g])} 평균 순수익 {fmt(mean(groups[g]))}%")
    out("\n주의: D2 는 가상 매수 집계임. 실주문 0. SUPPORTED 도 수익성 결론이 아님.")
    return res


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
    d2_self_test()
    print("self-test OK")


def d2_self_test():
    """D2 합성 픽스처: 하락일/상승일/지수 결측·차이 초과/ETF·ETN·미분류/영수증 보충/US 자정 넘는 창/서머타임."""
    utc = dt.timezone.utc

    def z(s):
        return dt.datetime.fromisoformat(s).replace(tzinfo=utc)

    def st(mkt, sym, obs, bars=(), idx=None):
        return {"market": mkt, "symbol": sym, "observed_at": z(obs), "minutes_since_open": 0, "change_pct": 0.0,
                "bars": [{"t": t, "o": c, "h": c, "l": c, "c": c, "v": 1} for t, c in bars], "index_pct": idx}

    states = {
        # KR 10-23: 10:00 KOSPI -3.5(B1, 차이 0분), 13:00 -0.5(B3, 차이 5분)
        "k1": st("KR", "IDX", "2026-10-23T01:00:00", idx=-3.5),
        "k2": st("KR", "IDX", "2026-10-23T03:55:00", idx=-0.5),
        "x1": st("KR", "X", "2026-10-23T01:00:00", [("10:00", 100)]),
        "x2": st("KR", "X", "2026-10-23T05:10:00", [("13:00", 110), ("14:10", 121)]),
        "y1": st("KR", "Y", "2026-10-23T01:00:00", [("09:57", 50)]),   # 10:00 없음 → 09:57 대체, 청산은 영수증
        "z1": st("KR", "Z", "2026-10-23T01:00:00", [("10:00", 10)]),   # 청산 결측
        "w1": st("KR", "W", "2026-10-23T00:50:00", [("10:00", 5), ("14:10", 6)]),  # 09:50 회차만 → 목록 밖
        "w2": st("KR", "W2", "2026-10-23T05:10:00", [("10:00", 5), ("13:00", 5), ("14:10", 6)]),  # 10:10 회차만 → 목록 밖
        # KR 10-26: 지수 관측이 09:40 뿐 → 10:00 차이 20분 > 10 → 결측
        "k3": st("KR", "IDX", "2026-10-26T00:40:00", idx=-5.0),
        "x3": st("KR", "X", "2026-10-26T05:10:00", [("10:00", 1), ("13:00", 1), ("14:10", 2)]),
        # KR 10-22: 창 밖
        "k0": st("KR", "IDX", "2026-10-22T01:00:00", idx=-9.0),
        "x0": st("KR", "X", "2026-10-22T05:10:00", [("10:00", 1), ("14:10", 2)]),
        # US 10-23(EDT): 10:00 ET = 14:00Z = 23:00 KST, 11:00 ET = 00:00 KST(다음 날), 13:00 ET = 02:00 KST, 14:00 ET = 03:00 KST
        "s1": st("US", "IDX", "2026-10-23T14:00:00", idx=-1.5),
        "s2": st("US", "IDX", "2026-10-23T16:55:00", idx=0.2),
        "u1": st("US", "U", "2026-10-23T14:00:00", [("23:00", 20)]),
        "u2": st("US", "U", "2026-10-23T15:10:00", [("23:57", 22)]),   # [23:55, 00:00] 창 — 자정 앞 봉
        "u3": st("US", "U", "2026-10-23T18:00:00", [("02:00", 30), ("03:00", 27)]),
        # US 11-02(EST): 10:00 ET = 15:00Z. 14:00Z 판단(09:00 ET)은 목록 밖
        "s3": st("US", "IDX", "2026-11-02T15:00:00", idx=-4.0),
        "v1": st("US", "V", "2026-11-02T16:00:00", [("00:00", 40), ("01:00", 44)]),
        "v2": st("US", "V2", "2026-11-02T16:00:00", [("00:00", 40), ("01:00", 44)]),
    }

    def jd(sym, mkt, at, p):
        return {"id": sym + at, "at": z(at), "market": mkt, "symbol": sym, "p": p, "base_price": 1.0, "digest": "-"}

    j2 = [jd("X", "KR", "2026-10-23T01:00:00", 0.7), jd("Y", "KR", "2026-10-23T01:00:00", 0.2),
          jd("Z", "KR", "2026-10-23T01:00:00", 0.3), jd("W", "KR", "2026-10-23T00:50:00", 0.9),
          jd("W2", "KR", "2026-10-23T01:10:00", 0.9),
          jd("X", "KR", "2026-10-26T01:00:00", 0.1), jd("X", "KR", "2026-10-22T01:00:00", 0.1),
          jd("U", "US", "2026-10-23T14:00:00", 0.65), jd("V", "US", "2026-11-02T15:00:00", 0.1),
          jd("V2", "US", "2026-11-02T14:00:00", 0.1)]
    receipts = {("KR", "2026-10-23"): {"Y": [{"time": "2026-10-23T14:08:00+09:00", "close": 55}]}}
    types = {"X": "ETF", "Y": "ETN", "U": "FOREIGN_STOCK", "Q": "WEIRD"}
    trades, missing = d2_trades(j2, states, receipts, types)
    got = {(t["market"], t["entry"], t["date"], t["symbol"]): t for t in trades}
    c_kr, c_us = COST_RT["KR"], COST_RT["US"]
    # KR X: 10:00 진입 100 → 14:10 121, 13:00 진입 110 → 121. ETF, 신호(p 0.7). 지수 10:00=-3.5(B1), 13:00=-0.5(차이 5분)
    x10, x13 = got[("KR", "10:00", "2026-10-23", "X")], got[("KR", "13:00", "2026-10-23", "X")]
    assert abs(x10["net"] - (0.21 - c_kr)) < 1e-12 and x10["idx"] == -3.5 and x10["lag_min"] == 0
    assert abs(x13["net"] - (0.10 - c_kr)) < 1e-12 and x13["idx"] == -0.5 and x13["lag_min"] == 5
    assert x10["type"] == "ETF" and x10["signal"] and x13["signal"]  # 13:00 도 목록 회차 분류
    # KR Y: 09:57 대체 진입(state) + 영수증 14:08 청산, ETN, 비신호. 13:00 진입은 봉 없음 → 결측
    y10 = got[("KR", "10:00", "2026-10-23", "Y")]
    assert abs(y10["net"] - (0.10 - c_kr)) < 1e-12 and y10["src"] == ("state", "receipt") and y10["type"] == "ETN"
    assert not y10["signal"]
    # Z 는 청산 결측(미분류), W 는 09:50 회차라 목록 밖, 10-22 는 창 밖, 10-26 은 지수 차이 초과
    assert ("KR", "10:00", "2026-10-23", "Z") not in got and ("KR", "10:00", "2026-10-23", "W") not in got
    assert not any(k[3] == "W2" for k in got)  # 10:10 회차는 목록 회차가 아님
    assert not any(k[2] in ("2026-10-22", "2026-10-26") for k in got)
    assert missing == {"KR-10:00:exit_missing": 1, "KR-13:00:entry_missing": 2,
                       "KR-10:00:index_missing_dates": 1, "KR-13:00:index_missing_dates": 1,
                       "US-13:00:index_missing_dates": 1}, missing
    assert d2_type("Z", types) == "미분류" and d2_type("Q", types) == "미분류" and d2_type("U", types) == "STOCK"
    # US U(EDT): 10:00 ET 23:00 KST 진입 20 → 11:00 ET 창의 23:57 봉 22(자정 앞), 13:00 ET 30 → 14:00 ET 27
    u10, u13 = got[("US", "10:00", "2026-10-23", "U")], got[("US", "13:00", "2026-10-23", "U")]
    assert abs(u10["net"] - (0.10 - c_us)) < 1e-12 and u10["idx"] == -1.5 and u10["signal"]
    assert abs(u13["net"] - (-0.10 - c_us)) < 1e-12 and u13["idx"] == 0.2 and u13["lag_min"] == 5
    # US V(EST): 10:00 ET = 15:00Z 회차만 목록. V2(14:00Z = 09:00 EST)는 목록 밖
    assert abs(got[("US", "10:00", "2026-11-02", "V")]["net"] - (0.10 - c_us)) < 1e-12
    assert not any(k[3] == "V2" for k in got)
    # 구간 경계와 판정·상관
    assert [d2_bin(v) for v in (-3.0, -2.9999, -1.0, -0.0001, 0.0, 1.0)] == [
        D2_BINS[0], D2_BINS[1], D2_BINS[1], D2_BINS[2], D2_BINS[3], D2_BINS[3]]
    assert d2_verdict({D2_BINS[0]: [0.02] * 5, D2_BINS[2]: [0.01] * 5}) == "SUPPORTED"
    assert d2_verdict({D2_BINS[0]: [0.01] * 5, D2_BINS[2]: [0.01] * 5}) == "NOT_SUPPORTED"
    assert d2_verdict({D2_BINS[0]: [0.02] * 4, D2_BINS[2]: [0.01] * 9}) == "INSUFFICIENT"
    assert d2_verdict({D2_BINS[0]: [0.02] * 9}) == "INSUFFICIENT"
    assert spearman([1, 2, 3], [3, 2, 1]) == -1.0 and spearman([1, 2], [1, 2]) is None
    assert abs(spearman([1, 1, 2], [1, 2, 3]) - 0.8660254037844387) < 1e-12
    assert spearman([1, 1, 1], [1, 2, 3]) is None
    # 창 끝 전 = 중간 집계 표시, 뒤 = 최종
    lines = []
    r = evaluate_d2(j2, states, receipts, types, z("2026-12-24T18:59:00"), out=lines.append)
    assert "중간 집계(서술용, 판정 아님)" in lines[0]
    assert r["KR-10:00"]["verdict"] is None and r["KR-10:00"]["interim"] == "INSUFFICIENT"
    assert any(s.startswith("   2026-10-23 지수 -3.50% 차이 0분 n=2") for s in lines), lines
    assert any("— 중간 집계(판정 아님)" in s for s in lines)
    lines = []
    r = evaluate_d2(j2, states, receipts, types, D2_FINAL_AFTER, out=lines.append)
    assert r["KR-10:00"]["verdict"] == "INSUFFICIENT" and r["KR-10:00"]["interim"] is None
    assert "최종 판정" in lines[0] and not any("— 중간 집계(판정 아님)" in s for s in lines)
    assert index_pct({"market": "KR", "market_context": {"indices": [{"symbol": "KOSDAQ", "change_today_pct": 1},
                                                                     {"symbol": "KOSPI", "change_today_pct": -2}]}}) == -2.0
    assert index_pct({"market": "US", "market_context": {"major_indices": [{"code": "SPX.CBI",
                                                                            "change_pct_vs_base": 0.5}]}}) == 0.5
    assert index_pct({"market": "US", "market_context": {}}) is None
    assert index_pct({"market": "KR", "market_context": "x"}) is None
    assert index_pct({"market": "KR", "market_context": {"indices": [{"symbol": "KOSPI", "change_today_pct": "n/a"}]}}) is None
    assert index_pct({"market": "US", "market_context": {"major_indices": None}}) is None
    # 같은 회차 중복 판단은 시각 첫 행(원장 순서 무관), 목록 없는 날은 세어 보고
    dup = [jd("X", "KR", "2026-10-23T01:05:00", 0.1), jd("X", "KR", "2026-10-23T01:00:00", 0.7)]
    t2, m2 = d2_trades(dup + [jd("Q", "KR", "2026-10-27T00:50:00", 0.1)], states, {}, {})
    assert all(t["signal"] for t in t2 if t["symbol"] == "X") and m2.get("KR:no_list_dates") == 1, m2
    # 유형 영수증: 평면 맵(문서 7절)과 {"types": …} 래퍼 둘 다, 파일 없으면 빈 값(전부 미분류)
    import tempfile
    with tempfile.TemporaryDirectory() as tmp:
        assert d2_load_receipts(tmp) == ({}, {})
        tpath = os.path.join(tmp, "security-types.json")
        for body in ({"X": "ETF", "fetched_at": "2026-10-23T06:00:00Z", "calls": 1}, {"types": {"X": "ETF"}, "calls": 1}):
            with open(tpath, "w", encoding="utf-8") as fh:
                json.dump(body, fh)
            assert d2_type("X", d2_load_receipts(tmp)[1]) == "ETF"


def main(argv=None):
    ap = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    g = ap.add_mutually_exclusive_group(required=True)
    g.add_argument("--window", choices=["in-sample", "out-of-sample"])
    g.add_argument("--self-test", action="store_true")
    g.add_argument("--calibrate-x", action="store_true")
    g.add_argument("--d2", action="store_true")
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
    if a.d2:
        receipts, types = d2_load_receipts()
        evaluate_d2(j2, states, receipts, types, dt.datetime.now(dt.timezone.utc))
        return 0
    evaluate(j2, labels, states, a.window)
    return 0


if __name__ == "__main__":
    sys.exit(main())

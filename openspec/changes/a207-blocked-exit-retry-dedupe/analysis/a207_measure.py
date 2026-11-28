#!/usr/bin/env python3
"""a207 막힌 청산 재시도 빈도 측정 하네스 — 집계만 출력, 운영 저널에는 읽기 전용으로만 닿음.

사용:
  python3 -I a207_measure.py <원본 journal.db> <작업 디렉터리>

동작: 원본을 `?mode=ro` 로 열고 SQLite backup API 로 작업 디렉터리에 사본을 만든 뒤 사본에서 집계만 하고 사본을 지움.
쓰기 · 시나리오 없음(a206 하네스와 달리 UPDATE · VACUUM 0).

출력에는 계좌 · 종목 · 수량 · 금액 · 가격 · 주문/포지션 id 가 없음(건수 · 바이트 · 월 · 사유 코드 · 분포만).
"""
import collections
import os
import sqlite3
import sys

HOOK_ACTIONS = ("PROPOSAL_REFUSED", "PROPOSAL_CANCELLED", "PROPOSAL_FILLED", "OPENED", "COMPLETED", "ADJUSTMENT_CLOSED")
JSON_COLS = ("saved_snapshot_json", "recomputed_snapshot_json", "effective_snapshot_json")
FLAT_EVIDENCE = ("position_generation", "policy_id", "policy_version", "policy_digest", "snapshot_id", "decision_id",
                 "observation_id", "next_target", "next_protection", "observation_source", "observed_at",
                 "projected_quantity", "proposal_ratio", "state_only", "suppressed_reason", "effective_source",
                 "arm_suppressed_reason")
BASE_COLS = ("id", "position_id", "observed_price", "high_water", "baseline_after", "level_after", "action",
             "proposed_intent_id", "created_at", "lifecycle_generation")
BUCKETS = (1, 2, 10, 100, 1000, 10000)


def refuse_live(path: str) -> None:
    live = os.path.realpath(os.path.expanduser("~/.config/tossctl"))
    real = os.path.realpath(path)
    if real == live or real.startswith(live + os.sep):
        sys.exit(f"refuse: {path} is the live journal directory or under it")


def bucket(n: int) -> str:
    for b in BUCKETS:
        if n <= b:
            return f"<={b}"
    return f">{BUCKETS[-1]}"


def length_sum(cols, prefix=""):
    return "+".join(f"length(coalesce({prefix}{c},''))" for c in cols)


def main() -> None:
    if len(sys.argv) != 3:
        sys.exit(__doc__)
    src_path, work = sys.argv[1], sys.argv[2]
    refuse_live(work)
    os.makedirs(work, exist_ok=True)
    base = os.path.join(work, "a207-base.db")
    if os.path.realpath(src_path) == os.path.realpath(base):
        sys.exit("refuse: the source is the copy path")
    src = sqlite3.connect("file:" + os.path.abspath(src_path) + "?mode=ro", uri=True)
    c = sqlite3.connect(base)
    src.backup(c)
    src.close()
    try:
        report(c)
    finally:
        c.close()
        os.remove(base)


def report(c) -> None:
    q = lambda sql, *a: c.execute(sql, a).fetchall()
    print("## copy pragmas user_version, page_count:", [c.execute(f"PRAGMA {p}").fetchone()[0] for p in ("user_version", "page_count")])
    print("## exit_events rows, max created_at month:", q("SELECT count(*), max(substr(created_at,1,10)) FROM exit_events")[0])

    # 발의(무장) 행: 훅 행이 아닌 proposed_intent_id 를 가진 행.
    hooks = ",".join(f"'{a}'" for a in HOOK_ACTIONS)
    arms = q(f"""SELECT e.id, e.position_id, coalesce(e.lifecycle_generation,1), e.action, coalesce(e.level_after,''),
                        e.proposed_intent_id, e.created_at,
                        ({length_sum(JSON_COLS, 'e.')}), ({length_sum(FLAT_EVIDENCE, 'e.')}), ({length_sum(BASE_COLS, 'e.')}),
                        coalesce(e.effective_source,''), (e.recomputed_snapshot_json IS NOT NULL AND e.effective_snapshot_json IS NOT NULL)
                   FROM exit_events e
                  WHERE e.proposed_intent_id IS NOT NULL AND e.proposed_intent_id <> ''
                    AND coalesce(e.action,'') NOT IN ({hooks})
                  ORDER BY e.position_id, e.id""")
    # 그 intent 의 해제 훅(PROPOSAL_REFUSED · PROPOSAL_CANCELLED) 유무 = 「거절 · 해제되어 재무장 대상이 됨」.
    released = dict(q(f"""SELECT proposed_intent_id, max(action) FROM exit_events
                          WHERE action IN ('PROPOSAL_REFUSED','PROPOSAL_CANCELLED') GROUP BY 1"""))
    # 그 intent 의 마지막 PLACE attempt 사유(attempt 없으면 '<no attempt>').
    reason = dict(q("""SELECT intent_id, reason_code FROM mutation_attempts a WHERE kind='PLACE'
                         AND attempt_no = (SELECT max(attempt_no) FROM mutation_attempts b WHERE b.intent_id=a.intent_id AND b.kind='PLACE')"""))
    print("## arm rows:", len(arms))

    # 1. 월 x 사유: 해제된 무장 행(= 재무장을 부른 행) 수
    by_month = collections.Counter()
    for r in arms:
        if r[5] in released:
            by_month[(r[6][:7], reason.get(r[5], "<no attempt>"), released[r[5]])] += 1
    print("## released arm rows by month x PLACE reason x release hook")
    for k in sorted(by_month):
        print("  ", k, by_month[k])

    # 2. 연속 구간(streak): 같은 포지션의 연속한 해제된 무장 행 가운데 (세대, action, 수준, 사유)가 같은 최대 구간.
    #    구간 끝: 다음 무장이 해제되지 않음(제출 수용 · 결과 미상) = resolved, 다음 무장이 다른 키 = key_changed, 다음 무장 없음 = no_next_arm.
    streaks = []  # (month, reason, length, seconds, end)
    episodes = []  # 키와 무관하게 연속 해제 구간(막힘 에피소드): (month, first_reason, length, seconds, end)
    by_pos = collections.defaultdict(list)
    for r in arms:
        by_pos[r[1]].append(r)
    import datetime

    def ts(s):
        return datetime.datetime.fromisoformat(s.replace("Z", "+00:00"))

    for rows in by_pos.values():
        i = 0
        while i < len(rows):
            r = rows[i]
            if r[5] not in released:
                i += 1
                continue
            # 에피소드
            j = i
            while j + 1 < len(rows) and rows[j + 1][5] in released:
                j += 1
            end = "no_next_arm" if j + 1 >= len(rows) else "resolved"
            episodes.append((r[6][:7], reason.get(r[5], "<no attempt>"), j - i + 1,
                             (ts(rows[j][6]) - ts(r[6])).total_seconds(), end))
            # 에피소드 안의 키 구간
            k = i
            while k <= j:
                key = (rows[k][2], rows[k][3], rows[k][4], reason.get(rows[k][5], "<no attempt>"))
                m = k
                while m + 1 <= j and (rows[m + 1][2], rows[m + 1][3], rows[m + 1][4], reason.get(rows[m + 1][5], "<no attempt>")) == key:
                    m += 1
                send = "key_changed" if m < j else end
                streaks.append((rows[k][6][:7], key[3], m - k + 1, (ts(rows[m][6]) - ts(rows[k][6])).total_seconds(), send))
                k = m + 1
            i = j + 1

    def summarize(name, items):
        print(f"## {name}: count {len(items)}; by month x reason: count, sum length, max length, max seconds")
        g = collections.defaultdict(list)
        for it in items:
            g[(it[0], it[1])].append(it)
        for k in sorted(g):
            v = g[k]
            print("  ", k, len(v), sum(x[2] for x in v), max(x[2] for x in v), int(max(x[3] for x in v)))
        print(f"## {name}: length buckets (arms per streak) -> count")
        print("  ", sorted(collections.Counter(bucket(x[2]) for x in items).items(), key=lambda kv: BUCKETS.index(int(kv[0].lstrip('<=>'))) + (0.5 if kv[0].startswith('>') else 0)))
        print(f"## {name}: end kind -> count")
        print("  ", sorted(collections.Counter(x[4] for x in items).items()))
        lengths = sorted(x[2] for x in items)
        if lengths:
            pct = lambda p: lengths[min(len(lengths) - 1, int(p * len(lengths)))]
            print(f"## {name}: length p50 {pct(0.5)} p90 {pct(0.9)} p99 {pct(0.99)} max {lengths[-1]}; rows in streaks of length>1: {sum(l for l in lengths if l > 1)}")
            print(f"## {name}: repeat rows (length-1 summed, i.e. rows after the first of each streak): {sum(l - 1 for l in lengths)}")

    summarize("key streaks (generation, action, level, reason)", streaks)
    summarize("block episodes (consecutive released arms, any key)", episodes)
    print("## block episodes, each: first month, first reason, arms, seconds first->last, seconds per arm, end")
    for e in sorted(episodes):
        print("  ", e[0], e[1], e[2], int(e[3]), round(e[3] / (e[2] - 1), 1) if e[2] > 1 else "-", e[4])

    # 2b. design D1 · D2 그대로의 압축 대상(리뷰 r1 #2): 이 무장에 증거(recomputed · effective)가 있고 effective_source='recomputed' 이며,
    #     같은 포지션의 직전 무장이 같은 (세대, action, level_after) 이고 그 intent 에 해제 훅이 있음. 사유 · 이 무장의 해제 여부는 보지 않음.
    d2 = d2_bytes = d2_unreleased = 0
    for rows in by_pos.values():
        for prev, cur in zip(rows, rows[1:]):
            if cur[11] and cur[10] == "recomputed" and prev[5] in released and (prev[2], prev[3], prev[4]) == (cur[2], cur[3], cur[4]):
                d2 += 1
                d2_bytes += cur[7]
                if cur[5] not in released:
                    d2_unreleased += 1
    print("## D1+D2 compaction targets (exact predicate): rows, their JSON bytes, of which the arm itself was not released (accepted / in doubt / still armed):",
          d2, d2_bytes, d2_unreleased)
    print("## D1+D2 not-released targets by the arm's last PLACE reason:",
          sorted(collections.Counter(reason.get(cur[5], "<no attempt>") for rows in by_pos.values() for prev, cur in zip(rows, rows[1:])
                 if cur[11] and cur[10] == "recomputed" and prev[5] in released and (prev[2], prev[3], prev[4]) == (cur[2], cur[3], cur[4])
                 and cur[5] not in released).items()))
    # 에피소드가 'resolved' 로 끝난 그 다음 무장의 PLACE 사유(= 수용 확정인지 확인 — 리뷰 r1 #2)
    ends = collections.Counter()
    for rows in by_pos.values():
        for prev, cur in zip(rows, rows[1:]):
            if prev[5] in released and cur[5] not in released:
                ends[reason.get(cur[5], "<no attempt>")] += 1
    print("## arm right after a released arm that was itself not released, by its last PLACE reason:", sorted(ends.items()))

    # 3. 무장 행 바이트: JSON 셋 · 평탄화 증거 · 기본 열(평균)
    if arms:
        n = len(arms)
        print("## arm row avg bytes: json, flattened evidence, base columns:",
              round(sum(r[7] for r in arms) / n), round(sum(r[8] for r in arms) / n), round(sum(r[9] for r in arms) / n))
    print("## hook row avg bytes (REFUSED/CANCELLED, base columns):",
          q(f"SELECT count(*), round(avg({length_sum(BASE_COLS)})) FROM exit_events WHERE action IN ('PROPOSAL_REFUSED','PROPOSAL_CANCELLED')")[0])

    # 4. 재시도 한 번이 주문 원장에 남기는 행(지울 대상 아님 — 비교용): 표 바이트 / 행 수
    print("## order-ledger tables: rows, dbstat bytes, bytes/row")
    for t in ("intents", "decisions", "mutation_attempts", "attempt_transitions"):
        rows = q(f"SELECT count(*) FROM {t}")[0][0]
        size = q("SELECT sum(pgsize) FROM dbstat WHERE name=?", t)[0][0] or 0
        print("  ", t, rows, size, round(size / rows) if rows else 0)
    print("## transitions per exit PLACE attempt (avg):",
          q("""SELECT round(avg(n),2) FROM (SELECT count(t.id) n FROM mutation_attempts a JOIN attempt_transitions t ON t.attempt_id=a.id
               WHERE a.kind='PLACE' AND a.intent_id IN (SELECT proposed_intent_id FROM exit_events WHERE proposed_intent_id IS NOT NULL) GROUP BY a.id)""")[0][0])

    # 5. PLACE attempt 사유 — 전체(청산 · 진입 모두) 월별, 11월 dispatch_rejected
    print("## all PLACE attempts by month x reason (non-accepted only)")
    for r in q("""SELECT substr(recorded_at,1,7), reason_code, count(*) FROM mutation_attempts
                  WHERE kind='PLACE' AND reason_code NOT IN ('broker_accepted','') GROUP BY 1,2 ORDER BY 1,3 DESC"""):
        print("  ", r)
    print("## 2026-11 dispatch_rejected_by_broker PLACE attempts by day (all intents):")
    for r in q("""SELECT substr(recorded_at,1,10), count(*) FROM mutation_attempts WHERE kind='PLACE'
                  AND reason_code='dispatch_rejected_by_broker' AND recorded_at >= '2026-11' GROUP BY 1"""):
        print("  ", r)
    print("## 2026-11 opposite_pending_order_exists PLACE attempts by day (all intents):")
    for r in q("""SELECT substr(recorded_at,1,10), count(*) FROM mutation_attempts WHERE kind='PLACE'
                  AND reason_code='opposite_pending_order_exists' AND recorded_at >= '2026-11' GROUP BY 1"""):
        print("  ", r)

    # 6. 알림 outbox(키 중복 제거 · 재알림 창 a096): exit.* 유형별 행 수 · 상태
    print("## alert_outbox exit.* rows by type x state")
    for r in q("SELECT event_type, state, count(*) FROM alert_outbox WHERE event_type LIKE 'exit.%' GROUP BY 1,2 ORDER BY 1,2"):
        print("  ", r)
    print("## alert_outbox exit.proposal_refused: rows, distinct positions (by key part), max keys per position")
    keys = [r[0] for r in q("SELECT event_key FROM alert_outbox WHERE event_type='exit.proposal_refused'")]
    per = collections.Counter(k.split("|")[1] if k.count("|") >= 1 else "" for k in keys)
    print("  ", len(keys), len(per), max(per.values()) if per else 0)
    print("## alert_outbox exit.proposal_refused created_at by month:",
          q("SELECT substr(created_at,1,7), count(*) FROM alert_outbox WHERE event_type='exit.proposal_refused' GROUP BY 1"))


if __name__ == "__main__":
    main()

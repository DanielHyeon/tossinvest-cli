#!/usr/bin/env python3
"""a206 exit_events 용량 측정 하네스 — 집계만 출력, 운영 저널에는 읽기 전용으로만 닿음.

사용:
  python3 -I a206_measure.py <원본 journal.db> <작업 디렉터리>

동작:
  1. 원본을 `?mode=ro` 로 열고 SQLite backup API 로 작업 디렉터리에 사본(base.db)을 만듦.
  2. 사본에서 집계(행 수 · 바이트 · 중복 비율 · 반복 기록 원인 분포)만 출력함.
  3. 시나리오마다 base.db 를 다시 복사해 UPDATE + VACUUM 한 뒤 파일 크기를 재고 그 사본을 지움.
  4. 끝에 base.db 도 지움.

출력에는 계좌 · 종목 · 수량 · 금액 · 가격 · id 가 없음(건수 · 바이트 · 월 · 사유 코드만).
"""
import os
import shutil
import sqlite3
import sys

SNAP = "length(coalesce({p}saved_snapshot_json,''))+length(coalesce({p}recomputed_snapshot_json,''))+length(coalesce({p}effective_snapshot_json,''))"
EVIDENCE_COLS = [
    "position_generation", "policy_id", "policy_version", "policy_digest", "snapshot_id", "decision_id",
    "observation_id", "next_target", "next_protection", "observation_source", "observed_at",
    "projected_quantity", "proposal_ratio", "state_only", "suppressed_reason", "saved_snapshot_json",
    "recomputed_snapshot_json", "effective_snapshot_json", "effective_source", "arm_suppressed_reason",
]
# design D1: 닫힌 포지션(CLOSED 는 최종 — core_domain.go positions 주석, 그리고 exit_states.completed=1)의 행 중,
# 그 intent 의 attempt 가운데 broker_order_id 를 가진 것이 하나라도 있는 행(콘솔 orders 가 링크할 수 있는 행의 상위 집합)은 제외함.
CLOSED_UNLINKED = """position_id IN (SELECT p.id FROM positions p JOIN exit_states s ON s.position_id = p.id
                    WHERE p.state='CLOSED' AND s.completed=1)
  AND NOT EXISTS (SELECT 1 FROM mutation_attempts a WHERE a.intent_id = exit_events.proposed_intent_id
                   AND coalesce(a.broker_order_id,'') <> '')"""
CLOSED_ALL = "position_id IN (SELECT id FROM positions WHERE state='CLOSED')"

SCENARIOS = {
    "S0-vacuum-only": [],
    "S1-closed-unlinked-json-null": [
        f"UPDATE exit_events SET saved_snapshot_json=NULL, recomputed_snapshot_json=NULL, effective_snapshot_json=NULL WHERE {CLOSED_UNLINKED}"],
    "S2-closed-unlinked-all-evidence-null": [
        f"UPDATE exit_events SET {', '.join(c + '=NULL' for c in EVIDENCE_COLS)} WHERE {CLOSED_UNLINKED}"],
    "S3-closed-all-json-null": [
        f"UPDATE exit_events SET saved_snapshot_json=NULL, recomputed_snapshot_json=NULL, effective_snapshot_json=NULL WHERE {CLOSED_ALL}"],
    # 앞으로의 중복 저장 줄이기(b) 모의: effective=recomputed 와 saved=직전 effective 를 저장하지 않았다면.
    "S4-dedupe-simulation": [
        # saved 를 먼저 비움 — lag(effective) 가 아직 남아 있을 때 계산해야 함(리뷰 r1 #6).
        """UPDATE exit_events SET saved_snapshot_json=NULL WHERE id IN (
             SELECT id FROM (SELECT id, saved_snapshot_json s,
               lag(effective_snapshot_json) OVER (PARTITION BY position_id ORDER BY id) pe
               FROM exit_events WHERE recomputed_snapshot_json IS NOT NULL) WHERE s = pe)""",
        "UPDATE exit_events SET effective_snapshot_json=NULL WHERE effective_snapshot_json = recomputed_snapshot_json"],
}


def refuse_live(path: str) -> None:
    live = os.path.realpath(os.path.expanduser("~/.config/tossctl"))
    real = os.path.realpath(path)
    if real == live or real.startswith(live + os.sep):
        sys.exit(f"refuse: {path} is the live journal directory or under it")


def q(c, sql, *args):
    return c.execute(sql, args).fetchall()


def stats(c) -> None:
    print("## tables (dbstat, bytes)")
    for name, size in q(c, "SELECT name, sum(pgsize) FROM dbstat GROUP BY name ORDER BY 2 DESC LIMIT 8"):
        print(f"  {name}\t{size}")
    print("## exit_events by month: rows, snapshot bytes")
    for r in q(c, f"SELECT substr(created_at,1,7), count(*), sum({SNAP.format(p='')}) FROM exit_events GROUP BY 1"):
        print("  ", r)
    print("## by action: rows, snapshot bytes")
    for r in q(c, f"SELECT coalesce(action,'<judgement-only>'), count(*), sum({SNAP.format(p='')}) FROM exit_events GROUP BY 1 ORDER BY 2 DESC"):
        print("  ", r)
    print("## rows per position, top 5 (counts only); positions:", q(c, "SELECT count(DISTINCT position_id) FROM exit_events")[0][0])
    print("  ", [r[0] for r in q(c, "SELECT count(*) n FROM exit_events GROUP BY position_id ORDER BY n DESC LIMIT 5")])
    print("## days with > 500 rows: day, rows, positions")
    for r in q(c, "SELECT substr(created_at,1,10), count(*), count(DISTINCT position_id) FROM exit_events GROUP BY 1 HAVING count(*) > 500"):
        print("  ", r)
    print("## arm rows by month x PLACE attempt reason_code (loop causes)")
    for r in q(c, """SELECT substr(e.created_at,1,7), coalesce(a.reason_code,'<no attempt>'), count(*) FROM exit_events e
                     LEFT JOIN mutation_attempts a ON a.intent_id=e.proposed_intent_id AND a.kind='PLACE'
                     WHERE e.action IN ('LADDER_PARTIAL','STOP_LOSS_LADDER') GROUP BY 1,2 ORDER BY 1,3 DESC"""):
        print("  ", r)
    print("## November arm rows by day x PLACE reason_code (not broker_accepted)")
    for r in q(c, """SELECT substr(e.created_at,1,10), a.reason_code, count(*) FROM exit_events e
                     JOIN mutation_attempts a ON a.intent_id=e.proposed_intent_id AND a.kind='PLACE'
                     WHERE e.created_at >= '2026-11' AND e.action IN ('LADDER_PARTIAL','STOP_LOSS_LADDER')
                       AND a.reason_code <> 'broker_accepted' GROUP BY 1,2"""):
        print("  ", r)
    print("## duplication among evidence rows: rows, eff=recomputed, saved=prev effective, bytes(eff dup), bytes(saved dup)")
    print("  ", q(c, f"""WITH ev AS (SELECT saved_snapshot_json s, recomputed_snapshot_json r, effective_snapshot_json e,
                lag(effective_snapshot_json) OVER (PARTITION BY position_id ORDER BY id) pe
                FROM exit_events WHERE recomputed_snapshot_json IS NOT NULL)
              SELECT count(*), sum(e=r), sum(s=pe), sum(CASE WHEN e=r THEN length(e) ELSE 0 END),
                     sum(CASE WHEN s=pe THEN length(s) ELSE 0 END) FROM ev""")[0])
    print("## by position state: positions, rows, snapshot bytes")
    for r in q(c, f"""SELECT p.state, count(DISTINCT p.id), count(*), sum({SNAP.format(p='e.')})
                     FROM exit_events e JOIN positions p ON p.id=e.position_id GROUP BY 1"""):
        print("  ", r)
    print("## closed rows linked to an attempt with broker_order_id (kept by S1/S2):",
          q(c, f"SELECT count(*) FROM exit_events WHERE {CLOSED_ALL} AND NOT ({CLOSED_UNLINKED})")[0][0])
    print("## CLOSED positions whose exit_states is not completed:",
          q(c, "SELECT count(*) FROM positions p JOIN exit_states s ON s.position_id=p.id WHERE p.state='CLOSED' AND s.completed=0")[0][0])


def main() -> None:
    if len(sys.argv) != 3:
        sys.exit(__doc__)
    src_path, work = sys.argv[1], sys.argv[2]
    refuse_live(work)
    if os.path.realpath(src_path).startswith(os.path.realpath(work) + os.sep):
        sys.exit("refuse: the source must not be inside the work directory (scenario copies are overwritten and deleted)")
    os.makedirs(work, exist_ok=True)
    base = os.path.join(work, "base.db")
    src = sqlite3.connect("file:" + os.path.abspath(src_path) + "?mode=ro", uri=True)
    dst = sqlite3.connect(base)
    src.backup(dst)
    src.close()
    print("## source pragmas: user_version, page_size, page_count, freelist:",
          [dst.execute(f"PRAGMA {p}").fetchone()[0] for p in ("user_version", "page_size", "page_count", "freelist_count")])
    stats(dst)
    dst.close()
    print("## base copy bytes:", os.path.getsize(base))
    print("## scenarios: name, rows changed, bytes after VACUUM, exit_events table bytes after")
    for name, stmts in SCENARIOS.items():
        path = os.path.join(work, f"{name}.db")
        shutil.copyfile(base, path)
        c = sqlite3.connect(path)
        changed = 0
        for s in stmts:
            changed += c.execute(s).rowcount
        c.commit()
        c.execute("VACUUM")
        table = q(c, "SELECT sum(pgsize) FROM dbstat WHERE name='exit_events'")[0][0]
        c.close()
        print(f"  {name}\t{changed}\t{os.path.getsize(path)}\t{table}")
        os.remove(path)
    os.remove(base)


if __name__ == "__main__":
    main()

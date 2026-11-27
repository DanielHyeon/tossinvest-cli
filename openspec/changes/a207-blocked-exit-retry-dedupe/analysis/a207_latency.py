# a207 압축 판정 질의 두 개의 비용 실측 — 원본은 ?mode=ro 로만 열고 backup API 사본에서 재며 사본은 지움. 출력은 질의 계획 · 시간 통계뿐(id 없음).
# 사용: python3 -I a207_latency.py <원본 journal.db> <작업 디렉터리>
import os, sqlite3, sys, time, statistics
src, work = sys.argv[1], sys.argv[2]
live = os.path.realpath(os.path.expanduser("~/.config/tossctl"))
if os.path.realpath(work) == live or os.path.realpath(work).startswith(live + os.sep):
    sys.exit("refuse: the work directory is the live journal directory or under it")
base = os.path.join(work, "lat.db")
s = sqlite3.connect("file:" + os.path.abspath(src) + "?mode=ro", uri=True); c = sqlite3.connect(base); s.backup(c); s.close()
hooks = "('PROPOSAL_REFUSED','PROPOSAL_CANCELLED','PROPOSAL_FILLED','OPENED','COMPLETED','ADJUSTMENT_CLOSED')"
pos, rows = c.execute("SELECT position_id, count(*) FROM exit_events GROUP BY 1 ORDER BY 2 DESC LIMIT 1").fetchone()
q1 = f"""SELECT id, coalesce(lifecycle_generation,1), action, level_after, proposed_intent_id FROM exit_events
        WHERE position_id=? AND proposed_intent_id IS NOT NULL AND coalesce(action,'') NOT IN {hooks} ORDER BY id DESC LIMIT 1"""
q2 = "SELECT EXISTS(SELECT 1 FROM exit_events WHERE proposed_intent_id=? AND action IN ('PROPOSAL_REFUSED','PROPOSAL_CANCELLED'))"
plan = [r[-1] for r in c.execute("EXPLAIN QUERY PLAN " + q1, (pos,))] + [r[-1] for r in c.execute("EXPLAIN QUERY PLAN " + q2, ("x",))]
ts = []
for _ in range(2000):
    t = time.perf_counter(); r = c.execute(q1, (pos,)).fetchone(); c.execute(q2, (r[4],)).fetchone(); ts.append((time.perf_counter() - t) * 1e6)
ts.sort()
print("plan:", plan)
print("two queries, largest position (id hidden, %d rows), n=2000:" % rows, " p50 %.1f us, p99 %.1f us, max %.1f us" % (ts[1000], ts[1980], ts[-1]))
c.close(); os.remove(base)

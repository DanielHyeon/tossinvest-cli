#!/usr/bin/env python3
"""a206 2.5 사본 하네스 — 운영 저널을 ?mode=ro 로 열어 SQLite backup API 로 <dst_dir>/journal.db 를 만들고,
a206_reader_check.go.txt 가 읽는 scopes.json · positions.json(id 포함 — scratch 에만 두고 지움)을 <dst_dir>/.. 에 씀.
사용: python3 -I a206_copy_journal.py <운영 journal.db> <scratch>/copy  → tossctl --config-dir <scratch>/copy engine exit-evidence-compact [--apply ...]
대상 디렉터리가 ~/.config/tossctl 이거나 그 아래면, 또는 이미 journal.db 가 있으면 거절함."""
import json, os, sqlite3, sys
src, dst_dir = sys.argv[1], sys.argv[2]
live = os.path.realpath(os.path.expanduser("~/.config/tossctl"))
real = os.path.realpath(dst_dir)
if real == live or real.startswith(live + os.sep):
    sys.exit("refuse: destination is the live directory")
dst = os.path.join(dst_dir, "journal.db")
if os.path.exists(dst):
    sys.exit("refuse: destination exists")
s = sqlite3.connect("file:" + os.path.abspath(src) + "?mode=ro", uri=True)
d = sqlite3.connect(dst)
s.backup(d)
s.close()
scopes = [dict(broker_order_id=b, account_ref=a, market=m, trading_day=t) for b, a, m, t in d.execute(
    """SELECT DISTINCT a.broker_order_id, i.account_ref, i.market, i.trading_day FROM mutation_attempts a
       JOIN intents i ON i.id = a.intent_id
       WHERE a.kind='PLACE' AND a.state='CONFIRMED' AND coalesce(a.broker_order_id,'') <> ''
         AND i.id IN (SELECT proposed_intent_id FROM exit_events WHERE proposed_intent_id IS NOT NULL)""")]
positions = [dict(ID=i, State=st) for i, st in d.execute(
    "SELECT DISTINCT p.id, p.state FROM positions p JOIN exit_events e ON e.position_id = p.id")]
d.close()
json.dump(scopes, open(os.path.join(dst_dir, "..", "scopes.json"), "w"))
json.dump(positions, open(os.path.join(dst_dir, "..", "positions.json"), "w"))
print("copy bytes", os.path.getsize(dst), "scopes", len(scopes), "positions", len(positions))

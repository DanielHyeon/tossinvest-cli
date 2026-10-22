#!/usr/bin/env python3
"""prereg_eval.py 두 판의 기존 네 팔 출력 대조(a128 tasks 6.3 영수증 재현용).

같은 프로세스에서 원장·state 를 한 번 읽어 두 판의 evaluate(in-sample, out-of-sample)·calibrate_x 출력을 비교함.
프로브가 원장에 덧붙이는 중이어도 두 판이 같은 입력을 보게 하려는 것. 읽기 전용.

사용: git show <rev>:<경로>/prereg_eval.py > /tmp/old.py
      python3 -I prereg_eval_compare.py /tmp/old.py prereg_eval.py shadow-ledger/ledger.jsonl shadow-ledger/states
"""
import hashlib, importlib.util, sys
def load(name, path):
    spec = importlib.util.spec_from_file_location(name, path); m = importlib.util.module_from_spec(spec); spec.loader.exec_module(m); return m
old, new = load("old", sys.argv[1]), load("new", sys.argv[2])
ledger, states_dir = sys.argv[3], sys.argv[4]
j2, labels, sk = old.load_ledger(ledger)
j2n, labelsn, skn = new.load_ledger(ledger)
so, uo = old.load_states(states_dir); sn, un = new.load_states(states_dir)
assert set(so) == set(sn) and uo == un, ("states changed between loads", len(so), len(sn))
assert len(j2) == len(j2n) and labels == labelsn
print(f"입력: J2 {len(j2)}건(new 로더 {len(j2n)}), 라벨 {len(labels)}, state {len(so)}(읽기 실패 {uo}), 잘린 줄 {sk}")
for w in ("in-sample", "out-of-sample", "calibrate-x"):
    outs = []
    for m, st in ((old, so), (new, sn)):
        buf = []
        if w == "calibrate-x": m.calibrate_x(j2, st, out=buf.append)
        else: m.evaluate(j2, labels, st, w, out=buf.append)
        outs.append("\n".join(buf).encode())
    h = [hashlib.sha256(o).hexdigest()[:16] for o in outs]
    print(f"{w}: old {h[0]} ({len(outs[0])}B) new {h[1]} ({len(outs[1])}B) -> {'IDENTICAL' if outs[0]==outs[1] else 'DIFFERENT'}")

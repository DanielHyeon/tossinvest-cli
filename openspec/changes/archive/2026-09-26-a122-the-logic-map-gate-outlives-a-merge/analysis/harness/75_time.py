"""task 7.5 MEASURE — `compute_landing` 의 **실제** 벽시계와 spawn 수.

census 의 spawn 수는 예측이다(조기 종료를 모른다). 여기서는 `subprocess.run` 을 감싸서
**실제로 몇 번 돌았는지**와 시간을 같이 잰다. 인자는 change 디렉터리 이름들이다.
"""
import importlib.util
import subprocess
import sys
import time
from pathlib import Path

# 루트는 세어서 올라가지 않고 **유도**한다 — 이 파일이 사는 깊이는 change 경로가
# 바뀌면 같이 바뀌고, 체크아웃 이름이 바뀌면 절대경로는 고아가 된다
# ([[renamed-checkout-strands-absolute-path-state]]).
ROOT = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
SP = Path(__file__).resolve().parent / "_work"
SP.mkdir(exist_ok=True)
sys.path.insert(0, str(ROOT / "tools" / "logic-map"))
spec = importlib.util.spec_from_file_location("ca75t", ROOT / "tools" / "logic-map" / "check_analysis.py")
ca = importlib.util.module_from_spec(spec)
sys.modules["ca75t"] = ca
spec.loader.exec_module(ca)

REAL = subprocess.run
COUNT = {"n": 0, "argv0": {}}


def counting(*args, **kwargs):
    COUNT["n"] += 1
    argv = args[0] if args else kwargs.get("args", [])
    key = " ".join(str(a) for a in argv[:3]) if isinstance(argv, list) else str(argv)[:40]
    COUNT["argv0"][key] = COUNT["argv0"].get(key, 0) + 1
    return REAL(*args, **kwargs)


ca.subprocess.run = counting
changes = ROOT / "openspec" / "changes"

for name in sys.argv[1:]:
    d = changes / "archive" / name if (changes / "archive" / name).is_dir() else changes / name
    m = ca.ARCHIVED_CHANGE.fullmatch(name)
    cid = m.group("change") if m else name
    base = ca.resolve_base(d, ROOT, {}, change_id=cid, head=ca._head_commit(ROOT))
    analysis = d / "analysis" / "function-logic"
    COUNT["n"] = 0
    COUNT["argv0"] = {}
    start = time.monotonic()
    try:
        # 7.5.2.1 부터 입력 한 벌을 호출자가 잰다 — 그 잼(하한 · 수리 신호)까지 시간에 넣는다(옛 판본의
        # `compute_landing(…, analysis)` 가 안에서 재던 것과 같은 범위).
        inputs = ca._measure_landing_inputs(ROOT, ca._head_commit(ROOT), ca._read_evidence(analysis))
        value, why = ca.compute_landing(ROOT, base, inputs)
    except ca.GATE_FAULTS as exc:      # 7.5.1 부터 git 결함은 예외다 — 한 change 가 나머지를 멈추지 않게
        value, why = "", f"RAISED {type(exc).__name__}: {exc}"
    elapsed = time.monotonic() - start
    print(f"{name[:46]:48s} {elapsed:8.2f}s · spawn {COUNT['n']:7d} · "
          f"{(value[:12] if value else 'none')}")
    for key, n in sorted(COUNT["argv0"].items(), key=lambda kv: -kv[1])[:4]:
        print(f"{'':50s} {n:7d} × {key}")
    if why:
        print(f"{'':50s} 사유 {why[:100]}")

"""task 7.5 MEASURE — 오늘 HEAD 에서 walk 가 무엇을 얼마나 쓰는가.

리뷰가 적은 133.7s·219.6s 는 **리뷰 시점**의 수다. 그 뒤 7.2.1·7.2.2·7.2.4·7.2.6·7.6·7.7 이
같은 경로를 고쳤으므로 수리 직전에 다시 센다([[caller-count-is-not-fix-site-count]]).

여기서는 **걷지 않는다** — 걸으면 change 하나에 수백 초다. 대신 걷기 전에 정해지는
(후보 수 · 번들 수)를 전수로 재고, 그 둘로 spawn 수를 예측한 뒤 실제 시간은 따로 잰다.
"""
import importlib.util
import json
import subprocess
import sys
from pathlib import Path

# 루트는 세어서 올라가지 않고 **유도**한다 — 이 파일이 사는 깊이는 change 경로가
# 바뀌면 같이 바뀌고, 체크아웃 이름이 바뀌면 절대경로는 고아가 된다
# ([[renamed-checkout-strands-absolute-path-state]]).
ROOT = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
SP = Path(__file__).resolve().parent / "_work"
SP.mkdir(exist_ok=True)
sys.path.insert(0, str(ROOT / "tools" / "logic-map"))
spec = importlib.util.spec_from_file_location("ca75", ROOT / "tools" / "logic-map" / "check_analysis.py")
ca = importlib.util.module_from_spec(spec)
sys.modules["ca75"] = ca
spec.loader.exec_module(ca)

changes = ROOT / "openspec" / "changes"
names = sorted(p.name for p in changes.iterdir() if p.is_dir() and p.name != "archive")
names += sorted(p.name for p in (changes / "archive").iterdir() if p.is_dir())

rows = []
for name in names:
    m = ca.ARCHIVED_CHANGE.fullmatch(name)
    cid = m.group("change") if m else name
    d = changes / "archive" / name if (changes / "archive" / name).is_dir() else changes / name
    try:
        base = ca.resolve_base(d, ROOT, {}, change_id=cid, head=ca._head_commit(ROOT))
    except Exception as exc:                      # 재는 스크립트다 — 못 재면 사유를 적는다
        rows.append({"change": name, "skip": f"base: {type(exc).__name__}: {exc}"[:120]})
        continue
    analysis = d / "analysis" / "function-logic"
    try:
        # 7.5.2.1 부터 하한은 호출자가 한 번 푼 `HEAD` 와 한 번 읽은 증거로 잰다 — 도구가 쓰는 것과 같은
        # 한 벌을 여기서도 만든다. (7.5 · 7.5.2 가 `_walk_floor` 의 반환을 바꿀 때마다 이 줄이 깨졌다 —
        # 7.5.2 재리뷰가 셋을 푸는 판본이 126 건 전부를 "못 걸음" 으로 찍는 것을 짚었다.)
        head = ca._head_commit(ROOT)
        pinned = ca._select_pinning(ROOT, ca._read_evidence(analysis))
        bundles = len(pinned)
        floor, why = ca._walk_floor(ROOT, pinned, head)
    except Exception as exc:
        rows.append({"change": name, "skip": f"floor: {type(exc).__name__}: {exc}"[:120]})
        continue
    if why:
        rows.append({"change": name, "bundles": bundles, "skip": "no walk: " + why[:60]})
        continue
    start = floor if ca._is_ancestor(ROOT, base, floor) else base
    walked = subprocess.run(["git", "rev-list", "--count", f"{start}..{head}"],
                            cwd=ROOT, capture_output=True, text=True, timeout=120, check=False)
    candidates = int(walked.stdout.strip() or 0) + 1 if not walked.returncode else -1
    rows.append({"change": name, "bundles": bundles, "candidates": candidates,
                 "floor": floor[:12], "base": base[:12],
                 # 거절이 `_pinning_at` 에서 끝나는 후보의 비용 = merge-base 1 + show N
                 "min_spawn": candidates * (bundles + 1),
                 # 마지막 가드까지 가는 후보의 비용 = 1 + N + 1 + N + 2N + 1
                 "max_spawn": candidates * (4 * bundles + 3)})

walkers = [r for r in rows if "candidates" in r]
print(f"changes {len(rows)} · 걷는 것 {len(walkers)} · 안 걷는 것 {len(rows) - len(walkers)}")
print(f"후보 합계 {sum(r['candidates'] for r in walkers)} · "
      f"최소 spawn 합계 {sum(r['min_spawn'] for r in walkers)} · "
      f"최대 spawn 합계 {sum(r['max_spawn'] for r in walkers)}")
print("\n비싼 순 12:")
for r in sorted(walkers, key=lambda r: -r["min_spawn"])[:12]:
    print(f"  {r['change'][:44]:46s} 후보 {r['candidates']:5d} · 번들 {r['bundles']:4d} "
          f"· spawn {r['min_spawn']:7d}~{r['max_spawn']:7d}")
json.dump(rows, open(SP / "75_census.json", "w"), indent=1, ensure_ascii=False)

"""task 7.5 VERIFY — 판정 A/B 전수. 성능 수리가 **값이나 사유를 바꾸면** 그것은 성능 수리가 아니다.

사본 대상 + 시작 sha 단언 ([[mutation-revert-needs-the-right-baseline]]). 편집 전 판본은
워킹트리를 되돌려서가 아니라 HEAD blob 에서 꺼낸다.
"""
import importlib.util
import json
import shutil
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
# 비교 기준은 **인자**다. `HEAD` 로 굳혀 두면 수리가 랜딩한 다음 날 대조군이 오염된다
# ([[a-recorded-boundary-stops-being-rechecked]] — 기록한 경계는 다시 대조되지 않는다).
# 기본값은 7.5 직전 커밋이다.
BEFORE_DEFAULT = "8091e6c4d4b6c7685b2e02cdc14cab932d41874a"
BEFORE = next((a for a in sys.argv[1:] if len(a) >= 7 and not a.replace(".", "").isdigit()),
              BEFORE_DEFAULT)


def build(tag: str, revision: str | None) -> object:
    where = SP / f"75_ab_{tag}"
    if where.exists():
        shutil.rmtree(where)
    shutil.copytree(ROOT / "tools" / "logic-map", where,
                    ignore=shutil.ignore_patterns("__pycache__", "test_*.py"))
    if revision:                                  # 편집 전 판본은 blob 에서 꺼낸다
        blob = subprocess.run(["git", "show", f"{revision}:tools/logic-map/check_analysis.py"],
                              cwd=ROOT, capture_output=True, check=True).stdout
        (where / "check_analysis.py").write_bytes(blob)
    sys.path.insert(0, str(where))
    spec = importlib.util.spec_from_file_location(f"ca_{tag}", where / "check_analysis.py")
    module = importlib.util.module_from_spec(spec)
    sys.modules[f"ca_{tag}"] = module
    spec.loader.exec_module(module)
    sys.path.pop(0)
    return module


before = build("before", BEFORE)
after = build("after", None)
# 대조군이 대조군인지는 **이름이 아니라 소스**로 확인한다. 이름으로 보면 기준을 중간
# 커밋으로 잡는 정당한 사용(추이적 A/B)을 막고, 정작 같은 소스를 두 번 재는 오염은 못 잡는다.
_before_src = (SP / "75_ab_before" / "check_analysis.py").read_bytes()
_after_src = (SP / "75_ab_after" / "check_analysis.py").read_bytes()
assert _before_src != _after_src, (
    f"{BEFORE[:12]} 사본이 지금 워킹트리와 **같다** — 대조군이 오염됐다. 비교하려는 변경 "
    "**직전** 커밋을 첫 인자로 줄 것 (시간 예산은 숫자로 준다)")
print(f"대조군 OK · 비교 기준 blob = {BEFORE[:12]}")

changes = ROOT / "openspec" / "changes"
names = sorted(p.name for p in changes.iterdir() if p.is_dir() and p.name != "archive")
names += sorted(p.name for p in (changes / "archive").iterdir() if p.is_dir())

# 파일 이름에 baseline 을 넣는다 — 안 넣으면 다른 `BEFORE` 로 한 번 더 돌릴 때 두 baseline 의
# 결과가 **조용히 섞인다** (2026-09-18 독립 리뷰 P2).
DONE = SP / f"75_ab_done.{BEFORE[:12]}.json"
done = json.load(open(DONE)) if DONE.exists() else {}
budget = next((float(a) for a in sys.argv[1:] if a.replace(".", "").isdigit()), 480.0)
spent = 0.0

same = differ = skipped = 0
rows, slowest = [], []
for name in names:
    if name in done:                              # 이어 달린다 — 한 번에 다 못 돈다
        row = done[name]
        if row.get("skip"):
            skipped += 1
            continue
        same += row["same"]
        differ += not row["same"]
        slowest.append((row["before_s"], row["after_s"], name))
        if not row["same"]:
            rows.append(row)
        continue
    if spent > budget:
        continue
    d = changes / "archive" / name if (changes / "archive" / name).is_dir() else changes / name
    m = after.ARCHIVED_CHANGE.fullmatch(name)
    cid = m.group("change") if m else name
    analysis = d / "analysis" / "function-logic"
    try:
        base = after.resolve_base(d, ROOT, {}, change_id=cid)
    except Exception:
        skipped += 1
        done[name] = {"change": name, "skip": "base"}
        json.dump(done, open(DONE, "w"), indent=1, ensure_ascii=False)
        continue
    out = {}
    for tag, module in (("before", before), ("after", module_after := after)):
        start = time.monotonic()
        try:
            out[tag] = (module.compute_landing(ROOT, base, analysis), "")
        except Exception as exc:                  # 예외도 판정이다 — 타입과 문장을 비교한다
            out[tag] = (None, f"{type(exc).__name__}: {exc}")
        out[tag + "_s"] = time.monotonic() - start
    ok = out["before"] == out["after"]
    same += ok
    differ += not ok
    spent += out["before_s"] + out["after_s"]
    slowest.append((out["before_s"], out["after_s"], name))
    done[name] = {"change": name, "same": bool(ok), "before_s": out["before_s"],
                  "after_s": out["after_s"], "before": str(out["before"])[:300],
                  "after": str(out["after"])[:300]}
    json.dump(done, open(DONE, "w"), indent=1, ensure_ascii=False)
    if not ok:
        rows.append({"change": name, "before": str(out["before"])[:200],
                     "after": str(out["after"])[:200]})
        print(f"  DIFFERENT {name}\n    before {out['before']}\n    after  {out['after']}")

print(f"\n전수 {len(names)} · 남은 것 {len(names) - len(done)} · 비교 {same + differ} · **SAME {same} · DIFFERENT {differ}** · base 없음 {skipped}")
slowest.sort(key=lambda r: -r[0])
print("가장 느렸던 8 (before → after):")
for b, a, name in slowest[:8]:
    print(f"  {name[:46]:48s} {b:8.2f}s → {a:6.2f}s  ({b / a:5.1f}×)" if a > 0.004
          else f"  {name[:46]:48s} {b:8.2f}s → {a:6.2f}s")
print(f"합계 {sum(r[0] for r in slowest):.1f}s → {sum(r[1] for r in slowest):.1f}s")
json.dump(rows, open(SP / "75_ab_diff.json", "w"), indent=1, ensure_ascii=False)

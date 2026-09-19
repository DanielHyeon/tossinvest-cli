"""task 7.5.1 VERIFY — `check()` **전체**의 판정 A/B. 이어 달리기 · 기준 소스 대조군.

7.5 의 A/B 는 `compute_landing` 만 덮었다 (2026-09-19 gstack 리뷰 — 서브에이전트가 "문장보다
좁다" 고 짚었다). 7.5.1 은 `check()` 안의 미리 읽기 · 한 벌 측정 · 예외 목록을 바꾸므로
판정 경로 **전체**의 출력(오류 줄 목록)을 비교한다.

    python3 751_check_ab.py [<before-sha>] [<초 예산>]
"""
import importlib.util
import json
import shutil
import subprocess
import sys
import time
from pathlib import Path

# 루트는 **유도**한다 ([[renamed-checkout-strands-absolute-path-state]]).
ROOT = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
SP = Path(__file__).resolve().parent / "_work"
SP.mkdir(exist_ok=True)
BEFORE = next((a for a in sys.argv[1:] if len(a) >= 7 and not a.replace(".", "").isdigit()),
              "b29e1f4e15ee2ab2e8d2fc9a3e9cf2417c649f9e")
budget = next((float(a) for a in sys.argv[1:] if a.replace(".", "").isdigit()), 480.0)


def build(tag: str, revision: str | None):
    where = SP / f"751_{tag}"
    if where.exists():
        shutil.rmtree(where)
    shutil.copytree(ROOT / "tools" / "logic-map", where,
                    ignore=shutil.ignore_patterns("__pycache__", "test_*.py"))
    if revision:
        blob = subprocess.run(["git", "show", f"{revision}:tools/logic-map/check_analysis.py"],
                              cwd=ROOT, capture_output=True, check=True).stdout
        (where / "check_analysis.py").write_bytes(blob)
    sys.path.insert(0, str(where))
    spec = importlib.util.spec_from_file_location(f"ca751_{tag}", where / "check_analysis.py")
    module = importlib.util.module_from_spec(spec)
    sys.modules[f"ca751_{tag}"] = module
    spec.loader.exec_module(module)
    sys.path.pop(0)
    return module, (where / "check_analysis.py").read_bytes()


before, before_src = build("before", BEFORE)
after, after_src = build("after", None)
# 대조군이 대조군인지는 **소스**로 확인한다 — 같으면 A/B 가 공허하다.
assert before_src != after_src, f"{BEFORE[:12]} 사본이 워킹트리와 같다 — 대조군 오염"
print(f"대조군 OK · 비교 기준 blob = {BEFORE[:12]}")

changes = ROOT / "openspec" / "changes"
ids = sorted(p.name for p in changes.iterdir() if p.is_dir() and p.name != "archive")
ids += sorted(after.ARCHIVED_CHANGE.fullmatch(p.name).group("change")
              for p in (changes / "archive").iterdir()
              if p.is_dir() and after.ARCHIVED_CHANGE.fullmatch(p.name))

DONE = SP / f"751_check_ab_done.{BEFORE[:12]}.json"
done = json.load(open(DONE)) if DONE.exists() else {}
spent = 0.0
for cid in ids:
    if cid in done or spent > budget:
        continue
    out = {}
    for tag, module in (("before", before), ("after", after)):
        start = time.monotonic()
        try:
            out[tag] = sorted(module.check(cid, ROOT, {}))
        except Exception as exc:                  # 예외도 판정이다 — 타입과 문장까지 비교한다
            out[tag] = [f"RAISED {type(exc).__name__}: {exc}"]
        out[tag + "_s"] = time.monotonic() - start
    spent += out["before_s"] + out["after_s"]
    done[cid] = {"same": out["before"] == out["after"], "before": out["before"][:6],
                 "after": out["after"][:6], "before_s": out["before_s"], "after_s": out["after_s"]}
    json.dump(done, open(DONE, "w"), indent=1, ensure_ascii=False)
    if not done[cid]["same"]:
        print(f"  DIFFERENT {cid}\n    before {out['before'][:3]}\n    after  {out['after'][:3]}")

same = sum(row["same"] for row in done.values())
print(f"\n전수 {len(ids)} · 남은 것 {len(ids) - len(done)} · 비교 {len(done)} · "
      f"**SAME {same} · DIFFERENT {len(done) - same}**")
print(f"합계 {sum(r['before_s'] for r in done.values()):.1f}s → "
      f"{sum(r['after_s'] for r in done.values()):.1f}s")

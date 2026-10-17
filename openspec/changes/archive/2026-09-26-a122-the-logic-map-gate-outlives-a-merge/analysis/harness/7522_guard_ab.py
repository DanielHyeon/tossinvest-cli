"""task 7.5.22 — `_safe_changed_go_paths` 하나만 A/B. 형식만 바꿨다는 것을 답으로 증명한다.

`--name-only` 판본(기준 blob)과 `--numstat` 판본(워킹트리)을 `base-commit.txt` 를 가진 change 전부에
대해 target 둘(워킹트리 · HEAD)로 돌려 **올라오는 것까지 글자 그대로** 비교한다.

    python3 7522_guard_ab.py
"""
import importlib.util, shutil, subprocess, sys, tempfile
from pathlib import Path

ROOT = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
BEFORE = "197a0355"
tmp = Path(tempfile.mkdtemp())

def load(tag, revision):
    where = tmp / tag
    shutil.copytree(ROOT / "tools" / "logic-map", where,
                    ignore=shutil.ignore_patterns("__pycache__", "test_*.py"))
    if revision:
        blob = subprocess.run(["git", "show", f"{revision}:tools/logic-map/check_analysis.py"],
                              cwd=ROOT, capture_output=True, check=True).stdout
        (where / "check_analysis.py").write_bytes(blob)
    sys.path.insert(0, str(where))
    spec = importlib.util.spec_from_file_location(f"ca_{tag}", where / "check_analysis.py")
    mod = importlib.util.module_from_spec(spec)
    sys.modules[f"ca_{tag}"] = mod
    spec.loader.exec_module(mod)
    return mod

old, new = load("before", BEFORE), load("after", None)

def answer(mod, base, target):
    try:
        mod._safe_changed_go_paths(ROOT, base, target)
        return "OK"
    except Exception as e:
        return f"{type(e).__name__}: {e}"

same = diff = 0
for bc in sorted((ROOT / "openspec" / "changes").rglob("base-commit.txt")):
    base = bc.read_text(encoding="utf-8").strip()
    for target in ("", "HEAD"):
        a, b = answer(old, base, target), answer(new, base, target)
        if a == b:
            same += 1
        else:
            diff += 1
            print(f"DIFFERENT {bc.parent.name} target={target!r}\n  before {a}\n  after  {b}")
print(f"\nSAME {same} · DIFFERENT {diff}")
shutil.rmtree(tmp)

"""task 7.5.22 — `.gitattributes` 한 줄이 **판정 줄 전체**에 무엇을 하는지 CLI 수준으로 잰다.

`git worktree add --detach <PROBE> HEAD` 로 **격리 worktree** 를 만들고 그 안에만 `.gitattributes` 를
놓는다 — 공유 워크트리(병행 세션이 같이 쓴다)를 건드리지 않는다. 편집 전/후 두 판본을 같은 입력에 돌려
표를 만든다.

    git worktree add --detach <PROBE> HEAD
    python3 7522_switch.py
    git worktree remove --force <PROBE>
"""
import importlib.util, shutil, subprocess, sys
from pathlib import Path
ROOT = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
PROBE = Path(sys.argv[1]) if len(sys.argv) > 1 else Path("/tmp/a122-probe")
CHANGE = "a112-run-four-strategy-families-independently"

def load(tag, revision):
    where = PROBE.parent / f"ca_{tag}"
    if where.exists(): shutil.rmtree(where)
    shutil.copytree(ROOT / "tools" / "logic-map", where,
                    ignore=shutil.ignore_patterns("__pycache__", "test_*.py"))
    if revision:
        blob = subprocess.run(["git", "show", f"{revision}:tools/logic-map/check_analysis.py"],
                              cwd=ROOT, capture_output=True, check=True).stdout
        (where / "check_analysis.py").write_bytes(blob)
    sys.path.insert(0, str(where))
    spec = importlib.util.spec_from_file_location(f"ca_{tag}", where / "check_analysis.py")
    m = importlib.util.module_from_spec(spec); sys.modules[f"ca_{tag}"] = m; spec.loader.exec_module(m)
    return m

old, new = load("before", "197a0355"), load("after", None)
attrs = PROBE / ".gitattributes"

for label, put in (("(1) 평소", False), ("(2) `.gitattributes` 한 줄 — 추적 안 함", True)):
    if put: attrs.write_text("*.go binary\n", encoding="utf-8")
    elif attrs.exists(): attrs.unlink()
    print(f"\n=== {label} ===")
    for tag, mod in (("편집 전", old), ("편집 후", new)):
        facts = {}
        try:
            errors = mod.check(CHANGE, PROBE, facts)
        except Exception as e:
            print(f"  {tag}: 예외 {type(e).__name__}: {e}"); continue
        first = errors[0][:95] if errors else "(판정 줄 0 — evidence complete)"
        print(f"  {tag}: 판정 줄 {len(errors):2d} · required {facts.get('required_count')} · {first}")
if attrs.exists(): attrs.unlink()

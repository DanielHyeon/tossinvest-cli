#!/usr/bin/env python3
"""이 로트(task 7.5.9 · 7.5.10 · 7.5.11 · 7.5.15 · 7.5.16)의 `main()` A/B 에서 DIFFERENT 로 찍힌 change 를 **전체 출력**으로 다시 본다.

`7521_main_ab.py` 는 DIFFERENT 의 앞 여섯 줄만 남긴다. 7.5.9 는 거절 문장 하나(`… anywhere in the tree` → `… in any tracked file`)를
바꿨으므로 그 문장이 든 change 는 전부 DIFFERENT 로 찍힌다. 여기서는 기준 리비전의 코드와 워킹트리의 코드를 각각 CLI 로 돌려
**그 꼬리 하나만** 정규화한 뒤 줄 단위로 견준다 — 정규화 뒤에도 다르면 그 줄을 찍는다(판정이 바뀐 것).

    python3 759_ab_verify.py <before-rev> <done.json>
"""
import json
import subprocess
import sys
import tempfile
from pathlib import Path

REPO = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
MODULES = ("check_analysis.py", "execution_baseline.py", "role_check.py")
OLD, NEW = "which is not a Go test function anywhere in the tree", "which is not a Go test function in any tracked file"


def run(script: Path, change: str) -> list[str]:
    process = subprocess.run([sys.executable, str(script), "--change", change, "--root", str(REPO)],
                             capture_output=True, text=True, timeout=3600)
    return (process.stdout + process.stderr).splitlines() + [f"rc={process.returncode}"]


def main() -> None:
    before_rev, done_path = sys.argv[1], Path(sys.argv[2])
    done = json.loads(done_path.read_text(encoding="utf-8"))
    different = sorted(change for change, row in done.items() if not row["same"])
    with tempfile.TemporaryDirectory() as raw:
        before = Path(raw)
        for name in MODULES:
            (before / name).write_bytes(subprocess.run(["git", "show", f"{before_rev}:tools/logic-map/{name}"],
                                                       cwd=REPO, capture_output=True, check=True).stdout)
        verdict_changed = 0
        for change in different:
            old = [line.replace(OLD, NEW) for line in run(before / "check_analysis.py", change)]
            new = run(REPO / "tools" / "logic-map" / "check_analysis.py", change)
            if old == new:
                print(f"SAME-AFTER-WORDING {change} ({len(new)} lines)")
                continue
            verdict_changed += 1
            print(f"VERDICT-CHANGED {change}")
            for a, b in zip(old, new):
                if a != b:
                    print(f"  - {a[:300]}\n  + {b[:300]}")
            if len(old) != len(new):
                print(f"  lines {len(old)} → {len(new)}")
        print(f"\nDIFFERENT {len(different)} · wording-only {len(different) - verdict_changed} · verdict changed {verdict_changed}")


if __name__ == "__main__":
    main()

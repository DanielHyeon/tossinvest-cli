#!/usr/bin/env python3
"""task 7.5.29 A/B — 자기 수리 신호(`_self_repair_commits`)가 저장소의 **모든** change 디렉터리에서 같은 커밋 목록을 내는가.

편집은 `--name-only` 출력을 줄로 자르던 것을 `-z` 로 바꿨다. 기준 리비전의 `check_analysis.py`(아래 정정)과 워킹트리를 각자 불러 활성 + 아카이브의
change 디렉터리 전부에서 `(root, <dir>/analysis/function-logic, HEAD)` 로 부르고 목록(또는 예외 타입 + 문장)을 견준다.

    python3 7529_repairs_ab.py [<before-rev>]     # 기본 26e5bb3f

**모듈 하나만 기준 리비전이다 (2026-09-27 정정 — 독립 적대 리뷰).** 이 스크립트는 워킹트리의 `check_analysis` 를 먼저 import 하고, 옛 판을 같은 프로세스에서 불러서 옛 판의 `import execution_baseline` · `from role_check import …` 는 **이미 불린 워킹트리 모듈**을 받는다(`sys.modules`). 기준 리비전의 파일 셋을 임시 디렉터리에 쓰지만 쓰이는 것은 `check_analysis.py` 하나다. 세 모듈을 다 그 리비전으로 돌리는 것은 자식 프로세스를 쓰는 `759_ab_verify.py` · `7516_order.py` 다.
"""
import importlib.util
import subprocess
import sys
import tempfile
from pathlib import Path

REPO = next(parent for parent in Path(__file__).resolve().parents if (parent / "tools" / "logic-map").is_dir())
sys.path.insert(0, str(REPO / "tools" / "logic-map"))
import check_analysis as new  # noqa: E402


def main() -> None:
    before = sys.argv[1] if len(sys.argv) > 1 else "26e5bb3f825e148516ca5c15b1e035f511044b9b"
    with tempfile.TemporaryDirectory() as raw:
        for name in ("check_analysis.py", "execution_baseline.py", "role_check.py"):
            (Path(raw) / name).write_bytes(subprocess.run(["git", "show", f"{before}:tools/logic-map/{name}"], cwd=REPO,
                                                          capture_output=True, check=True).stdout)
        sys.path.insert(0, raw)
        spec = importlib.util.spec_from_file_location("ca_before_7529", Path(raw) / "check_analysis.py")
        old = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(old)
        head = new._head_commit(REPO)
        changes = REPO / "openspec" / "changes"
        dirs = sorted([p for p in changes.iterdir() if p.is_dir() and p.name != "archive"]
                      + [p for p in (changes / "archive").iterdir() if p.is_dir()])

        def answer(module, directory):
            try:
                return ("ok", module._self_repair_commits(REPO, directory / "analysis" / "function-logic", head))
            except Exception as exc:  # 예외도 답이다
                return (type(exc).__name__, str(exc))

        rows = [(d.name, answer(old, d), answer(new, d)) for d in dirs]
        different = [row for row in rows if row[1] != row[2]]
        flags = sum(len(row[2][1]) for row in rows if row[2][0] == "ok")
        print(f"before {before[:12]} · HEAD {head[:12]} · change 디렉터리 {len(rows)} · 깃발 합계(after) {flags} · "
              f"SAME {len(rows) - len(different)} · DIFFERENT {len(different)}")
        for row in different:
            print(row)


if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""task 7.5.9~7.5.16 보수 A/B — 아카이브를 "이름 먼저" 로 고른 해소기가 저장소의 **모든** change 디렉터리에서 같은 답을 내는가.

`resolve_referenced_change` 는 모든 판정의 첫 입력이다(보수 P1-2 가 바꿨다). 기준 리비전의 세 모듈과 워킹트리를 각자 불러
활성 + 아카이브의 change id 전부를 해소하고 답(경로 또는 예외 타입 + 문장)을 견준다.

    python3 759r_resolver_ab.py [<before-rev>]     # 기본 1d1e5ca7
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
    before = sys.argv[1] if len(sys.argv) > 1 else "1d1e5ca781d7716e1d5c2afbeb5a127ceb953425"
    with tempfile.TemporaryDirectory() as raw:
        for name in ("check_analysis.py", "execution_baseline.py", "role_check.py"):
            (Path(raw) / name).write_bytes(subprocess.run(["git", "show", f"{before}:tools/logic-map/{name}"], cwd=REPO,
                                                          capture_output=True, check=True).stdout)
        sys.path.insert(0, raw)
        spec = importlib.util.spec_from_file_location("ca_before_759r", Path(raw) / "check_analysis.py")
        old = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(old)
        changes = REPO / "openspec" / "changes"
        ids = sorted({p.name for p in changes.iterdir() if p.is_dir() and p.name != "archive"}
                     | {new._archived_change_id(p.name) or p.name for p in (changes / "archive").iterdir()})

        def answer(module, change):
            try:
                return ("ok", str(module.resolve_referenced_change(REPO, change)))
            except Exception as exc:  # 예외도 답이다
                return (type(exc).__name__, str(exc))

        different = [(c, answer(old, c), answer(new, c)) for c in ids if answer(old, c) != answer(new, c)]
        head = subprocess.run(["git", "rev-parse", "--short=12", "HEAD"], cwd=REPO, capture_output=True, text=True).stdout.strip()
        print(f"before {before[:12]} · HEAD {head} · ids {len(ids)} · SAME {len(ids) - len(different)} · DIFFERENT {len(different)}")
        for row in different:
            print(row)


if __name__ == "__main__":
    main()

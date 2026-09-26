#!/usr/bin/env python3
"""task 7.5.14 A/B — 기록 명령의 걷기 전 거절(`_recording_refusal`)이 저장소의 **모든** change 디렉터리에서 같은 답을 내는가.

기준 리비전의 `check_analysis.py`(아래 정정)과 워킹트리를 각자 불러, 활성 + 아카이브 전부에서 증거를 읽고 `_recording_refusal` 을 부른다. 답(거절 문장 · base,
또는 예외 타입 + 문장)을 견준다. 조언 줄(`main`)도 같은 함수를 부른다.

    python3 7514_refusal_ab.py [<before-rev>]     # 기본 HEAD

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
    before = sys.argv[1] if len(sys.argv) > 1 else "HEAD"
    with tempfile.TemporaryDirectory() as raw:
        for name in ("check_analysis.py", "execution_baseline.py", "role_check.py"):
            (Path(raw) / name).write_bytes(subprocess.run(["git", "show", f"{before}:tools/logic-map/{name}"], cwd=REPO,
                                                          capture_output=True, check=True).stdout)
        sys.path.insert(0, raw)
        spec = importlib.util.spec_from_file_location("ca_before_7514", Path(raw) / "check_analysis.py")
        old = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(old)
        head = new._head_commit(REPO)
        changes = REPO / "openspec" / "changes"
        dirs = sorted([p for p in changes.iterdir() if p.is_dir() and p.name != "archive"]
                      + [p for p in (changes / "archive").iterdir() if p.is_dir()])

        def answer(module, directory):
            change = new._archived_change_id(directory.name) or directory.name
            try:
                evidence = module._read_evidence(directory / "analysis" / "function-logic")
                return module._recording_refusal(change, directory, REPO, head, evidence)
            except Exception as exc:  # 예외도 답이다
                return (type(exc).__name__, str(exc))

        different = [(d.name, answer(old, d), answer(new, d)) for d in dirs if answer(old, d) != answer(new, d)]
        print(f"before {before} · HEAD {head[:12]} · change 디렉터리 {len(dirs)} · SAME {len(dirs) - len(different)} · "
              f"DIFFERENT {len(different)}")
        for row in different:
            print(row)


if __name__ == "__main__":
    main()

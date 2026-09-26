#!/usr/bin/env python3
"""task 7.5.16 영수증 — 기록 명령이 "걷는 동안 움직인 입력" 과 "그 편집이 부른 거절" 중 무엇을 먼저 말하는가.

모양 하나를 세 리비전의 코드로 돌린다: `_own_work_fixture`(P → W → E) 에서 `--record-landing` 이 후보 순회
(`rev-list --reverse`)를 시작하는 순간 `base-commit.txt` 를 다른 커밋 id 로 고친다. 그 파일은 기록 경로의 원장에서
증거 밖의 **유일한** 읽기이고(증거의 움직임은 `compute_landing` 이 먼저 잡는다), 추적 파일이라 트리도 더러워진다.

- `5a54f78d` (6.4 이전): 리뷰가 적은 모양 — 거절 집합의 **더러운 트리**가 먼저 나온다고 기대한다.
- `1d1e5ca7` (편집 전): 6.4(b) 의 base 대조가 거절 집합에서 더러운 트리보다 앞에 선다.
- `worktree` (편집 뒤): 원장이 거절보다 먼저 — `base-commit.txt changed while this change was being judged`.

시험 헬퍼(`_own_work_fixture` · `_on_walk`)는 워킹트리의 시험 파일에서 쓰고, 판정 코드만 리비전마다 바꾼다.
스크래치 사본에서 자식 프로세스로 돈다 — 저장소는 건드리지 않는다.

    python3 7516_order.py [<rev> ...]     # 기본: 5a54f78d 1d1e5ca7 worktree
"""
import os
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

REPO = next(parent for parent in Path(__file__).resolve().parents
            if (parent / "tools" / "logic-map").is_dir())
MODULES = ("check_analysis.py", "execution_baseline.py", "role_check.py")
PROBE = """
import tempfile, test_check_analysis as t, check_analysis as ca
raw = tempfile.TemporaryDirectory()
with raw:
    root, marks = t._own_work_fixture(raw)
    base_file = root / "openspec" / "changes" / "mine" / "base-commit.txt"
    with t._on_walk(lambda: base_file.write_text(marks["W"] + "\\n")) as fired:
        code, lines = ca.record_landing("mine", root)
    written = (root / "openspec" / "changes" / "mine" / ca.LANDING_FILE).exists()
print(f"reached={bool(fired)} rc={code} written={written}")
for line in lines:
    print("  " + line)
"""


def main() -> None:
    revisions = sys.argv[1:] or ["5a54f78d", "1d1e5ca7", "worktree"]
    for revision in revisions:
        with tempfile.TemporaryDirectory(dir=os.environ.get("A122_HARNESS_WORK")) as raw:
            work = Path(raw)
            shutil.copytree(REPO / "tools" / "logic-map", work / "logic-map",
                            ignore=shutil.ignore_patterns("__pycache__"))
            (work / "sdd").mkdir()
            shutil.copy(REPO / "tools" / "sdd" / "sdd_doctor.py", work / "sdd")
            if revision != "worktree":
                for name in MODULES:
                    (work / "logic-map" / name).write_bytes(subprocess.run(
                        ["git", "show", f"{revision}:tools/logic-map/{name}"], cwd=REPO,
                        capture_output=True, check=True).stdout)
            process = subprocess.run(
                [sys.executable, "-c", PROBE], cwd=work / "logic-map", capture_output=True, text=True, timeout=600,
                env={**os.environ, "GIT_CONFIG_GLOBAL": "/dev/null", "GOFLAGS": "-trimpath"})
            print(f"--- {revision}")
            print(process.stdout.rstrip() or process.stderr[-1500:])


if __name__ == "__main__":
    main()

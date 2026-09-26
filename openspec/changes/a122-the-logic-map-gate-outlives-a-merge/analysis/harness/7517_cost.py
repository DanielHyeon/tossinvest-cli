#!/usr/bin/env python3
"""task 7.5.17 재측정 — 엉터리 맨이름 좌표가 판정과 재확인에 드는 시간 (7.5.9 가 인용 해소를 트리 순회에서 추적 목록으로 바꾼 뒤).

실제 저장소를 **읽기만** 한다. 합성 `branch-test-map.md` 산문(맨이름 `nosuch<i>_test.go:1` N 개)을 `test_citation_errors` 에 넣고,
원장을 연 채 판정 시간과 끝의 재확인(`_reads_moved`) 시간을 잰다. 기준 리비전(7.5.9 이전 `1d1e5ca7`)의 코드와 워킹트리 코드를
같은 방식으로 돈다. 판정 줄 수도 센다(해소 실패는 오늘 오류 줄을 안 낸다 — 7.5.37 과 겹친다).

    python3 7517_cost.py [<before-rev>] [N ...]      # 기본 1d1e5ca7 · 20 1000 10000 (before 는 20 만)
"""
import importlib.util
import subprocess
import sys
import tempfile
import time
from pathlib import Path

REPO = next(parent for parent in Path(__file__).resolve().parents if (parent / "tools" / "logic-map").is_dir())
sys.path.insert(0, str(REPO / "tools" / "logic-map"))
import check_analysis as new  # noqa: E402


def measure(module, count: int) -> tuple[float, float, float, int, int]:
    text = "# Branch Test Map\n" + "\n".join(f"| B{i} | leaf | `nosuch{i}_test.go:1` | yes | yes |" for i in range(count))
    with module._ledger() as book:
        started = time.monotonic()
        index = module.test_index(REPO)
        indexed = time.monotonic() - started
        started = time.monotonic()
        errors = module.test_citation_errors("synthetic", text, REPO / "internal", REPO, index)
        judged = time.monotonic() - started
        started = time.monotonic()
        moved = module._reads_moved(REPO, book)
        rechecked = time.monotonic() - started
    assert moved == "", moved
    return indexed, judged, rechecked, len(errors), len(text.encode())


def main() -> None:
    before = next((a for a in sys.argv[1:] if not a.isdigit()), "1d1e5ca781d7716e1d5c2afbeb5a127ceb953425")
    counts = [int(a) for a in sys.argv[1:] if a.isdigit()] or [20, 1000, 10000]
    head = subprocess.run(["git", "rev-parse", "--short=12", "HEAD"], cwd=REPO, capture_output=True, text=True).stdout.strip()
    with tempfile.TemporaryDirectory() as raw:
        for name in ("check_analysis.py", "execution_baseline.py", "role_check.py"):
            (Path(raw) / name).write_bytes(subprocess.run(["git", "show", f"{before}:tools/logic-map/{name}"], cwd=REPO,
                                                          capture_output=True, check=True).stdout)
        sys.path.insert(0, raw)
        spec = importlib.util.spec_from_file_location("ca_before_7517", Path(raw) / "check_analysis.py")
        old = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(old)
        print(f"HEAD {head} · 저장소 {REPO} · {time.strftime('%F %T')}")
        rows = [("before " + before[:8], old, [counts[0]]), ("worktree", new, counts)]
        for label, module, sizes in rows:
            for count in sizes:
                indexed, judged, rechecked, lines, size = measure(module, count)
                print(f"{label:17s} N={count:6d} ({size:7d} B)  색인 {indexed:7.2f}s · 판정 {judged:7.2f}s · "
                      f"재확인 {rechecked:7.2f}s · 판정 줄 {lines}")


if __name__ == "__main__":
    main()

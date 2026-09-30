#!/usr/bin/env python3
"""a112 6.2 봉인 로트 — 이름 붙은 시험이 **실제로 돌았는지** `go test -json` 의 pass 이벤트로 확인한다(5.2.2.1 리뷰 이월 #1).

왜: `go test` 의 종료 코드 0 과 `ok` 는 「시험이 돌았다」를 증명하지 않는다 — 패키지 `init()` 이 시험 이진에서만 `os.Exit(0)` 하면
어떤 시험도 안 돌고 `ok` 가 나온다(codex 4차, ae8a5ff9). 여기서는 종료 코드가 아니라 **각 이름의 `{"Action":"pass","Test":…}` 사건**을
요구한다. 하나라도 pass 사건이 없으면(실행 안 됨 · skip · fail) 실패. 범위는 이 로트의 봉인 시험 이름뿐이다(저장소 전역판은 후속 후보).

사용: verify_named_tests.py [--tags tossos_testseams] <패키지> <시험 이름> [<시험 이름> …]   (저장소 루트 또는 사본 루트에서)
"""
from __future__ import annotations

import json
import re
import subprocess
import sys


def main() -> int:
    args = sys.argv[1:]
    tags = []
    if args and args[0] == "--tags":
        tags, args = ["-tags", args[1]], args[2:]
    package, names = args[0], args[1:]
    if not names:
        print("no test names given — nothing would be proven", file=sys.stderr)
        return 2
    pattern = "^(" + "|".join(re.escape(name) for name in names) + ")$"
    command = ["go", "test", "-json", "-count=1", *tags, "-run", pattern, package]
    completed = subprocess.run(command, capture_output=True, text=True)
    passed, failed, skipped = set(), set(), set()
    for line in completed.stdout.splitlines():
        try:
            event = json.loads(line)
        except json.JSONDecodeError:
            continue
        test = event.get("Test", "")
        if "/" in test:  # 하위 시험은 부모 이름의 증거가 아니다
            continue
        action = event.get("Action")
        if action == "pass":
            passed.add(test)
        elif action == "fail":
            failed.add(test)
        elif action == "skip":
            skipped.add(test)
    missing = [name for name in names if name not in passed]
    print(f"$ {' '.join(command)}")
    print(f"exit={completed.returncode} passed={len(passed & set(names))}/{len(names)}")
    for name in names:
        state = "PASS" if name in passed else "FAIL" if name in failed else "SKIP" if name in skipped else "NOT-RUN"
        print(f"  {state}\t{name}")
    if missing or completed.returncode != 0:
        print("named-test evidence INCOMPLETE — an exit code without pass events proves nothing", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

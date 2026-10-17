#!/usr/bin/env python3
"""a112 6.2 봉인 로트 — 이름 붙은 시험의 **pass 사건이 보고됐는지** `go test -json` 으로 확인한다(5.2.2.1 리뷰 이월 #1).

**증명하는 것 · 못 하는 것(6.2 리뷰 보이스 B #4 뒤 좁힘).** 증명: `-count=1` 로 빌드가 성공했고 test2json 스트림이 이름마다 pass 사건을
보고했으며, 요구한 부모 아래 하위 시험이 skip 되지 않았다. 조용한 조기 종료(`init` 의 `os.Exit(0)` — codex 4차)는 잡는다. **못 하는 것:**
(a) 같은 프로세스 안의 코드가 framing 을 위조하는 것(`init` 이 `--- PASS:` 줄을 찍고 끝나면 pass 사건이 생긴다 — 하네스로는 닫을 수 없다),
(b) 시험 몸이 아무것도 단언하지 않고 끝나는 것(조기 return). pass 사건은 「보고됐다」이지 「무엇을 쟀다」가 아니다.

왜: `go test` 의 종료 코드 0 과 `ok` 는 「시험이 돌았다」를 증명하지 않는다 — 패키지 `init()` 이 시험 이진에서만 `os.Exit(0)` 하면
어떤 시험도 안 돌고 `ok` 가 나온다(codex 4차, ae8a5ff9). 여기서는 종료 코드가 아니라 **각 이름의 `{"Action":"pass","Test":…}` 사건**을
요구한다. 하나라도 pass 사건이 없으면(실행 안 됨 · skip · fail) 실패. 범위는 이 로트의 봉인 시험 이름뿐이다(저장소 전역판은 후속 후보).

사용: verify_named_tests.py [--tags tossos_testseams] <패키지> <시험 이름> [<시험 이름> …]   (저장소 루트 또는 사본 루트에서)
하위 시험은 `Parent/child`(go 의 공백 → `_` 치환 이름 그대로)로 줄 수 있다. 요구한 부모 아래의 skip 은 실패다.
(한계는 위 머리말.)
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
    # 하위 시험 이름(`Parent/child`)도 받는다(6.2 리뷰 codex #4) — -run 은 최상위 이름으로만 거르고, 하위는 사건으로 확인한다.
    parents = sorted({name.split("/", 1)[0] for name in names})
    pattern = "^(" + "|".join(re.escape(name) for name in parents) + ")$"
    command = ["go", "test", "-json", "-count=1", *tags, "-run", pattern, package]
    completed = subprocess.run(command, capture_output=True, text=True)
    passed, failed, skipped = set(), set(), set()
    for line in completed.stdout.splitlines():
        try:
            event = json.loads(line)
        except json.JSONDecodeError:
            continue
        test = event.get("Test", "")
        action = event.get("Action")
        # 요구한 부모 아래의 하위 시험이 skip 되면 부모 pass 는 그 축의 증거가 아니다 — 실패로 센다(6.2 리뷰 codex #4).
        if "/" in test and action == "skip" and test.split("/", 1)[0] in parents:
            failed.add(test + " (subtest skipped under a required parent)")
        if action == "pass":
            passed.add(test)
        elif action == "fail":
            failed.add(test)
        elif action == "skip":
            skipped.add(test)
    missing = [name for name in names if name not in passed]
    skipped_under = sorted(name for name in failed if name.endswith("(subtest skipped under a required parent)"))
    print(f"$ {' '.join(command)}")
    print(f"exit={completed.returncode} passed={len(passed & set(names))}/{len(names)}")
    for name in names:
        state = "PASS" if name in passed else "FAIL" if name in failed else "SKIP" if name in skipped else "NOT-RUN"
        print(f"  {state}\t{name}")
    for name in skipped_under:
        print(f"  SKIP\t{name}")
    if missing or skipped_under or completed.returncode != 0:
        print("named-test evidence INCOMPLETE — an exit code without pass events proves nothing", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

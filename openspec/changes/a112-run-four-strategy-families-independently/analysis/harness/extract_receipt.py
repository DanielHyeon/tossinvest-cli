#!/usr/bin/env python3
"""a112 5.2.2.1 리뷰 수리 — 옮긴 몸통이 옮기기 전과 같은 코드임을 gofmt 정규형으로 비교한다(Manager 조건 ①).

옛 몸통: <before-rev>:internal/app/engine/strategy_entry_supervisor.go 의 `deliverEachStrategyHandoff(…, func(delivered …) error {…})`
새 몸통: 작업 트리 internal/app/engine/strategy_market_handoff_delivery.go 의 같은 closure.
허용 차이: 수신자 이름 둘(`c.Journal` → `campaigns`, `fresh.dispatch` → `dispatcher`) — 옛 쪽에 치환을 적용한 뒤 비교한다.
비교는 주석을 지운 뒤 `package p; var _ = <closure>` 로 감싸 gofmt 로 정규화한 바이트의 동일성이다(= 같은 AST 의 같은 인쇄).
사용: extract_receipt.py <before-rev>
"""
from __future__ import annotations

import hashlib
import re
import subprocess
import sys
from pathlib import Path

ROOT = Path(subprocess.run(["git", "rev-parse", "--show-toplevel"], capture_output=True, text=True, check=True).stdout.strip())
HEAD = "func(delivered strategyhandoff.Delivered) error {"


def closure(text: str) -> str:
    start = text.index(HEAD)
    depth, i = 0, start
    while True:
        ch = text[i]
        if ch == "{":
            depth += 1
        elif ch == "}":
            depth -= 1
            if depth == 0:
                return text[start:i + 1]
        i += 1


def canonical(body: str) -> bytes:
    body = "\n".join(re.sub(r"//.*$", "", line) for line in body.splitlines())
    source = "package p\n\nvar _ = " + body + "\n"
    out = subprocess.run(["gofmt"], input=source.encode(), capture_output=True, check=True).stdout
    return out


def main() -> None:
    rev = sys.argv[1]
    before_file = subprocess.run(["git", "show", f"{rev}:internal/app/engine/strategy_entry_supervisor.go"], cwd=ROOT,
                                 capture_output=True, text=True, check=True).stdout
    after_file = (ROOT / "internal/app/engine/strategy_market_handoff_delivery.go").read_text()
    before = closure(before_file).replace("c.Journal.", "campaigns.").replace("fresh.dispatch.dispatch", "dispatcher.dispatch")
    after = closure(after_file)
    a, b = canonical(before), canonical(after)
    print(f"before {rev}:strategy_entry_supervisor.go closure (renamed) sha256 {hashlib.sha256(a).hexdigest()}")
    print(f"after  worktree strategy_market_handoff_delivery.go closure sha256 {hashlib.sha256(b).hexdigest()}")
    print("EQUAL" if a == b else "DIFFERENT")
    if a != b:
        sys.exit(1)


if __name__ == "__main__":
    main()

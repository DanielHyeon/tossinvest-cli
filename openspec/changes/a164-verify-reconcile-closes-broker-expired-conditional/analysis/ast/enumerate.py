#!/usr/bin/env python3
"""a164 — 대사·정리·HOLD 경로 함수의 AST 산출물에서 분기·반환·호출을 기계적으로 열거함.

손으로 고른 줄이 아니라 `tools/logic-map` 이 낸 좌표 **전부**의 원문 줄을 찍음. 판정 관련 여부는
정규식 `MARK` 하나로만 표시함(표시는 읽기 보조이며, 표시가 없는 줄도 빠짐없이 출력함).

사용(저장소 루트 아님 — 이 디렉터리 기준 경로를 스스로 유도함):
    python3 -I enumerate.py > enumerate-output.md
재생성(같은 디렉터리의 AST 갱신 — 저장소 루트에서):
    GOFLAGS=-trimpath go build -o <scratch>/logicmap ./tools/logic-map
    <scratch>/logicmap --file <file.go> --func <Receiver.Method> > <pkg>--<func>.ast.json
각 AST 의 `source_sha256` 을 실행 때 현재 워킹트리 바이트와 대조하고, 다르면 그 파일을 STALE 로 표시함.
"""

from __future__ import annotations

import hashlib
import json
import re
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[4]

# 판정 입력: 만료·종결 어휘, 보존(Q1)·신선도(Q3), 미결 artifact·HOLD, 거절 코드, 404·오류 분류.
MARK = re.compile(r"EXPIRED|COMPLETED|Status|Triggered|terminal|Cancelled|Filled|Reconciled|Outstanding|"
                  r"Retention|retention|Freshness|refuse\(|Refuse|HOLD|Passed|Prereq|Unsettled|404|NotFound|"
                  r"not-found|errors\.(Is|As)|HasSecond|ChainID")


def mark(text: str) -> str:
    return "MARK" if MARK.search(text) else ""


def main() -> int:
    paths = sorted(HERE.glob("*.ast.json"))
    print("# a164 AST 열거 출력 (`python3 -I enumerate.py` 생성 — 손편집 금지)\n")
    total_b = total_r = 0
    stale = 0
    for path in paths:
        ast = json.loads(path.read_text())
        src_path = ROOT / ast["file"]
        data = src_path.read_bytes()
        lines = data.decode().splitlines()
        fresh = hashlib.sha256(data).hexdigest() == ast["source_sha256"]
        stale += 0 if fresh else 1
        name = (ast.get("receiver", "") + "." if ast.get("receiver") else "") + ast["function"]
        print(f"## `{ast['file']}` `{name}` ({ast['start']['line']}–{ast['end']['line']})"
              f"{'' if fresh else ' **STALE**'}\n")
        branches = ast.get("branches") or []
        returns = ast.get("returns") or []
        print(f"- 산출물 `{path.name}` · 분기 {len(branches)} · 반환 {len(returns)}"
              f" · go {len(ast.get('go_statements') or [])} · defer {len(ast.get('defers') or [])}\n")
        print("| 종류 | id | 줄 | 원문 | 표시 |\n| --- | --- | --- | --- | --- |")
        for b in branches:
            ln = b["at"]["line"]
            text = lines[ln - 1].strip().replace("|", "\\|")
            print(f"| branch:{b['kind']} | {b['id']} | {ln} | `{text}` | {mark(text)} |")
        for r in returns:
            ln = r["at"]["line"]
            text = lines[ln - 1].strip().replace("|", "\\|")
            print(f"| return | | {ln} | `{text}` | {mark(text)} |")
        for c in (ast.get("calls") or []):
            if MARK.search(c.get("text", "")):
                print(f"| call | | {c['at']['line']} | `{c['text']}` | MARK |")
        print()
        total_b += len(branches)
        total_r += len(returns)
    print(f"합계: AST {len(paths)} · 분기 {total_b} · 반환 {total_r} · STALE {stale}")
    return 0


if __name__ == "__main__":
    sys.exit(main())

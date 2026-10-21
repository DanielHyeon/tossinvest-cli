#!/usr/bin/env python3
"""a163 — 시세 신선도 생산자·소비자 함수의 AST 산출물에서 분기·반환·시각 호출을 기계적으로 열거함.

손으로 고른 줄이 아니라 `tools/logic-map` 이 낸 좌표 **전부**의 원문 줄을 찍음. 시각·신선도 관련 여부는
정규식 `CLOCK` 하나로만 표시함(표시는 읽기 보조이며, 표시가 없는 줄도 빠짐없이 출력함).

사용:
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

# 시각·신선도 입력: 벽시계·경과·브로커 시각·관측 시각을 읽는 이름.
CLOCK = re.compile(r"Now|Since|Sub\(|\.Sub\b|After|Before|IsZero|Sleep|Lease|FetchedAt|Timestamp|ObservedAt|"
                   r"ReceivedAt|SourceInstant|QueryPriceEvidenceDuration|MaxQuoteAge|quoteUsable|Stale|stale|"
                   r"evaluated|observation\.at|time\.")


def mark(text: str) -> str:
    return "CLOCK" if CLOCK.search(text) else ""


def main() -> int:
    paths = sorted(HERE.glob("*.ast.json"))
    print("# a163 AST 열거 출력 (`python3 -I enumerate.py` 생성 — 손편집 금지)\n")
    total_b = total_r = 0
    for path in paths:
        ast = json.loads(path.read_text())
        src_path = ROOT / ast["file"]
        data = src_path.read_bytes()
        lines = data.decode().splitlines()
        fresh = hashlib.sha256(data).hexdigest() == ast["source_sha256"]
        name = (ast.get("receiver", "") + "." if ast.get("receiver") else "") + ast["function"]
        print(f"## `{ast['file']}` `{name}` ({ast['start']['line']}–{ast['end']['line']})"
              f"{'' if fresh else ' **STALE**'}\n")
        print(f"- 분기 {len((ast["branches"] or []))} · 반환 {len((ast["returns"] or []))} · go {len(ast.get('go_statements') or [])}"
              f" · defer {len(ast.get('defers') or [])}\n")
        print("| 종류 | id | 줄 | 원문 | 표시 |\n| --- | --- | --- | --- | --- |")
        for b in (ast["branches"] or []):
            ln = b["at"]["line"]
            text = lines[ln - 1].strip().replace("|", "\\|")
            print(f"| branch:{b['kind']} | {b['id']} | {ln} | `{text}` | {mark(text)} |")
        for r in (ast["returns"] or []):
            ln = r["at"]["line"]
            text = lines[ln - 1].strip().replace("|", "\\|")
            print(f"| return | | {ln} | `{text}` | {mark(text)} |")
        clock_calls = [c for c in (ast["calls"] or []) if CLOCK.search(c.get("text", ""))]
        for c in clock_calls:
            print(f"| call | | {c['at']['line']} | `{c['text']}` | CLOCK |")
        print()
        total_b += len((ast["branches"] or []))
        total_r += len((ast["returns"] or []))
    print(f"합계: AST {len(paths)} · 분기 {total_b} · 반환 {total_r}")
    return 0


if __name__ == "__main__":
    sys.exit(main())

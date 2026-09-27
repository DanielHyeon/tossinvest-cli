#!/usr/bin/env python3
"""a095 3판 FLM 분기 표 생성기 — AST 좌표와 커버리지 프로파일을 표 행으로 옮김.

손으로 읽은 분기 주장을 막기 위한 도구임. 표의 각 칸은 다음에서만 옴.

- 조건: 소스의 그 줄 원문
- 창의 호출/return: `ast.json` 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것 (의미가 아니라 위치)
- 진입 실측: `go test -covermode=set` 프로파일에서 **그 줄로 시작하는 블록**의 count 가 0보다 큰지.
  자체 블록이 없는 분기는 `—`

사용:
    python3 branch_rows.py <ast.json> <coverprofile>... > rows.md

저장소 루트(이 파일에서 다섯 단계 위)를 소스 기준으로 삼음 — 절대경로를 적지 않고 유도함.
"""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[5]
MODULE = "github.com/JungHoonGhae/tossinvest-cli/"
PROFILE_LINE = re.compile(r"^(.+?):(\d+)\.(\d+),(\d+)\.(\d+) (\d+) (\d+)$")


def load_blocks(profiles: list[str], relative: str) -> dict[int, int]:
    """파일 하나에 대해 `시작 줄 → 최대 count` 를 모음. 같은 줄에서 시작하는 블록이 여럿이면 최댓값."""
    starts: dict[int, int] = {}
    for profile in profiles:
        for line in Path(profile).read_text(encoding="utf-8").splitlines():
            match = PROFILE_LINE.match(line)
            if not match or match.group(1) != MODULE + relative:
                continue
            start, count = int(match.group(2)), int(match.group(7))
            starts[start] = max(starts.get(start, 0), count)
    return starts


def cell(text: str) -> str:
    """표 칸에 넣을 수 있게 파이프를 이스케이프함."""
    return text.replace("|", "\\|")


def main() -> int:
    ast_path, profiles = sys.argv[1], sys.argv[2:]
    value = json.loads(Path(ast_path).read_text(encoding="utf-8"))
    relative = value["file"]
    source = (ROOT / relative).read_text(encoding="utf-8").splitlines()
    blocks = load_blocks(profiles, relative)
    branches = value.get("branches") or []
    calls = value.get("calls") or []
    returns = value.get("returns") or []
    end = value["end"]["line"] + 1
    print("| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |")
    print("|---|---|---|---|---|---|")
    for index, branch in enumerate(branches):
        line = branch["at"]["line"]
        # 창의 끝은 다음 분기 줄. 같은 줄의 분기(예: `else if`)가 이어지면 창이 비어 있음
        upper = branches[index + 1]["at"]["line"] if index + 1 < len(branches) else end
        window_calls = sorted({c.get("text") or "(unnamed)" for c in calls
                               if line <= c["at"]["line"] < max(upper, line + 1)})
        window_returns = [f":{r['at']['line']}" for r in returns
                          if line <= r["at"]["line"] < max(upper, line + 1)]
        entered = "—" if line not in blocks else ("예" if blocks[line] > 0 else "아니오")
        text = source[line - 1].strip()
        print(f"| {branch['id']} | {branch['kind']} | `:{line}` `{cell(text)}` | "
              f"{', '.join(f'`{cell(c)}`' for c in window_calls) or '—'} | "
              f"{', '.join(window_returns) or '—'} | {entered} |")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

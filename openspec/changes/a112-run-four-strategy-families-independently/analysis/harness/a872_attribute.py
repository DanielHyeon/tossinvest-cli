#!/usr/bin/env python3
"""a112 8.7.2 — AST 분기마다 arm 진입 수와 귀속 시험을 커버리지 프로파일에서 읽는다.

쓰는 법:
  a872_attribute.py <ast.json> <프로파일 디렉터리> [--markdown <스위트 이름>]

--markdown 을 주면 FLM/BTM 에 그대로 붙일 표 행(`| B1 | if | 72:2 | … |`)과 호출 표를 낸다.

프로파일 디렉터리는 a872_pertest_cover.sh 의 출력이다(suite.cov + <Test>.cov).

arm 은 분기 좌표 **뒤에서 처음 시작하는** 커버리지 블록이다(`if`/`range` 의 몸통 블록).
귀속 완전성은 등식으로 잰다: 시험별 진입 수의 합 == 스위트 진입 수. 어긋나면 그 행에
`ATTRIBUTION MISMATCH` 를 찍는다 — 조용히 넘기면 표가 모르는 시험이 arm 에 들어간 것을
아무도 못 본다.
"""

from __future__ import annotations

import json
import sys
from pathlib import Path

MODULE = "github.com/JungHoonGhae/tossinvest-cli/"


def read_profile(path: Path, source: str) -> dict[tuple[int, int, int, int], int]:
    """(시작줄, 시작열, 끝줄, 끝열) → 진입 수. 같은 블록이 여러 번 나오면 더한다."""
    blocks: dict[tuple[int, int, int, int], int] = {}
    target = MODULE + source
    for line in path.read_text().splitlines()[1:]:
        location, _, rest = line.partition(" ")
        name, _, span = location.rpartition(":")
        if name != target:
            continue
        start, end = span.split(",")
        sl, sc = (int(x) for x in start.split("."))
        el, ec = (int(x) for x in end.split("."))
        count = int(rest.split()[1])
        key = (sl, sc, el, ec)
        blocks[key] = blocks.get(key, 0) + count
    return blocks


def arm_block(blocks, line: int, column: int):
    after = sorted(key for key in blocks if (key[0], key[1]) > (line, column))
    return after[0] if after else None


def main() -> int:
    ast = json.loads(Path(sys.argv[1]).read_text())
    profiles = Path(sys.argv[2])
    source = ast["file"]
    suite = read_profile(profiles / "suite.cov", source)
    tests = {p.stem: read_profile(p, source) for p in sorted(profiles.glob("Test*.cov"))}
    markdown = sys.argv[4] if len(sys.argv) > 4 and sys.argv[3] == "--markdown" else ""
    if not markdown:
        print(f"# {ast['function']} ({source}) per-test profiles={len(tests)}")
    for branch in ast.get("branches") or []:
        line, column = branch["at"]["line"], branch["at"]["column"]
        block = arm_block(suite, line, column)
        if block is None:
            print(f"{branch['id']}\t{branch['kind']}\t{line}:{column}\tNO BLOCK")
            continue
        total = suite[block]
        entered = {name: prof.get(block, 0) for name, prof in tests.items() if prof.get(block, 0)}
        summed = sum(entered.values())
        verdict = "" if summed == total else f"  ATTRIBUTION MISMATCH (sum {summed} != suite {total})"
        names = ", ".join(f"`{name}`" for name in sorted(entered))
        if markdown:
            state = f"arm entered {total}x ({markdown})" if total else f"arm not entered ({markdown})"
            print(f"| {branch['id']} | {branch['kind']} | {line}:{column} | {state}; "
                  f"{names or 'no per-test profile entered it'}{verdict} |")
        else:
            print(f"{branch['id']}\t{branch['kind']}\t{line}:{column}\tsuite={total}\t{names or '(not entered)'}{verdict}")
    if markdown:
        print()
        print("| Callee expression | Position |")
        print("|---|---|")
        for call in ast.get("calls") or []:
            print(f"| `{call.get('text')}` | {call['at']['line']}:{call['at']['column']} |")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

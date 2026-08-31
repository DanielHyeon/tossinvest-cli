#!/usr/bin/env python3
"""좌표가 **무엇을 가리키는지**(역할) 대조한다.

check_analysis.py 의 좌표 검사는 두 가지만 본다: `- Source: … (L-L)` 범위와
"AST branches N" 이라는 개수 주장. 그래서 분기 표의 한 줄에 분기가 아닌 좌표를
적어도, 호출 표를 중간에서 잘라도 게이트는 통과했다 — a112 3라운드 적대 리뷰가
그 구멍에 네 개의 오류를 심어 전부 통과시켰다.

**여기서 보는 것은 좌표뿐이다.** 표의 "AST kind" 칸에는 산문이 들어오고(예:
`case KindKRNetFlow → KR only`), 이름 칸에는 `(unnamed)` 같은 표기가 들어온다.
그 칸들을 대조하면 **맞는 줄에서 터진다**. 좌표는 그럴 수 없다 — 좌표는
ast.json 이 분기라고 부른 자리이거나 아니거나 둘 중 하나다.

**무엇을 검사하지 않는지도 적는다.** 침묵은 주장이 아니다. 반환 좌표를
나열하지 않은 문서, 분기 표가 없는 문서, 손으로 쓴 호출 분석 표는 여기서
아무것도 요구받지 않는다. 저장소의 401개 번들에 돌려 측정했고, 그 규칙이라야
맞는 줄에서 터지지 않는다.
"""

from __future__ import annotations

import re

COORD = re.compile(r"(\d+):(\d+)")
RETURN_LINE = re.compile(r"Exact AST return (?:nodes|positions)\s*:\s*(.+)")
BRANCH_ROW = re.compile(r"^\|\s*`?B\d+`?\s*\|")
TABLE_SEPARATOR = re.compile(r"^\|[\s|:-]+\|?\s*$")
# 열거형 호출 표의 표지. 손으로 쓴 분석 표(`| Callee | Why called | …`, 저장소에
# 204개)와 구별하는 유일한 표지이며, **그 표만 완전성을 주장한다.** 손으로 쓴
# 표는 한 줄이 호출 넷을 묶기도 하고 줄 번호만 적기도 해서 1:1 대조 대상이 아니다.
ENUMERATED_CALLS = re.compile(r"^\|\s*Callee expression\s*\|")


def _coords(nodes) -> list[str] | None:
    """`line:column` 로 온전한 좌표만 돌려준다.

    둘 중 하나라도 없는 노드는 비교 대상이 아니다. 없는 값을 0 으로 채워
    비교하면 문서가 옳아도 틀렸다고 말하게 된다 — 그 순간 이 검사는 맞는 줄에서
    터지는 검사가 된다."""
    out = []
    for node in nodes or []:
        if not isinstance(node, dict):
            continue
        at = node.get("at")
        if not isinstance(at, dict):
            continue
        line, column = at.get("line"), at.get("column")
        if line is None or column is None:
            # 한 노드라도 온전하지 않으면 그 역할 전체를 모르는 것으로 둔다.
            # 그 노드만 빼면 목록이 짧아지고, 짧아진 목록은 **맞는 표**를
            # 개수가 안 맞는다고 말한다.
            return None
        out.append(f"{line}:{column}")
    return out


def _sections(text: str) -> dict[str, list[str]]:
    sections: dict[str, list[str]] = {}
    current = None
    for line in text.splitlines():
        if line.startswith("## "):
            current = line[3:].strip()
            sections.setdefault(current, [])
        elif current is not None:
            sections[current].append(line)
    return sections


def _first_coord(line: str) -> str | None:
    found = COORD.search(line)
    return f"{found.group(1)}:{found.group(2)}" if found else None


def role_errors(target: str, text: str, value: dict) -> list[str]:
    """FLM 산문이 인용한 좌표를 ast.json 의 역할별 목록과 맞춰 본다."""
    errors: list[str] = []
    if not value:
        return errors
    returns = _coords(value.get("returns"))
    branches = _coords(value.get("branches"))
    calls = _coords(value.get("calls"))
    if returns is None or branches is None or calls is None:
        # ast.json 의 좌표가 온전하지 않다. 이 검사는 아무 말도 하지 않는다 —
        # 모르는 것을 근거로 삼는 것이 이 change 가 고치고 있는 결함이다.
        return errors
    sections = _sections(text)

    # 1) 반환 좌표를 나열하면 그 목록이 곧 완전성 주장이다("Exact"). 하나라도
    #    어긋나면 그 뒤의 모든 서술이 다른 함수를 가리킨다.
    stated = RETURN_LINE.search(text)
    if stated:
        claimed = [f"{a}:{b}" for a, b in COORD.findall(stated.group(1))]
        if claimed and claimed != returns:
            errors.append(
                f"{target}: function-logic-map.md lists return positions {claimed} "
                f"but ast.json returns are {returns}"
            )

    # 2) 분기 표의 좌표는 분기여야 한다. 분기가 없는 함수는 분기를 인용할 수
    #    없으므로, 그때는 반환·호출 좌표를 분기 자리에 적는 것만 막는다.
    cited_branches = [
        coord
        for line in sections.get("Branches and early returns", [])
        if BRANCH_ROW.match(line)
        for coord in [_first_coord(line)]
        if coord
    ]
    if branches:
        if cited_branches and cited_branches != branches:
            errors.append(
                f"{target}: branch table cites {cited_branches} but ast.json branches "
                f"are at {branches}"
            )
    else:
        for coord in cited_branches:
            if coord in returns or coord in calls:
                errors.append(
                    f"{target}: branch table cites {coord}, which ast.json classifies "
                    f"as a return or call, in a function with no branches"
                )

    # 3) 열거형 호출 표는 잘려 있으면 안 된다. 잘린 표는 "여기까지가 전부"라고
    #    말하면서 나머지를 숨긴다 — a112 의 세 번들이 정확히 40행에서 멈춰 있었고
    #    (실제 64·46·91개), 게이트는 그것을 통과시켰다.
    call_lines = [
        line
        for line in sections.get("Calls and live bindings", [])
        if line.startswith("|") and not TABLE_SEPARATOR.match(line)
    ]
    if not call_lines or not ENUMERATED_CALLS.match(call_lines[0]):
        return errors
    cited_calls = [_first_coord(line) for line in call_lines[1:]]
    if any(coord is None for coord in cited_calls):
        # 좌표 없이 줄 번호만 적은 표는 열거형이 아니다.
        return errors
    if cited_calls != calls:
        unknown = sorted({coord for coord in cited_calls if coord not in calls})
        if unknown:
            errors.append(
                f"{target}: call table cites {unknown[:5]}, which ast.json does not "
                f"list as call positions"
            )
        elif len(cited_calls) != len(calls):
            errors.append(
                f"{target}: call table enumerates {len(cited_calls)} call(s) but "
                f"ast.json has {len(calls)}; a table that stops early still reads as "
                f"the whole list"
            )
        else:
            errors.append(
                f"{target}: call table lists call positions in an order ast.json does "
                f"not have"
            )
    return errors

#!/usr/bin/env python3
"""좌표가 **무엇을 가리키는지**(역할) 대조한다.

check_analysis.py 의 좌표 검사는 두 가지만 본다: `- Source: … (L-L)` 범위와
"AST branches N" 이라는 개수 주장. 그래서 분기 표의 한 줄에 분기가 아닌 좌표를
적어도, 호출 표를 중간에서 잘라도 게이트는 통과했다 — a112 3라운드 적대 리뷰가
그 구멍에 네 개의 오류를 심어 전부 통과시켰다.

**여기서 보는 것은 좌표뿐이다.** 표의 "AST kind" 칸에는 산문이 들어오고(예:
`case KindKRNetFlow → KR only`), 이름 칸에는 `(unnamed)` 같은 표기가 들어온다.
그 칸들을 대조하면 **맞는 줄에서 터진다**.

**좌표도 그럴 수 있다.** 앞 판본은 "좌표는 그럴 수 없다"고 적고 행 아무 데서나
첫 `숫자:숫자` 를 집었다. 4차 적대 리뷰가 저장소 전체(3014 번들)에 돌려 오탐 셋을
찾았다 — 대비율 `4.5:1`, 장 시작 시각 `09:00`, 주소 `127.0.0.1:0`. 앞 판본이
"저장소 전체"라고 적은 401 개는 **비아카이브 부분집합**이었고, 결함이 사는
차원(산문의 다양함)은 아카이브 쪽에 있었다. 개수만 늘리고 종류를 안 바꾼
측정이었다. 이제 좌표는 **그것만 담은 칸**에서만 읽는다.

**무엇을 검사하지 않는지도 적는다.** 침묵은 주장이 아니다. 반환 좌표를
나열하지 않은 문서, 분기 표가 없는 문서, 그리고 **호출 좌표를 빠짐없이 같은
순서로 적은 표가 한 번도 없는 change** 의 손으로 쓴 호출 분석 표는 여기서
아무것도 요구받지 않는다. 저장소의 **비아카이브 404 개와 아카이브 2610 개를
합한 3014 개 전부**에 돌려 측정했고, 그 규칙이라야 맞는 줄에서 터지지 않는다.

**호출 표의 면제는 없어졌고, 강제 판정은 이름이 아니라 내용으로 한다.** 앞
판본들은 완전성 표지의 **철자**로 판정했고 세 라운드 연속 뚫렸다 — 표지를 첫
표의 첫 줄에서만 찾아 위에 한 줄 얹기(5차), 공백 하나·들여쓰기·굵게(6차),
표를 다른 `##` 절로 옮기기(7차). 셋 다 화면에서는 아무것도 안 달라진다.
이제 판정은 "이 번들의 산문 어딘가에 그 함수의 호출 좌표를 빠짐없이 같은
순서로 적은 표가 있는가"이고, 그것이 참이면 그 change 의 **모든** 번들이
`| Callee expression | Position |` 표로 섹션을 열어야 한다.
재측정: 3014 개, 강제가 켜진 change 1 개, findings 0.
"""

from __future__ import annotations

import re

COORD = re.compile(r"(\d+):(\d+)")
# 좌표는 **그것만 담은 칸**에서 읽는다. 실제 모양을 3014 번들에서 세어 정했다:
# function-logic-map 은 `N:N` 780 칸, branch-test-map 은 `if at N:N` 356 ·
# `range at N:N` 51 · `N:N` 22 · `case at N:N` 22 · `switch at N:N` 6 ·
# `else at N:N` 2 · `branchless happy path at N:N` 2, 그리고 그 뒤에 ` — 산문`.
# 산문 칸(`computed contrast is below 4.5:1`)은 이 모양이 아니다.
CELL_COORD = re.compile(
    r"^`?(?:[A-Za-z][\w ]*\bat\s+)?(\d+):(\d+)`?(?:\s*[—-].*)?$"
)
RETURN_LINE = re.compile(r"Exact AST return (?:nodes|positions)\s*:\s*(.+)")
BRANCH_ROW = re.compile(r"^\|\s*`?B\d+`?\s*\|")
TABLE_SEPARATOR = re.compile(r"^\|[\s|:-]+\|?\s*$")
# 열거형 호출 표의 표지. 손으로 쓴 분석 표(`| Callee | Why called | …`, 저장소에
# **226개** — 앞 판본이 적은 204 는 그중 최빈 헤더 철자의 수였다)와 구별한다.
#
# **이 표지는 강제를 켜는 스위치가 아니다.** 앞 판본들에서는 그랬고, 그래서 세
# 라운드 연속 뚫렸다 — 표지를 저자가 고르면 감사를 받을지를 저자가 정한다.
# 강제 판정은 `call_enumeration_in_use` 가 표의 **내용**으로 한다.
#
# 이 표지가 하는 일은 하나다: 강제가 켜진 change 에서 **어느 표를 1:1 로
# 대조할지** 고른다. 그 표는 섹션을 열어야 하고, 표지를 달고 좌표를 빼면
# 오류다(호출이 없는 함수만 좌표 없는 표를 적을 수 있고, 그때 `ast.json` 의
# 호출이 비어 있는지 확인한다). 표지가 없으면 그 번들이 터진다 — 표지를
# 바꾸거나 지우는 것은 이득이 아니라 손해다.
#
# 철자는 `enumeration_header` 가 정규화해서 본다. 공백 둘·들여쓰기·굵게는
# 마크다운이 똑같이 그리므로 똑같이 읽어야 한다.
EMPHASIS = re.compile(r"[*_`]")


def enumeration_header(line: str) -> bool:
    """이 줄이 열거 표의 표지인가. **철자를 정규화해서** 본다.

    앞 판본은 머리글 칸의 철자를 그대로 맞춰 보는 정규식이었다. 6차 적대 리뷰가
    공백 하나(`Callee  expression`), 앞 들여쓰기, 굵게(`**Callee expression**`)
    셋으로 표지를 무력화했다 — 셋 다 화면에서는 똑같이 보인다. 보이지 않는
    회피는 회피 중에 가장 나쁘다."""
    cell = EMPHASIS.sub("", _first_cell(line))
    return " ".join(cell.split()).lower() == "callee expression"


def _node_pairs(nodes) -> list[tuple[str, str]] | None:
    """노드에서 `(좌표, 철자)` 쌍을 **한 번에** 만든다.

    좌표 목록과 철자 목록을 따로 만들면 둘이 다른 규칙으로 노드를 건너뛰다가
    어긋난다. 그러면 **맞는 문서**가 한 칸씩 밀린 목록 때문에 터진다 — 6차·7차
    적대 리뷰가 dict 아닌 노드, `at` 없는 노드, `at` 이 dict 아닌 노드 셋으로
    각각 보였다. 한 곳에서 만들면 어긋날 수가 없다.

    좌표가 온전하지 않은 노드가 하나라도 있으면 `None` 을 돌려준다 — 그 노드만
    빼면 목록이 짧아지고, 짧아진 목록은 **맞는 표**를 개수가 안 맞는다고 말한다."""
    out: list[tuple[str, str]] = []
    for node in nodes or []:
        if not isinstance(node, dict):
            continue
        at = node.get("at")
        if not isinstance(at, dict):
            continue
        line, column = at.get("line"), at.get("column")
        if line is None or column is None:
            return None
        out.append((f"{line}:{column}", node.get("text") or "(unnamed)"))
    return out


def _coords(nodes) -> list[str] | None:
    """`line:column` 로 온전한 좌표만 돌려준다."""
    pairs = _node_pairs(nodes)
    return None if pairs is None else [coord for coord, _ in pairs]


def _call_names(nodes) -> list[str]:
    """호출 노드의 **철자**를 좌표와 같은 순서·같은 규칙으로 돌려준다.

    철자가 없는 노드는 기존 문서가 쓰는 `(unnamed)` 로 읽는다. 표지를 단 133개
    번들 안에 15개, 저장소 전체(3014 번들)로는 189 번들에 281개다 —
    `extract_go_ast` 가 함수 리터럴·괄호 호출·형 변환의 callee 철자를 비워 두기
    때문이다. 두 번째 change 가 표지를 다는 순간 이 탈출구는 15가 아니라 281이 된다."""
    pairs = _node_pairs(nodes)
    return [] if pairs is None else [name for _, name in pairs]


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
    """행에서 좌표를 읽는다. **칸 하나가 통째로 좌표일 때만** 읽는다.

    행 아무 데서나 첫 숫자쌍을 집으면 시각·대비율·포트가 좌표가 된다. 그러면
    맞는 문서가 게이트에서 터진다 — 이 검사가 고치려던 바로 그 병이다."""
    for cell in line.strip().strip("|").split("|"):
        found = CELL_COORD.match(cell.strip())
        if found:
            return f"{found.group(1)}:{found.group(2)}"
    return None


def _first_cell(line: str) -> str:
    """행의 첫 칸을 백틱을 벗겨 돌려준다. 열거 표에서 그 칸이 callee 철자다."""
    cells = line.strip().strip("|").split("|")
    return cells[0].strip().strip("`").strip() if cells else ""


def _branch_coord_errors(
    target: str, where: str, rows: list[str],
    branches: list[str], returns: list[str], calls: list[str],
) -> list[str]:
    """분기 표 한 벌의 좌표가 정말 분기인지 본다.

    분기가 없는 함수는 분기를 인용할 수 없으므로, 그때는 반환·호출 좌표를
    분기 자리에 적는 것만 막는다."""
    errors: list[str] = []
    cited = [c for line in rows for c in [_first_coord(line)] if c]
    if branches:
        if cited and cited != branches:
            errors.append(
                f"{target}: {where} cites {cited} but ast.json branches "
                f"are at {branches}"
            )
        return errors
    for coord in cited:
        if coord in returns or coord in calls:
            errors.append(
                f"{target}: {where} cites {coord}, which ast.json classifies "
                f"as a return or call, in a function with no branches"
            )
    return errors


def _tables(lines: list[str]) -> list[list[str]]:
    """섹션 안의 **연속 표를 전부** 돌려준다.

    강제 판정은 첫 표만 봐서는 안 된다 — 5차 적대 리뷰가 열거 위에 `|` 한 줄을
    얹어 첫 표를 가짜로 만들었다. 열거가 섹션 어디에 있든 이 change 는 열거를
    쓰는 것이다. "열거가 **첫 표**여야 한다"는 별개의 요구이고 그쪽은
    `_call_table_errors` 가 본다."""
    tables: list[list[str]] = []
    current: list[str] = []
    for line in lines:
        if line.strip().startswith("|"):
            current.append(line)
        elif current:
            tables.append(current)
            current = []
    if current:
        tables.append(current)
    return [
        [line for line in table if not TABLE_SEPARATOR.match(line.strip())]
        for table in tables
    ]


def _first_table(lines: list[str]) -> list[str]:
    """섹션에서 **첫 번째 연속 표 한 벌**만 돌려준다.

    섹션 전체의 `|` 줄을 다 모으면, 표 아래에 붙인 손으로 쓴 주석 표의 행이
    열거 안으로 섞여 들어온다. 그러면 잘린 표가 뒤 표의 줄 수로 채워져 개수가
    맞아 버린다 — 열거를 세는 검사가 세는 대상을 저자에게 넘겨주는 셈이다."""
    table: list[str] = []
    for line in lines:
        if line.strip().startswith("|"):
            table.append(line)
        elif table:
            break
    return [line for line in table if not TABLE_SEPARATOR.match(line)]


def call_enumeration_in_use(bundles) -> bool:
    """이 change 가 열거형 호출 표를 **실제로** 쓰는가.

    `bundles` 는 `(function-logic-map.md 본문, ast.json 값)` 쌍이다. 어느 한
    번들의 첫 표가 그 함수의 호출 좌표를 **빠짐없이 같은 순서로** 적고 있으면
    이 change 는 열거를 쓰는 것이고, 그러면 모든 번들이 써야 한다.

    **판정 근거가 표지 철자였던 판본이 두 번 뚫렸다.** 5차는 표지를 첫 표의 첫
    줄에서만 찾아서 위에 `|` 한 줄만 얹으면 꺼졌고, 6차는 철자 자체를 공백 하나로
    비켜 갔다. 둘 다 화면에서는 아무것도 달라지지 않는다. 이름이 있는지 보는
    검사는 반드시 뚫린다 — 그래서 이제 이름이 아니라 **내용**을 본다.

    표를 **문서 어디에서나** 찾는다. 7차 적대 리뷰가 `## Calls and live bindings`
    안만 보던 판본을 반증했다 — 133 개 번들의 열거를 통째로 다른 `##` 절로
    옮기면 표지도 행도 그대로 남은 채 감사만 꺼졌다. 판정을 절 이름에 걸면
    그것도 이름 검사다.

    **남는 회피를 정확히 적는다.** 호출자(`check_analysis._bundle_text`)가 번들
    디렉터리에서 **읽히는 파일을 전부** 이어서 넘긴다 — 이름으로도 확장자로도
    거르지 않는다. 그러니 이 판정이 꺼지려면 그 change 의 어느 번들의 어느
    파일에도 호출 좌표를 빠짐없이 같은 순서로 적은 표가 없어야 한다. 길은
    둘이고 둘 다 크다: (a) 좌표를 지워 열거를 정말 그만두거나, (b) 표를 번들
    디렉터리 밖으로 내보내거나. 어느 쪽이든 133 개 번들을 다 고쳐야 하고,
    (a) 는 완전성 주장까지 함께 지운다.

    이 문단은 **다섯 라운드 동안** 매번 코드가 보는 범위보다 한 칸 넓게 적혀
    있었다: 절 → 표지 철자 → `##` 절 → 파일 이름 둘 → 확장자 `.md` 하나.
    같은 실수가 다섯 번 난 이유는 하나다 — 세는 **범위**를 확인하지 않고
    문장을 썼다. 이번엔 호출자가 아무 목록도 갖지 않는다.

    저장소 3014 번들로 재 보면 이 판정에 걸리는 change 는 a112 하나다 — 다른
    16 개 change 의 손으로 쓴 표 중 우연히 1:1 인 것은 없다."""
    for text, value in bundles:
        calls = _coords((value or {}).get("calls"))
        if not calls:
            continue
        for table in _tables(text.splitlines()):
            cited = [_first_coord(row) for row in table[1:]]
            if cited and cited == calls:
                return True
    return False


def _call_table_errors(
    target: str, section: list[str], calls: list[str], names: list[str],
    required: bool,
) -> list[str]:
    """호출 표가 무엇을 주장하는지 보고, 주장한 만큼 대조한다."""
    table = _first_table(section)
    if not table or not enumeration_header(table[0]):
        if required:
            return [
                f"{target}: this change enumerates call positions elsewhere, so "
                f"'## Calls and live bindings' must open with a "
                f"'| Callee expression | Position |' table; it does not"
            ]
        return []
    rows = table[1:]
    cited = [_first_coord(row) for row in rows]
    blank = [index for index, coord in enumerate(cited) if coord is None]
    if len(blank) == len(rows):
        # 좌표를 하나도 안 적은 표는 "적을 것이 없다"는 뜻으로만 읽는다.
        if calls:
            return [
                f"{target}: call table wears the completeness header but cites no "
                f"call position, while ast.json has {len(calls)}"
            ]
        return []
    if blank:
        # 앞 판본은 여기서 조용히 빠져나갔다 — 그 문이 39개를 감사 밖에 뒀다.
        return [
            f"{target}: call table wears the completeness header but "
            f"{len(blank)} of {len(rows)} row(s) carry no coordinate cell "
            f"(first at row {blank[0] + 1}); a partly numbered table still reads "
            f"as the whole list"
        ]
    if cited == calls:
        # 좌표가 맞았으면 이제 **철자**를 본다. 좌표만 보는 검사는 "그 자리에
        # 호출이 있다"까지만 증명하고, 그 행이 그 호출을 **가리킨다**는 것은
        # 증명하지 않는다 — 표지가 완전성을 주장하는 열이 바로 이 열이다.
        # 5차 적대 리뷰가 adaptPrices 의 두 행을 `os.Exit`·`exec.Command` 로
        # 바꿔도 게이트가 조용한 것을 보였다.
        wrong = [
            (index + 1, _first_cell(row), names[index])
            for index, row in enumerate(rows)
            if index < len(names) and _first_cell(row) != names[index]
        ]
        if wrong:
            row, said, want = wrong[0]
            return [
                f"{target}: call table row {row} at {cited[row - 1]} names "
                f"`{said}` but ast.json calls `{want}` there"
                + (f" ({len(wrong)} row(s) disagree)" if len(wrong) > 1 else "")
            ]
        return []
    unknown = sorted({coord for coord in cited if coord not in calls})
    if unknown:
        return [
            f"{target}: call table cites {unknown[:5]}, which ast.json does not "
            f"list as call positions"
        ]
    if len(cited) != len(calls):
        return [
            f"{target}: call table enumerates {len(cited)} call(s) but "
            f"ast.json has {len(calls)}; a table that stops early still reads as "
            f"the whole list"
        ]
    return [
        f"{target}: call table lists call positions in an order ast.json does "
        f"not have"
    ]


def role_errors(
    target: str, text: str, value: dict, branch_map: str = "",
    require_calls: bool = False,
) -> list[str]:
    """FLM 산문이 인용한 좌표를 ast.json 의 역할별 목록과 맞춰 본다.

    require_calls 는 이 번들이 속한 change 가 열거형 호출 표를 쓰는지다. 쓰면
    이 번들도 그 표를 갖춰야 한다 — 표지를 번들마다 고르는 것이 곧 면제였다.

    branch_map 을 주면 branch-test-map.md 의 앵커 좌표도 같은 규칙으로 본다.
    **그 파일은 같은 분기 좌표를 다시 적는다.** 앞 판본은 FLM 만 받았고, 그래서
    이 로트가 ast.json 과 FLM 을 재기준화하면서 branch-test-map 네 벌 36 줄을
    옛 줄번호로 남겼는데 게이트가 조용히 통과시켰다."""
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
    errors.extend(_branch_coord_errors(
        target, "branch table",
        [l for l in sections.get("Branches and early returns", []) if BRANCH_ROW.match(l)],
        branches, returns, calls,
    ))
    errors.extend(_branch_coord_errors(
        target, "branch test map",
        [l for l in branch_map.splitlines() if BRANCH_ROW.match(l)],
        branches, returns, calls,
    ))

    # 3) 열거형 호출 표는 잘려 있으면 안 된다. 잘린 표는 "여기까지가 전부"라고
    #    말하면서 나머지를 숨긴다 — a112 의 세 번들이 정확히 40행에서 멈춰 있었고
    #    (실제 64·46·91개), 게이트는 그것을 통과시켰다.
    errors.extend(_call_table_errors(
        target, sections.get("Calls and live bindings", []), calls,
        _call_names(value.get("calls")), require_calls,
    ))
    return errors

# Function Logic Map: `_is_ancestor` (Python, a122 task 7.5.2.1)

## task 7.5.2.1 — 판정이 읽는 입력을 전부 세고 하나씩 묶는다 (7.5.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:548-564` · 분기 1 · 반환 1 · raise 1 (편집 전 L530-534 · 분기 0 · 반환 1 · raise 0, `ast.before-7.5.2.1.json` = revision `fc35eb2d`).

git 이 "예"(0) · "아니오"(1) 말고 답하면 결함(`RuntimeError`, git 의 말을 담아). rc 128 을 "조상 아님" 으로 읽던 판본은 결함을 착지 거절 사유로 만들어 복구 조언까지 붙였다(레드팀 repro_a3). 호출자가 이제 `HEAD` 대신 고정한 sha 를 넘긴다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 559 | If | `if process.returncode in (0, 1):` |

| raise 줄 | 소스 |
|---|---|
| 561 | `raise RuntimeError(f'cannot tell whether {older[:12]} precedes {newer[:12]}: ' + _first_li` |

# Function Logic Map: `_commits_after` (Python, a122 task 7.5.2.1)

## task 7.5.2.1 — 판정이 읽는 입력을 전부 세고 하나씩 묶는다 (7.5.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:581-590` · 분기 1 · 반환 1 · raise 0 (편집 전 L551-557 · 분기 1 · 반환 1 · raise 0, `ast.before-7.5.2.1.json` = revision `fc35eb2d`).

상징 `HEAD` 대신 명령이 **한 번** 푼 sha(`head`)를 받는다. 7.5.2 는 `HEAD` 를 열 자리에서 따로 읽고 지문에 표본 하나만 넣었다 — 기록은 표본 **전에** 읽혔고 뒤의 git 호출은 살아 있는 `HEAD` 를 다시 읽어서 가지 전환 한 번(레드팀 repro_b)도, 떠났다 돌아온 `HEAD`(이 세션이 재현)도 rc 0 이었다. 창 줄의 커밋 수가 판정과 같은 역사를 말한다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 590 | IfExp | `'' if process.returncode else process.stdout.strip()` |

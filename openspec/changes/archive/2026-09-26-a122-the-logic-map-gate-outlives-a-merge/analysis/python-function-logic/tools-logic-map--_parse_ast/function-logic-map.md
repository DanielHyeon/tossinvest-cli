# Function Logic Map: `_parse_ast` (Python, a122 task 7.5.2.1)

## task 7.5.2.1 — 판정이 읽는 입력을 전부 세고 하나씩 묶는다 (7.5.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:1430-1447` · 분기 9 · 반환 4 · raise 0 (새 함수).

새 함수. 모양 검사의 **유일한** 자리 — 바이트를 값으로 푸는 곳. 사전 · `start`/`end` 사전 · `branches`/`calls`/`returns` 목록(비었거나 없으면 된다). 7.5.2 는 대상 판정 안에서만 봐서, 대상 판정보다 **먼저** 같은 값을 쓰는 열거형 호출 판정이 열 가지 모양에서 traceback 이었다(이 세션이 재현). 저장소 3,048 번들 중 걸리는 것 0.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1438 | Try | `try:` |
| B2 | 1440 | ExceptHandler | `except ValueError:` |
| B3 | 1442 | If | `if not isinstance(value, dict):` |
| B4 | 1444 | BoolOp | `any((value.get(key) and (not isinstance(value[key], dict)) for key in _AST_OBJECTS)) or any((value.get(key) and (not isi` |
| B5 | 1444 | BoolOp | `value.get(key) and (not isinstance(value[key], dict))` |
| B6 | 1444 | If | `if any((value.get(key) and (not isinstance(value[key], dict)) for key in _AST_OBJECTS)) or any((value.get(key) and (not ` |
| B7 | 1444 | comprehension | ` for key in _AST_OBJECTS` |
| B8 | 1445 | BoolOp | `value.get(key) and (not isinstance(value[key], list))` |
| B9 | 1445 | comprehension | ` for key in _AST_LISTS` |

## task 7.5.2.2 — 스냅숏은 끝에서 디스크와 다시 대조한다 (7.5.2.1 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:1476-1495` · 분기 9 · 반환 4 · raise 0 (편집 전 L1430-1447 · 분기 9 · 반환 4 · raise 0, `ast.before-7.5.2.2.json` = revision `908a8a36`).

`RecursionError`(아주 깊은 JSON)도 `invalid` — `RuntimeError` 의 하위형이라 결함으로 올라가 대상 이름 없이 판정 전체를 한 줄로 바꿨다(재리뷰 적대).

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1484 | Try | `try:` |
| B2 | 1486 | ExceptHandler | `except (ValueError, RecursionError):` |
| B3 | 1490 | If | `if not isinstance(value, dict):` |
| B4 | 1492 | BoolOp | `any((value.get(key) and (not isinstance(value[key], dict)) for key in _AST_OBJECTS)) or any((value.get(key) and (not isi` |
| B5 | 1492 | BoolOp | `value.get(key) and (not isinstance(value[key], dict))` |
| B6 | 1492 | If | `if any((value.get(key) and (not isinstance(value[key], dict)) for key in _AST_OBJECTS)) or any((value.get(key) and (not ` |
| B7 | 1492 | comprehension | ` for key in _AST_OBJECTS` |
| B8 | 1493 | BoolOp | `value.get(key) and (not isinstance(value[key], list))` |
| B9 | 1493 | comprehension | ` for key in _AST_LISTS` |

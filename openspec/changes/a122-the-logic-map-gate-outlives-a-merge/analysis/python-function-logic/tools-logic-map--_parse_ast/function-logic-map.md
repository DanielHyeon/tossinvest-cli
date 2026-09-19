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

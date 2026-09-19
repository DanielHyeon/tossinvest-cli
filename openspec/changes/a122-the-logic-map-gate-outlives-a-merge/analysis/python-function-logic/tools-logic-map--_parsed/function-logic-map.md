# Function Logic Map: `_parsed` (Python, a122 task 7.5.2)

## task 7.5.2 — 판정은 한 번 읽은 바이트로 선다 (수리한 트리의 재리뷰, 2026-09-19)

`tools/logic-map/check_analysis.py:1355-1367` · 분기 3 · 반환 3 · raise 0 (새 함수).

새 함수. 읽은 바이트의 JSON 값, 없거나 깨졌으면 `{}` — 옛 `_ast_value(path)` 의 규칙 그대로. `_ast_value` 는 호출자가 0 이 되어 지웠다(디스크를 읽는 둘째 철자였다).

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1362 | If | `if raw is None:` |
| B2 | 1364 | Try | `try:` |
| B3 | 1366 | ExceptHandler | `except ValueError:` |

## task 7.5.2.1 — 판정이 읽는 입력을 전부 세고 하나씩 묶는다 (7.5.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:1450-1457` · 분기 1 · 반환 1 · raise 0 (편집 전 L1355-1367 · 분기 3 · 반환 3 · raise 0, `ast.before-7.5.2.1.json` = revision `fc35eb2d`).

`_parse_ast` 에 위임 — 못 쓰는 값은 `{}`. 언제나 사전을 돌려준다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1457 | IfExp | `{} if raw is None else _parse_ast(raw)[0]` |

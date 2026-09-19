# Function Logic Map: `_bundle_text` (Python, a122 task 7.5.2)

## task 7.5.2 — 판정은 한 번 읽은 바이트로 선다 (수리한 트리의 재리뷰, 2026-09-19)

`tools/logic-map/check_analysis.py:1312-1343` · 분기 7 · 반환 1 · raise 0 (편집 전 L1217-1240 · 분기 4 · 반환 1 · raise 0, `ast.before-7.5.2.json` = revision `e9f905bd`).

`known`(한 번 읽은 `ast.json`)이 있으면 `ast.json` 은 그 바이트만 쓴다 — 그 뒤에 생긴 것도 안 읽는다. **GREEN 도중 편집 집합에 들어왔다**(review.md).

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1325 | For | `for path in sorted(map_path.parent.glob('*')):` |
| B2 | 1326 | If | `if not path.is_file():` |
| B3 | 1328 | Try | `try:` |
| B4 | 1329 | BoolOp | `known is not None and path.name == 'ast.json'` |
| B5 | 1329 | If | `if known is not None and path.name == 'ast.json':` |
| B6 | 1333 | If | `if raw is None:` |
| B7 | 1338 | ExceptHandler | `except (OSError, UnicodeDecodeError):` |

## task 7.5.2.1 — 판정이 읽는 입력을 전부 세고 하나씩 묶는다 (7.5.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:1378-1409` · 분기 6 · 반환 1 · raise 0 (편집 전 L1312-1343 · 분기 7 · 반환 1 · raise 0, `ast.before-7.5.2.1.json` = revision `fc35eb2d`).

`(대상, 한 번 읽은 ast 바이트)` 를 받는다. `ast.json` 은 목록과 **무관하게** 그 바이트다 — 7.5.2 는 살아 있는 목록에 있을 때만 썼고, 한 번 읽은 뒤 지워지면 판정한 바이트가 빠져 열거형 호출 판정이 꺼졌다(Codex P1 재현: fc35eb2d `[]`). 목록은 `iterdir()` — 못 열면 `OSError` 가 올라가 `check` 가 이름 댄 줄로 만든다(`glob` 은 그것을 삼켰다).

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1396 | comprehension | ` for path in target.iterdir() if path.name != 'ast.json'` |
| B2 | 1397 | If | `if ast_raw is not None:` |
| B3 | 1400 | For | `for name in sorted(paths):` |
| B4 | 1401 | Try | `try:` |
| B5 | 1402 | IfExp | `_decoded(ast_raw) if name == 'ast.json' else paths[name].read_text(encoding='utf-8')` |
| B6 | 1404 | ExceptHandler | `except (OSError, UnicodeDecodeError):` |

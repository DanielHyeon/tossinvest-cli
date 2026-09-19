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

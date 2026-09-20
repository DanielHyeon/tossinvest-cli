# Function Logic Map: `test_index` (Python, a122 task 7.5.2.3)

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:1559-1567` · 분기 3 · 반환 1 · raise 0 (편집 전 L1317-1325 · 분기 3 · 반환 1 · raise 0, `ast.before-7.5.2.3.json` = revision `1d12520c`).

트리 순회를 `_globbed` 로 — `*_test.go` 집합이 원장에 남는다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1562 | For | `for path in _globbed(root, '*_test.go'):` |
| B2 | 1563 | If | `if '.git' in path.parts:` |
| B3 | 1565 | For | `for name, (start, end) in test_spans(path).items():` |

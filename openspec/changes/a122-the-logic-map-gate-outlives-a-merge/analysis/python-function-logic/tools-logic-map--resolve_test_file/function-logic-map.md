# Function Logic Map: `resolve_test_file` (Python, a122 task 7.5.2.3)

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:1570-1590` · 분기 5 · 반환 3 · raise 0 (편집 전 L1328-1346 · 분기 5 · 반환 3 · raise 0, `ast.before-7.5.2.3.json` = revision `1d12520c`).

인용한 파일을 **고르는** 세 갈래(정규화된 경로 · 패키지 안 · 트리 전체)가 전부 깔때기를 지난다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1583 | If | `if '/' in cited:` |
| B2 | 1585 | IfExp | `qualified if _kind(qualified) == 'reg' else None` |
| B3 | 1587 | If | `if _kind(local) == 'reg':` |
| B4 | 1589 | comprehension | ` for path in _globbed(root, cited) if '.git' not in path.parts` |
| B5 | 1590 | IfExp | `matches[0] if len(matches) == 1 else None` |

# Function Logic Map: `_read_evidence` (Python, a122 task 7.5.2)

## task 7.5.2 — 판정은 한 번 읽은 바이트로 선다 (수리한 트리의 재리뷰, 2026-09-19)

`tools/logic-map/check_analysis.py:594-610` · 분기 4 · 반환 1 · raise 0 (새 함수).

새 함수. 증거 디렉터리의 모든 `ast.json` 을 **한 번씩** 읽는 유일한 자리다. 못 읽으면 `None` — 지문에 빈 해시로 남으므로 다음 읽기에서 달라진다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 604 | For | `for ast_path in sorted(analysis.glob('*/ast.json')) if analysis.is_dir() else ():` |
| B2 | 604 | IfExp | `sorted(analysis.glob('*/ast.json')) if analysis.is_dir() else ()` |
| B3 | 605 | Try | `try:` |
| B4 | 607 | ExceptHandler | `except OSError:` |

# Function Logic Map: `_kind_outcome` (Python, a122 task 7.5.2.3)

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:838-845` · 분기 4 · 반환 2 · raise 0 (새 함수).

새 함수. 종류의 지문 — `dir` · `reg` · `other` · 못 물으면 빈 글자.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 840 | Try | `try:` |
| B2 | 842 | ExceptHandler | `except OSError as exc:` |
| B3 | 844 | IfExp | `'dir' if stat.S_ISDIR(mode) else 'reg' if stat.S_ISREG(mode) else 'other'` |
| B4 | 844 | IfExp | `'reg' if stat.S_ISREG(mode) else 'other'` |

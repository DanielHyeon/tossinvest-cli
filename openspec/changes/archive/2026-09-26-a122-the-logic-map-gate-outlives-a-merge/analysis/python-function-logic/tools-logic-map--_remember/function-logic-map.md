# Function Logic Map: `_remember` (Python, a122 task 7.5.2.3)

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:732-741` · 분기 4 · 반환 1 · raise 0 (새 함수).

새 함수. 읽기 결과를 원장에 적는다. 재확인 중에는 적지 않는다(자기 읽기를 자기가 견주게 된다). 한 판에서 같은 경로가 **갈리면** 그 자리에서 움직임이다 — 끝까지 기다리지 않는다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 735 | BoolOp | `book is None or not book.recording` |
| B2 | 735 | If | `if book is None or not book.recording:` |
| B3 | 738 | BoolOp | `previous is not None and previous != outcome and (not book.diverged)` |
| B4 | 738 | If | `if previous is not None and previous != outcome and (not book.diverged):` |

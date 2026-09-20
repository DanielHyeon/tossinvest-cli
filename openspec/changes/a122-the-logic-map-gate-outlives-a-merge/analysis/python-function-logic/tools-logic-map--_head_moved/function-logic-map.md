# Function Logic Map: `_head_moved` (Python, a122 task 7.5.2.3)

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:2203-2208` · 분기 1 · 반환 2 · raise 0 (새 함수).

새 함수. 역사를 다시 묻는 **한 자리**. `_judged_state_moved` 가 앞뒤로 두 번 부른다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 2206 | If | `if now != head:` |

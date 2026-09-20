# Function Logic Map: `_recording_moved` (Python, a122 task 7.5.2.3)

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:2447-2467` · 분기 2 · 반환 2 · raise 0 (새 함수).

새 함수. 쓰기 직전에 **거절 집합 전체**를 다시 묻는다. 순회는 133~219초이고 그 사이의 Go 편집은 역사에도 원장에도 없다 — 7.5.2.2 는 그 상태에서 기록을 **영구히** 만들었다. 순서가 곧 사유다: 역사 → 거절 → 원장.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 2462 | If | `if moved:` |
| B2 | 2467 | BoolOp | `refusal or _judged_state_moved(root, head, book)` |

# Function Logic Map: `_listed` (Python, a122 task 7.5.2.3)

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:815-821` · 분기 1 · 반환 1 · raise 1 (새 함수).

새 함수. 디렉터리를 **여는 유일한 자리**. `iterdir` 가 이 함수 밖에 있으면 그 목록은 원장에 안 남고, 재확인의 집합이 판정의 집합보다 좁아진다 — 7.5.2.2 가 깨진 바로 그 모양이다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 819 | If | `if isinstance(value, OSError):` |

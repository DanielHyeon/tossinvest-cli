# Function Logic Map: `_why` (Python, a122 task 7.5.2.3)

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:871-876` · 분기 2 · 반환 2 · raise 0 (새 함수).

새 함수. 실패를 사람의 말로. 해독 실패의 기본 문장은 바이트 위치까지 담아 길어서 "not UTF-8 text" 로 줄인다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 873 | If | `if isinstance(exc, UnicodeDecodeError):` |
| B2 | 876 | BoolOp | `getattr(exc, 'strerror', '') or exc` |

# Function Logic Map: `_listing_outcome` (Python, a122 task 7.5.2.3)

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:805-812` · 분기 5 · 반환 2 · raise 0 (새 함수).

새 함수. 디렉터리 목록의 지문 — 이름과 **종류**를 같이 넣는다(같은 이름의 파일↔폴더 교체도 변화다).

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 807 | Try | `try:` |
| B2 | 808 | comprehension | ` for child in path.iterdir()` |
| B3 | 809 | ExceptHandler | `except OSError as exc:` |
| B4 | 811 | IfExp | `'d' if is_dir else 'f'` |
| B5 | 811 | comprehension | ` for name, is_dir in entries` |

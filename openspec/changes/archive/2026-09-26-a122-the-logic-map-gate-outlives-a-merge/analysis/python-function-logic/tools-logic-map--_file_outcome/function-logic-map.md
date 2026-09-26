# Function Logic Map: `_file_outcome` (Python, a122 task 7.5.2.3)

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:779-789` · 분기 2 · 반환 2 · raise 0 (새 함수).

새 함수. `(지문, 바이트 또는 실패)` — 파일 읽기의 지문을 만드는 **한 자리**다. 판정의 읽기와 끝의 재확인이 같은 함수를 부르므로 둘이 갈릴 수 없다([[two-judgements-cover-for-each-other]]). 실패도 지문이 된다(`이름:errno`) — 그래야 "여는 순간에만 FIFO" 와 "이름을 뺐다 되돌리기" 가 끝에서 다르게 보인다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 785 | Try | `try:` |
| B2 | 787 | ExceptHandler | `except OSError as exc:` |

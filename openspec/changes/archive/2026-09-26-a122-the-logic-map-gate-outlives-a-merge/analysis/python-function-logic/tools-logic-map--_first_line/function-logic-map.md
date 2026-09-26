# Function Logic Map: `_first_line` (Python, a122 task 7.5.2.1)

## task 7.5.2.1 — 판정이 읽는 입력을 전부 세고 하나씩 묶는다 (7.5.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:541-545` · 분기 3 · 반환 1 · raise 0 (새 함수).

새 함수. git 이 한 말의 첫 줄(160자) — 결함 문장이 git 의 말로 이름을 대게(7.5.1 F6). 새 결함 넷과 `_head_commit` 이 같이 쓴다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 543 | BoolOp | `said or ''` |
| B2 | 543 | IfExp | `said.decode('utf-8', 'replace') if isinstance(said, bytes) else said or ''` |
| B3 | 545 | IfExp | `lines[0][:160] if lines else fallback` |

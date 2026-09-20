# Function Logic Map: `test_spans` (Python, a122 task 7.5.2.3)

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:1534-1556` · 분기 7 · 반환 2 · raise 0 (편집 전 L1295-1314 · 분기 7 · 반환 2 · raise 0, `ast.before-7.5.2.3.json` = revision `1d12520c`).

시험 파일을 깔때기로 읽는다(옛 `read_text` 는 그 자리의 FIFO 에 영원히 멎었다). 글자가 아닌 바이트는 옛 판본과 같이 `errors="replace"` 로 견딘다 — 이 판정은 Go 시험 선언만 찾는다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1537 | Try | `try:` |
| B2 | 1542 | ExceptHandler | `except OSError:` |
| B3 | 1544 | For | `for index, line in enumerate(lines):` |
| B4 | 1546 | If | `if not declared:` |
| B5 | 1549 | For | `for offset in range(index, len(lines)):` |
| B6 | 1551 | BoolOp | `depth <= 0 and offset > index` |
| B7 | 1551 | If | `if depth <= 0 and offset > index:` |

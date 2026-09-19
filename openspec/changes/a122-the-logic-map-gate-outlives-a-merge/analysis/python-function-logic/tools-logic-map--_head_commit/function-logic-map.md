# Function Logic Map: `_head_commit` (Python, a122 task 7.5.2)

## task 7.5.2 — 판정은 한 번 읽은 바이트로 선다 (수리한 트리의 재리뷰, 2026-09-19)

`tools/logic-map/check_analysis.py:1697-1706` · 분기 2 · 반환 1 · raise 1 (새 함수).

새 함수. 지문의 `HEAD`. 못 읽으면 결함 — 빈 값이면 두 번 다 못 읽은 실행이 "그대로" 가 된다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1703 | If | `if process.returncode:` |
| B2 | 1705 | IfExp | `said[0] if said else 'git rev-parse failed'` |

| raise 줄 | 소스 |
|---|---|
| 1705 | `raise RuntimeError(f'cannot read HEAD: {(said[0] if said else 'git rev-parse failed')}')` |

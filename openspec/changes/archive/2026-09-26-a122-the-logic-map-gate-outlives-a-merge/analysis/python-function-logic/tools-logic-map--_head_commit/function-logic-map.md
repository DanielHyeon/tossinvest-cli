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

## task 7.5.2.1 — 판정이 읽는 입력을 전부 세고 하나씩 묶는다 (7.5.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:1783-1794` · 분기 1 · 반환 1 · raise 1 (편집 전 L1697-1706 · 분기 2 · 반환 1 · raise 1, `ast.before-7.5.2.1.json` = revision `fc35eb2d`).

명령마다 **한 번** 부르는 고정 자리가 됐다(7.5.2 는 지문 표본으로 두 번). 실패 문장은 `_first_line`.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1792 | If | `if process.returncode:` |

| raise 줄 | 소스 |
|---|---|
| 1793 | `raise RuntimeError(f'cannot read HEAD: {_first_line(process.stderr, 'git rev-parse failed'` |

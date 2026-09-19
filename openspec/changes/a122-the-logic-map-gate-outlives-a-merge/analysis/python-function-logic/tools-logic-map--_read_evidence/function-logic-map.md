# Function Logic Map: `_read_evidence` (Python, a122 task 7.5.2)

## task 7.5.2 — 판정은 한 번 읽은 바이트로 선다 (수리한 트리의 재리뷰, 2026-09-19)

`tools/logic-map/check_analysis.py:594-610` · 분기 4 · 반환 1 · raise 0 (새 함수).

새 함수. 증거 디렉터리의 모든 `ast.json` 을 **한 번씩** 읽는 유일한 자리다. 못 읽으면 `None` — 지문에 빈 해시로 남으므로 다음 읽기에서 달라진다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 604 | For | `for ast_path in sorted(analysis.glob('*/ast.json')) if analysis.is_dir() else ():` |
| B2 | 604 | IfExp | `sorted(analysis.glob('*/ast.json')) if analysis.is_dir() else ()` |
| B3 | 605 | Try | `try:` |
| B4 | 607 | ExceptHandler | `except OSError:` |

## task 7.5.2.1 — 판정이 읽는 입력을 전부 세고 하나씩 묶는다 (7.5.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:646-670` · 분기 6 · 반환 2 · raise 0 (편집 전 L594-610 · 분기 4 · 반환 1 · raise 0, `ast.before-7.5.2.1.json` = revision `fc35eb2d`).

증거 디렉터리를 **한 번** 읽어 `Evidence`(디렉터리 · 있었나 · 번들 목록 · `ast.json` 바이트)로 돌려준다. 번들은 `iterdir` 로 세고 `ast.json` 은 **직접** 연다 — `glob("*/ast.json")` 은 목록(r)이 안 되는 번들을 조용히 건너뛰어 멀쩡히 읽히는 증거를 "없다" 로 만들었다(Codex P2 재현). 없음(`FileNotFoundError`)은 키가 없고 못 읽음은 `None` — 판정 줄이 다르다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 658 | If | `if not analysis.is_dir():` |
| B2 | 660 | comprehension | ` for path in analysis.iterdir() if path.is_dir()` |
| B3 | 662 | For | `for target in targets:` |
| B4 | 664 | Try | `try:` |
| B5 | 666 | ExceptHandler | `except FileNotFoundError:` |
| B6 | 668 | ExceptHandler | `except OSError:` |

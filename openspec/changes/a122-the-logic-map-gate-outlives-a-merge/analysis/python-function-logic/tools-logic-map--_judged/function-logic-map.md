# Function Logic Map: `_judged` (Python, a122 task 7.5.2.3)

## task 7.5.2.3 — 재확인의 입력 집합은 판정의 입력 집합이다 (7.5.2.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:1940-2066` · 분기 32 · 반환 13 · raise 0 (새 함수).

`check` 의 판정 본문. 읽는 자리가 전부 깔때기로 바뀌었다(`review.md` · `function-logic-reference.txt` · 증거 · 공존 판정). 증거를 읽는 자리는 **하나**로 남았다 — 빌리는 change 면 그 앞에서 `analysis` 가 빌려주는 쪽으로 바뀐다. 못 읽는 `review.md` · 참조 파일은 이제 판정 줄이다(옛 판본은 FIFO 에 멎었다).

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1957 | Try | `try:` |
| B2 | 1959 | ExceptHandler | `except ValueError as exc:` |
| B3 | 1967 | Try | `try:` |
| B4 | 1969 | ExceptHandler | `except FileNotFoundError:` |
| B5 | 1971 | ExceptHandler | `except (OSError, UnicodeDecodeError) as exc:` |
| B6 | 1973 | Try | `try:` |
| B7 | 1981 | ExceptHandler | `except GATE_FAULTS as exc:` |
| B8 | 1988 | Try | `try:` |
| B9 | 1990 | ExceptHandler | `except FileNotFoundError:` |
| B10 | 1992 | ExceptHandler | `except OSError as exc:` |
| B11 | 1994 | If | `if reference_raw is not None:` |
| B12 | 1997 | Try | `try:` |
| B13 | 1998 | comprehension | ` for name, is_dir in _listed(analysis) if is_dir` |
| B14 | 1999 | ExceptHandler | `except (FileNotFoundError, NotADirectoryError):` |
| B15 | 2001 | If | `if any((_listed(bundle) for bundle in local)):` |
| B16 | 2001 | comprehension | ` for bundle in local` |
| B17 | 2005 | BoolOp | `not re.fullmatch('[a-z0-9][a-z0-9-]*', referenced_change) or referenced_change == change` |
| B18 | 2005 | If | `if not re.fullmatch('[a-z0-9][a-z0-9-]*', referenced_change) or referenced_change == change:` |
| B19 | 2007 | Try | `try:` |
| B20 | 2010 | ExceptHandler | `except GATE_FAULTS as exc:` |
| B21 | 2012 | If | `if referenced_base != base:` |
| B22 | 2022 | If | `if _landing_record(change_dir, root, head) is not None:` |
| B23 | 2029 | Try | `try:` |
| B24 | 2031 | ExceptHandler | `except GATE_FAULTS as exc:` |
| B25 | 2034 | BoolOp | `adopted and _landing_record(change_dir, root, head) is not None` |
| B26 | 2034 | If | `if adopted and _landing_record(change_dir, root, head) is not None:` |
| B27 | 2040 | Try | `try:` |
| B28 | 2048 | IfExp | `str(facts.get('adoption_source', '')) if adopted else resolve_landing(change_dir, root, base, head, evidence)` |
| B29 | 2051 | ExceptHandler | `except GATE_FAULTS as exc:` |
| B30 | 2055 | If | `if not landing:` |
| B31 | 2060 | Try | `try:` |
| B32 | 2062 | ExceptHandler | `except GATE_FAULTS as exc:` |

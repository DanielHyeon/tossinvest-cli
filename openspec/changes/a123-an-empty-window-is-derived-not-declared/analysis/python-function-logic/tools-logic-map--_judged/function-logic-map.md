# Python Function Logic Map: `_judged` (편집 전)

- Source: `tools/logic-map/check_analysis.py` @ base `77e36cca22498ff5e022c0392f4166cec3260c7d`
- 열거: `python3 openspec/changes/archive/2026-09-26-a122-the-logic-map-gate-outlives-a-merge/analysis/python-function-logic/enumerate.py tools/logic-map/check_analysis.py _judged <base>` → `ast.before.json`
- source_sha256 `6aa2dfd3fa48662882fa81536a3583fa8a1bdf73a9ab14c3464ca7bd3652274d` · 줄 2772–2900 · 분기 32 · 반환 13 · raise 0

## a123 관점

**편집 대상.** B27 try 블록(착지 해소 → `changed_existing_functions`) 사이에 유도 호출 한 자리를 넣는다. 조건: `not adopted`, `not landing`, 빌림 아님(`reference_raw is None`). 유도 결과가 R1 이면 `required = {}` · 검증 기준 = base, R2 면 `required` 를 귀속 키로 거른다(부분집합 — 더 요구하지 않는다). 나머지 32 분기·13 반환은 불변. B30(`if not landing`) 의 base 모양 조언은 유도가 선 경우 건너뛴다(조언이 유도와 모순되지 않게).

## 분기 (AST 열거 — 손으로 고르지 않음)

| ID | 줄 | 종류 | 원문 |
|---|---:|---|---|
| B1 | 2790 | Try | `try:` |
| B2 | 2792 | ExceptHandler | `except ValueError as exc:` |
| B3 | 2800 | Try | `try:` |
| B4 | 2802 | ExceptHandler | `except FileNotFoundError:` |
| B5 | 2804 | ExceptHandler | `except (OSError, UnicodeDecodeError) as exc:` |
| B6 | 2806 | Try | `try:` |
| B7 | 2815 | ExceptHandler | `except GATE_FAULTS as exc:` |
| B8 | 2822 | Try | `try:` |
| B9 | 2824 | ExceptHandler | `except FileNotFoundError:` |
| B10 | 2826 | ExceptHandler | `except OSError as exc:` |
| B11 | 2828 | If | `if reference_raw is not None:` |
| B12 | 2831 | Try | `try:` |
| B13 | 2832 | comprehension | ` for name, is_dir in _listed(analysis) if is_dir` |
| B14 | 2833 | ExceptHandler | `except (FileNotFoundError, NotADirectoryError):` |
| B15 | 2835 | If | `if any((_listed(bundle) for bundle in local)):` |
| B16 | 2835 | comprehension | ` for bundle in local` |
| B17 | 2839 | BoolOp | `not re.fullmatch('[a-z0-9][a-z0-9-]*', referenced_change) or referenced_change == change` |
| B18 | 2839 | If | `if not re.fullmatch('[a-z0-9][a-z0-9-]*', referenced_change) or referenced_change == chang` |
| B19 | 2841 | Try | `try:` |
| B20 | 2844 | ExceptHandler | `except GATE_FAULTS as exc:` |
| B21 | 2846 | If | `if referenced_base != base:` |
| B22 | 2856 | If | `if _landing_record(change_dir, root, head) is not None:` |
| B23 | 2863 | Try | `try:` |
| B24 | 2865 | ExceptHandler | `except GATE_FAULTS as exc:` |
| B25 | 2868 | BoolOp | `adopted and _landing_record(change_dir, root, head) is not None` |
| B26 | 2868 | If | `if adopted and _landing_record(change_dir, root, head) is not None:` |
| B27 | 2874 | Try | `try:` |
| B28 | 2882 | IfExp | `str(facts.get('adoption_source', '')) if adopted else resolve_landing(change_dir, root, ba` |
| B29 | 2885 | ExceptHandler | `except GATE_FAULTS as exc:` |
| B30 | 2889 | If | `if not landing:` |
| B31 | 2894 | Try | `try:` |
| B32 | 2896 | ExceptHandler | `except GATE_FAULTS as exc:` |

## 반환

- L2793: `([str(exc)], False)`
- L2805: `([UNREADABLE.format(what='review.md', why=_why(exc))], False)`
- L2816: `([f'cannot derive modified Go functions: {exc}'], False)`
- L2827: `([UNREADABLE.format(what='function-logic-reference.txt', why=_why(exc))], False)`
- L2836: `(['function-logic reference cannot coexist with local function-logic evidence'], False)`
- L2840: `(['function-logic reference names an invalid or recursive change'], False)`
- L2845: `([f'function-logic reference base is invalid: {exc}'], False)`
- L2847: `(['function-logic reference must share the exact comparison base'], False)`
- L2857: `([BORROWED_REFUSES_A_LANDING], False)`
- L2866: `([f'cannot derive modified Go functions: {exc}'], False)`
- L2873: `([ADOPTION_REFUSES_A_LANDING], False)`
- L2886: `([f'cannot derive modified Go functions: {exc}'], False)`
- L2900: `(_verdict(root, base, landing, adopted, required, evidence, review_text), True)`

## raise

- (없음)

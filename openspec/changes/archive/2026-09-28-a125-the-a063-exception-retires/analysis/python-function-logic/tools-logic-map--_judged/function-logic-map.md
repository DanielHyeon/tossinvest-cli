# Python Function Logic Map: `_judged` (편집 전)

- Source `tools/logic-map/check_analysis.py` @ base · sha256 `6aa2dfd3fa48662882fa81536a3583fa8a1bdf73a9ab14c3464ca7bd3652274d` · L2772–2900
- 분기 32 · 반환 13 · raise 0 — 열거: a122 `enumerate.py`(기계, 손으로 고르지 않음)

## a125 편집 지점 (이관 관련 분기 — AST 원문 정규식 `(?i)adopt|audited|execution_baseline`)

- 분기: B25(L2868), B26(L2868), B28(L2882)
- 이관 값 반환: L2873, L2900
- 처분: 이관 분기·인자·문맥 키를 지운다. 나머지 분기는 불변 — 편집 후 재추출(3.3)에서 분기 수 차이가 이 목록과 같아야 한다.

## 분기 전수

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
| B18 | 2839 | If | `if not re.fullmatch('[a-z0-9][a-z0-9-]*', referenced_change) or referenced_change == change:` |
| B19 | 2841 | Try | `try:` |
| B20 | 2844 | ExceptHandler | `except GATE_FAULTS as exc:` |
| B21 | 2846 | If | `if referenced_base != base:` |
| B22 | 2856 | If | `if _landing_record(change_dir, root, head) is not None:` |
| B23 | 2863 | Try | `try:` |
| B24 | 2865 | ExceptHandler | `except GATE_FAULTS as exc:` |
| B25 | 2868 | BoolOp | `adopted and _landing_record(change_dir, root, head) is not None` |
| B26 | 2868 | If | `if adopted and _landing_record(change_dir, root, head) is not None:` |
| B27 | 2874 | Try | `try:` |
| B28 | 2882 | IfExp | `str(facts.get('adoption_source', '')) if adopted else resolve_landing(change_dir, root, base, head, ` |
| B29 | 2885 | ExceptHandler | `except GATE_FAULTS as exc:` |
| B30 | 2889 | If | `if not landing:` |
| B31 | 2894 | Try | `try:` |
| B32 | 2896 | ExceptHandler | `except GATE_FAULTS as exc:` |

## 편집 후 재추출 (a125 3.3, `9e63b681`)

- `ast.after.json` — 분기 32→29 · 반환 13→12 · raise 0→0
- 분기 · 반환 원문 대조(`difflib.ndiff`, 위치가 아니라 원문으로 — 분기 번호는 위치다):

    - BoolOp adopted and _landing_record(change_dir, root, head) is not None
    - If if adopted and _landing_record(change_dir, root, head) is not None:
    - IfExp str(facts.get('adoption_source', '')) if adopted else resolve_landing(change_dir, root, base, head, evidence)
    - ([ADOPTION_REFUSES_A_LANDING], False)
    - (_verdict(root, base, landing, adopted, required, evidence, review_text), True)
    + (_verdict(root, base, landing, required, evidence, review_text), True)
- 재추출 갱신: `5440edad`(리뷰 P2 의 docstring 두 자리 수정 뒤) — 분기 · 반환 원문이 `9e63b681` 판과 같다. `ast.after.json` 은 `5440edad` 판이다.

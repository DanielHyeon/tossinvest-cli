# Python Function Logic Map: `_verdict` (편집 전)

- Source `tools/logic-map/check_analysis.py` @ base · sha256 `6aa2dfd3fa48662882fa81536a3583fa8a1bdf73a9ab14c3464ca7bd3652274d` · L2903–3000
- 분기 26 · 반환 4 · raise 0 — 열거: a122 `enumerate.py`(기계, 손으로 고르지 않음)

## a125 편집 지점 (이관 관련 분기 — AST 원문 정규식 `(?i)adopt|audited|execution_baseline`)

- 분기: 없음(분기 없음 — 인자·문자열·호출만)
- 이관 값 반환: L2917
- 처분: 이관 분기·인자·문맥 키를 지운다. 나머지 분기는 불변 — 편집 후 재추출(3.3)에서 분기 수 차이가 이 목록과 같아야 한다.

## 분기 전수

| ID | 줄 | 종류 | 원문 |
|---|---:|---|---|
| B1 | 2912 | If | `if not evidence.present:` |
| B2 | 2913 | If | `if required:` |
| B3 | 2914 | comprehension | ` for source, function in required` |
| B4 | 2921 | IfExp | `[] if EXEMPTION in review_text else [f'missing analysis or `{EXEMPTION}` review marker']` |
| B5 | 2923 | If | `if not evidence.targets:` |
| B6 | 2938 | For | `for target in evidence.targets:` |
| B7 | 2941 | If | `if target in evidence.unlistable:` |
| B8 | 2946 | If | `if 'function-logic-map.md' not in names:` |
| B9 | 2948 | Try | `try:` |
| B10 | 2951 | ExceptHandler | `except OSError as exc:` |
| B11 | 2956 | IfExp | `exc.filename if isinstance(exc.filename, str) else ''` |
| B12 | 2957 | IfExp | `Path(named).name if named else target.name` |
| B13 | 2959 | comprehension | ` for target, text in bundle_texts.items()` |
| B14 | 2972 | If | `if landing:` |
| B15 | 2973 | Try | `try:` |
| B16 | 2974 | comprehension | ` for _, source, _ in _select_pinning(root, evidence)` |
| B17 | 2975 | ExceptHandler | `except ValueError:` |
| B18 | 2978 | For | `for target in evidence.targets:` |
| B19 | 2983 | If | `if binding:` |
| B20 | 2984 | If | `if binding in covered:` |
| B21 | 2987 | For | `for binding, expected in required.items():` |
| B22 | 2989 | If | `if target is None:` |
| B23 | 2994 | BoolOp | `expected.get('current_hash') or expected.get('base_hash')` |
| B24 | 2995 | If | `if ast_value.get('source_sha256') != expected_hash:` |
| B25 | 2997 | IfExp | `'current' if expected.get('current_hash') else 'base'` |
| B26 | 2998 | If | `if ast_value.get('revision', 'current') != expected_revision:` |

## 편집 후 재추출 (a125 3.3, `9e63b681`)

- `ast.after.json` — 분기 26→26 · 반환 4→4 · raise 0→0
- 분기 · 반환 원문 대조(`difflib.ndiff`, 위치가 아니라 원문으로 — 분기 번호는 위치다):

    - [f'missing Function Logic Map for {len(required)} function(s) modified between base {base[:12]} and {_target_text(landing, adopted)}: {nam
    + [f'missing Function Logic Map for {len(required)} function(s) modified between base {base[:12]} and {_target_text(landing)}: {names}']
- 재추출 갱신: `5440edad`(리뷰 P2 의 docstring 두 자리 수정 뒤) — 분기 · 반환 원문이 `9e63b681` 판과 같다. `ast.after.json` 은 `5440edad` 판이다.

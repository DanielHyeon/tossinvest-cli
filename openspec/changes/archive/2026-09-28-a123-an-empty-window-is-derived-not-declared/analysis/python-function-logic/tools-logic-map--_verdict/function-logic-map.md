# Python Function Logic Map: `_verdict` (편집 전)

- Source: `tools/logic-map/check_analysis.py` @ base `77e36cca22498ff5e022c0392f4166cec3260c7d`
- 열거: `python3 openspec/changes/archive/2026-09-26-a122-the-logic-map-gate-outlives-a-merge/analysis/python-function-logic/enumerate.py tools/logic-map/check_analysis.py _verdict <base>` → `ast.before.json`
- source_sha256 `6aa2dfd3fa48662882fa81536a3583fa8a1bdf73a9ab14c3464ca7bd3652274d` · 줄 2903–3000 · 분기 26 · 반환 4 · raise 0

## a123 관점

**편집 안 함.** 호출자가 넘기는 `landing` 인자가 R1 에서 base 가 된다 — 그 값은 prefetch(`_committed_many(root, landing, …)`)와 `validate_target(revision_ref=landing)` 에 쓰여 `revision: current` 해시를 base 에서 대조한다. R1 조건 1 이 그 등식을 이미 요구하므로 R1 에서 이 갈래는 번들 유효성(산문·분기·시험 인용)만 새로 판정한다. `evidence.present` 거짓 갈래(B1~)는 R1 에서 도달 불가(조건 1 이 번들 ≥1 요구).

## 분기 (AST 열거 — 손으로 고르지 않음)

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

## 반환

- L2917: `[f'missing Function Logic Map for {len(required)} function(s) modified between base {base[:12]} and `
- L2921: `[] if EXEMPTION in review_text else [f'missing analysis or `{EXEMPTION}` review marker']`
- L2924: `['function-logic analysis directory has no targets']`
- L3000: `errors`

## raise

- (없음)

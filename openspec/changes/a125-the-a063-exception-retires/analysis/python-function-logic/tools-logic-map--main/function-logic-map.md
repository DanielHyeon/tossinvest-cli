# Python Function Logic Map: `main` (편집 전)

- Source `tools/logic-map/check_analysis.py` @ base · sha256 `6aa2dfd3fa48662882fa81536a3583fa8a1bdf73a9ab14c3464ca7bd3652274d` · L3399–3519
- 분기 22 · 반환 3 · raise 0 — 열거: a122 `enumerate.py`(기계, 손으로 고르지 않음)

## a125 편집 지점 (이관 관련 분기 — AST 원문 정규식 `(?i)adopt|audited|execution_baseline`)

- 분기: B22(L3513)
- 이관 값 반환: 없음
- 처분: 이관 분기·인자·문맥 키를 지운다. 나머지 분기는 불변 — 편집 후 재추출(3.3)에서 분기 수 차이가 이 목록과 같아야 한다.

## 분기 전수

| ID | 줄 | 종류 | 원문 |
|---|---:|---|---|
| B1 | 3409 | If | `if reconfigure is not None:` |
| B2 | 3416 | If | `if args.record_landing:` |
| B3 | 3417 | Try | `try:` |
| B4 | 3419 | ExceptHandler | `except GATE_FAULTS as exc:` |
| B5 | 3421 | For | `for line in lines:` |
| B6 | 3428 | Try | `try:` |
| B7 | 3430 | ExceptHandler | `except GATE_FAULTS as exc:` |
| B8 | 3441 | BoolOp | `base and 'landing' in context` |
| B9 | 3441 | If | `if base and 'landing' in context:` |
| B10 | 3452 | If | `if not landing:` |
| B11 | 3457 | BoolOp | `landed_after or '?'` |
| B12 | 3463 | BoolOp | `context.get('base_shaped_bundles') or []` |
| B13 | 3463 | comprehension | ` for name in context.get('base_shaped_bundles') or []` |
| B14 | 3465 | If | `if fault:` |
| B15 | 3469 | If | `if base_shaped:` |
| B16 | 3473 | If | `if len(base_shaped) > 3:` |
| B17 | 3487 | Try | `try:` |
| B18 | 3495 | ExceptHandler | `except GATE_FAULTS as exc:` |
| B19 | 3497 | If | `if refusal:` |
| B20 | 3509 | If | `if errors:` |
| B21 | 3510 | For | `for error in errors:` |
| B22 | 3513 | If | `if context.get('execution_baseline_adoption'):` |

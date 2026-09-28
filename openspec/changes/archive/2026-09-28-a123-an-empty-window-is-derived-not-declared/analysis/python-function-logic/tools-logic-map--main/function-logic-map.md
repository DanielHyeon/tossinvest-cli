# Python Function Logic Map: `main` (편집 전)

- Source: `tools/logic-map/check_analysis.py` @ base `77e36cca22498ff5e022c0392f4166cec3260c7d`
- 열거: `python3 openspec/changes/archive/2026-09-26-a122-the-logic-map-gate-outlives-a-merge/analysis/python-function-logic/enumerate.py tools/logic-map/check_analysis.py main <base>` → `ast.before.json`
- source_sha256 `6aa2dfd3fa48662882fa81536a3583fa8a1bdf73a9ab14c3464ca7bd3652274d` · 줄 3399–3519 · 분기 22 · 반환 3 · raise 0

## a123 관점

**편집 대상.** B9 창 줄의 대상 텍스트와 B10(`if not landing`) 조언 갈래. 유도 사실(`derived_window`)이 있으면 창 줄 대상에 그 문장을 쓰고 워킹트리 조언(B10 이하 B11~B19)은 내지 않는다. B2~B5(기록 명령) · B20~B22(판정 줄 · 성공 줄) 불변.

## 분기 (AST 열거 — 손으로 고르지 않음)

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

## 반환

- L3423: `code`
- L3512: `1`
- L3519: `0`

## raise

- (없음)

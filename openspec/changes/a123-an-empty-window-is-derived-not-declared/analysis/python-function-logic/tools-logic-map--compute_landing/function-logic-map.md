# Python Function Logic Map: `compute_landing` (편집 전)

- Source: `tools/logic-map/check_analysis.py` @ base `77e36cca22498ff5e022c0392f4166cec3260c7d`
- 열거: `python3 openspec/changes/archive/2026-09-26-a122-the-logic-map-gate-outlives-a-merge/analysis/python-function-logic/enumerate.py tools/logic-map/check_analysis.py compute_landing <base>` → `ast.before.json`
- source_sha256 `6aa2dfd3fa48662882fa81536a3583fa8a1bdf73a9ab14c3464ca7bd3652274d` · 줄 3139–3198 · 분기 8 · 반환 4 · raise 1

## a123 관점

**편집 안 함.** 착지 계산 불변. `--record-landing` 은 R1 change 에서도 오늘처럼 거절한다(기록은 착지 규칙의 것).

## 분기 (AST 열거 — 손으로 고르지 않음)

| ID | 줄 | 종류 | 원문 |
|---|---:|---|---|
| B1 | 3154 | If | `if inputs.why:` |
| B2 | 3157 | IfExp | `floor if _is_ancestor(root, base, floor) else base` |
| B3 | 3162 | If | `if process.returncode:` |
| B4 | 3169 | For | `for candidate in [start, *process.stdout.split()]:` |
| B5 | 3172 | If | `if not refusal:` |
| B6 | 3181 | BoolOp | `first or refusal` |
| B7 | 3182 | If | `if names:` |
| B8 | 3185 | If | `if unheld:` |

## 반환

- L3155: `('', inputs.why)`
- L3177: `(candidate, '')`
- L3186: `('', 'no commit holds the evidence this verdict read (' + ', '.join(unheld) + ') — commit the bundle`
- L3195: `('', f'no commit at or after the evidence ({floor[:12]}) is accepted as the landing — at the first c`

## raise

- L3163: `raise RuntimeError(f'cannot walk the history after {start[:12]}: ' + _first_line(process.stderr, 'gi`

# Python Function Logic Map: `check` (편집 전)

- Source: `tools/logic-map/check_analysis.py` @ base `77e36cca22498ff5e022c0392f4166cec3260c7d`
- 열거: `python3 openspec/changes/archive/2026-09-26-a122-the-logic-map-gate-outlives-a-merge/analysis/python-function-logic/enumerate.py tools/logic-map/check_analysis.py check <base>` → `ast.before.json`
- source_sha256 `6aa2dfd3fa48662882fa81536a3583fa8a1bdf73a9ab14c3464ca7bd3652274d` · 줄 2727–2769 · 분기 3 · 반환 2 · raise 0

## a123 관점

**편집 안 함.** 원장 열기·재확인 래퍼. 새 유도 함수가 파이썬으로 디스크를 읽지 않고(git 자식 프로세스만) `Evidence` 를 받으므로 원장 재확인 범위 불변.

## 분기 (AST 열거 — 손으로 고르지 않음)

| ID | 줄 | 종류 | 원문 |
|---|---:|---|---|
| B1 | 2757 | IfExp | `{} if context is None else context` |
| B2 | 2765 | If | `if not judged:` |
| B3 | 2769 | IfExp | `[moved] if moved else verdict` |

## 반환

- L2766: `verdict`
- L2769: `[moved] if moved else verdict`

## raise

- (없음)

# Python Function Logic Map: `changed_existing_functions` (편집 전)

- Source: `tools/logic-map/check_analysis.py` @ base `77e36cca22498ff5e022c0392f4166cec3260c7d`
- 열거: `python3 openspec/changes/archive/2026-09-26-a122-the-logic-map-gate-outlives-a-merge/analysis/python-function-logic/enumerate.py tools/logic-map/check_analysis.py changed_existing_functions <base>` → `ast.before.json`
- source_sha256 `6aa2dfd3fa48662882fa81536a3583fa8a1bdf73a9ab14c3464ca7bd3652274d` · 줄 586–596 · 분기 1 · 반환 1 · raise 1

## a123 관점

**편집 안 함 — 재사용.** R2 는 귀속 커밋 c 마다 `changed_existing_functions(root, c^1, c)` 의 키를 모은다(선례: a067 0de27fab 커밋 메시지의 `changed_existing_functions(75cc736f^, 75cc736f) = 0`). 창 쪽 요구는 오늘처럼 `(base, "")` 이고 R2 는 그 결과를 키로 거른다.

## 분기 (AST 열거 — 손으로 고르지 않음)

| ID | 줄 | 종류 | 원문 |
|---|---:|---|---|
| B1 | 591 | If | `if not base:` |

## 반환

- L596: `_changed_existing_functions(root, base, comparison)`

## raise

- L592: `raise ValueError('Function Logic Map comparison base is required')`

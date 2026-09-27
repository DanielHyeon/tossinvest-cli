# Python Function Logic Map: `_landing_refusal` (편집 전)

- Source: `tools/logic-map/check_analysis.py` @ base `77e36cca22498ff5e022c0392f4166cec3260c7d`
- 열거: `python3 openspec/changes/archive/2026-09-26-a122-the-logic-map-gate-outlives-a-merge/analysis/python-function-logic/enumerate.py tools/logic-map/check_analysis.py _landing_refusal <base>` → `ast.before.json`
- source_sha256 `6aa2dfd3fa48662882fa81536a3583fa8a1bdf73a9ab14c3464ca7bd3652274d` · 줄 2047–2165 · 분기 11 · 반환 9 · raise 0

## a123 관점

**편집 안 함.** 규칙 1~8(K2 포함) 불변 — proposal Non-goals. a123 은 이 함수를 부르지 않는다: 유도는 착지가 **없을 때만**(`landing == ""`) 선다. 규칙 7(B: `all(at_base == at_candidate)`) 이 모양 B 를 거절하는 자리이고, R1 은 그 거절을 뒤집지 않고 별도 경로로 창 끝을 base 로 둔다.

## 분기 (AST 열거 — 손으로 고르지 않음)

| ID | 줄 | 종류 | 원문 |
|---|---:|---|---|
| B1 | 2076 | If | `if not _is_ancestor(root, base, candidate):` |
| B2 | 2079 | If | `if not pinning:` |
| B3 | 2088 | If | `if mismatched:` |
| B4 | 2099 | If | `if not floor:` |
| B5 | 2106 | If | `if not _is_ancestor(root, floor, candidate):` |
| B6 | 2116 | If | `if unheld:` |
| B7 | 2135 | comprehension | ` for _, source, _ in bundles` |
| B8 | 2142 | If | `if all((at_base[source] == at_candidate[source] for source in sources)):` |
| B9 | 2142 | comprehension | ` for source in sources` |
| B10 | 2143 | IfExp | `f', and {len(sources) - 3} more' if len(sources) > 3 else ''` |
| B11 | 2158 | If | `if later:` |

## 반환

- L2077: `(f'landing point precedes the comparison base {base[:12]}: {candidate}', [])`
- L2084: `(f'landing point {candidate[:12]} is pinned by no `revision: current` evidence: a declared landing m`
- L2089: `(f'landing point {candidate[:12]} is not the revision this evidence describes: ' + ', '.join(sorted(`
- L2101: `(f'landing point {candidate[:12]} is pinned by evidence that no ordinary commit on this history adds`
- L2107: `(f"landing point {candidate[:12]} precedes the evidence that pins it ({floor[:12]}): a declared land`
- L2117: `(f'landing point {candidate[:12]} does not hold the evidence this verdict read: ' + ', '.join(unheld`
- L2144: `(f"landing point {candidate[:12]} changes none of the sources its evidence pins since the comparison`
- L2159: `(f"landing point {candidate[:12]} is followed by {len(later)} later commit(s) of this change's own G`
- L2165: `('', [])`

## raise

- (없음)

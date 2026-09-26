# Function Logic Map: `_named` (Python, a122 task 7.5.9 보수 — 새 함수)

## 새 함수 (2026-09-26)

편집 후 `tools/logic-map/check_analysis.py:1462-1469` · 분기 1 · 반환 1 · raise 1 · 호출 4 (`ast.after-759r.json`, source sha `ff590b79db6e`)

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1467 | If | `if isinstance(value, OSError):` |

이름만 필요한 목록의 **유일한 자리**(아카이브 id 고르기). 결과를 원장 `names` 에 적고 실패면 올린다. 재확인은 `_PROBE_NOW["names"]` 로 다시 읽는다(시험 `test_an_archived_copy_that_appears_while_judged_asks_for_a_rerun` · 변이 MR4).

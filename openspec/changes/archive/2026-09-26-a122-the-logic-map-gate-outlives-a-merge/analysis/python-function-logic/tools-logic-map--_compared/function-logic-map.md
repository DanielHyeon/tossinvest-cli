# Function Logic Map: `_compared` (Python, a122 task 7.5.25 — 새 함수)

`ast.after-7.5.25.json` — 분기 1(`IfExp`) · 반환 1. 대상이 커밋이면 `[base, target]`(트리 둘), 워킹트리면
`["--cached", base]`(base 대 스냅숏 인덱스). 가드(`_safe_changed_go_paths`)와 판정(`_changed_existing_functions`)이
**이 함수 하나**에서 쌍을 받는다 — 구조 시험 `test_both_git_calls_see_the_same_files` 가 두 호출의 `*_compared` 와
`env=environment` 를 못 박는다.

## task 7.5.34 — 지웠다 (2026-09-24)

`ast.before-7.5.34.json` 이 지우기 직전의 열거다. 두 diff 는 `Comparison.trees` 를 견준다 — 대상이 무엇이든 격리한 트리 둘이다.

# Function Logic Map: `_ast_value` (Python, a122 task 7.5.2 — **지웠다**)

편집 전 `tools/logic-map/check_analysis.py:1243-1248` · 분기 2 · 반환 2 · raise 0
(`ast.before-7.5.2.json` = revision `e9f905bd`).

`ast.json` 을 **디스크에서 직접** 읽는 둘째 철자였다. 7.5.2 에서 증거는 `_read_evidence` 가 한 번 읽고
모든 판정이 그 바이트를 `_parsed` 로 푼다. 호출자 셋(`_pinning_bundles` 의 선별 · `check` 의 열거형 호출
판정 · 옛 `check` 요구 대조)이 전부 옮겨 가 호출자가 0 이 되어 지웠다. 규칙(없거나 깨졌으면 `{}`)은
`_parsed` 가 그대로 가진다 — `read_text` 의 `OSError`·`ValueError` 는 `_read_evidence` 의 `None` 과
`_parsed` 의 `ValueError` 로 나뉘어 같은 답(`{}`)이 된다.

# Branch Test Map: `TestNoMutatingAccountOrWriteSurfaceIsNamed` (tools/a140-breakout-signal-shadow)

실패 갈래는 사본 변이(`../../mutation/mutation-after-priority-pin.txt`)로만 탐. 행 번호는 편집 뒤 `ast.json` 기준.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 102 파일 순회 | 자신 | yes | yes |
| B2 | 104 `WithPriority` 직접 호출(main.go 156) | 자신 | 갈래 없음 | yes |
| B3 | 106 둘째 인자가 Auxiliary 아님 | 자신 — 변이 a140-h-priority-critical·a140-i-priority-literal | 갈래 없음 | 변이 CAUGHT |
| B4 | 111 비선택자 건너뜀 | 자신 | yes | yes |
| B5 | 114 금지 이름 | 자신 — 변이 a140-d-forbidden·a140-j-budget-handle | 해당 없음 | 변이 CAUGHT |
| B6 | 117 `WithPriority` 참조 계수 | 자신 | 갈래 없음 | yes |
| B7 | 120 X 식별자 | 자신 | yes | yes |
| B8 | 121 추적 패키지 | 자신 | yes | yes |
| B9 | 123 목록 밖 이름 | 자신 — RED `official.Option`, 변이 a140-a-budget-name·a140-a-runtime-name·a140-c-official-with·a140-h | yes(RED 원인) | 변이 CAUGHT |
| B10 | 131 키 순회 | 자신 | yes | yes |
| B11 | 132 키를 못 봄 | 자신 — 변이 a140-f-no-accountbudget·a140-f-alias-runtime | 해당 없음 | 변이 CAUGHT |
| B12 | 137 호출 0 또는 참조 ≠ 호출 | 자신 — 변이 a140-k-no-priority-call·a140-l-priority-func-value | 갈래 없음 | 변이 CAUGHT |

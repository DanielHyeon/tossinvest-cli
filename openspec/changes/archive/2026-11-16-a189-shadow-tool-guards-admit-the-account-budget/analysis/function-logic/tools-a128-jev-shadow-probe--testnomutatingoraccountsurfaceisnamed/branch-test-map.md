# Branch Test Map: `TestNoMutatingOrAccountSurfaceIsNamed` (tools/a128-jev-shadow-probe)

시험 함수라 정상 경로는 이 함수 자신이 실행함. 실패 갈래는 사본 변이(`../../mutation/mutation-after-priority-pin.txt`)로만 탐. 행 번호는 편집 뒤 `ast.json` 기준.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 113 소스 파일 순회 | 자신 | yes | yes |
| B2 | 115 `WithPriority` 직접 호출(main.go 164) | 자신 | 갈래 없음(편집 전) | yes |
| B3 | 117 둘째 인자가 Auxiliary 아님 | 자신 — 변이 a128-h-priority-critical·a128-i-priority-literal | 갈래 없음 | 변이 CAUGHT |
| B4 | 122 비선택자 노드 건너뜀 | 자신 | yes | yes |
| B5 | 125 금지 이름 | 자신 — 변이 a128-d-forbidden(`client.PlaceOrder`)·a128-j-budget-handle(`factory.Budget`·`b.Acquire`) | 해당 없음 | 변이 CAUGHT |
| B6 | 128 `WithPriority` 참조 계수 | 자신 | 갈래 없음 | yes |
| B7 | 131 `official.*` 봄 | 자신 | yes | yes |
| B8 | 133 `official` 목록 밖 | 자신 — RED `official.Option`, 변이 a128-c-official-with | yes(RED 원인) | 변이 CAUGHT |
| B9 | 137 `tossconfig.*` 봄 | 자신 | yes | yes |
| B10 | 139 `tossconfig` 목록 밖 | 없음(무편집 갈래) | — | — |
| B11 | 143 예산 패키지 봄 | 자신 | 갈래 없음 | yes |
| B12 | 145 예산 이름 목록 밖 | 자신 — 변이 a128-a-budget-name·a128-a-runtime-name·a128-h(`accountbudget.Critical`) | 갈래 없음 | 변이 CAUGHT |
| B13 | 152 official·tossconfig 못 봄 | 없음(무편집 갈래) | — | — |
| B14 | 155 예산 패키지 못 봄 | 자신 — 변이 a128-f-no-accountbudget·a128-f-alias-runtime | 갈래 없음 | 변이 CAUGHT |
| B15 | 159 호출 0 또는 참조 ≠ 호출 | 자신 — 변이 a128-k-no-priority-call·a128-l-priority-func-value | 갈래 없음 | 변이 CAUGHT |

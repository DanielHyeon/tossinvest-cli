# Branch Test Map: `TestNoMutatingAccountOrWriteSurfaceIsNamed` (tools/a140-breakout-signal-shadow)

실패 갈래는 사본 변이(`../../mutation/`)로만 탐. 행 번호는 편집 뒤 `ast.json` 기준.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 96 파일 순회 | 자신 | yes | yes |
| B2 | 99 비선택자 건너뜀 | 자신 | yes | yes |
| B3 | 102 금지 이름 | 자신 — 변이 a140-d-forbidden(`client.PlaceOrder`) | 해당 없음 | 변이 CAUGHT |
| B4 | 105 X 식별자 | 자신 | yes | yes |
| B5 | 106 추적 패키지(예산 둘 포함) | 자신(main.go 141·146·156) | yes | yes |
| B6 | 108 목록 밖 이름 | 자신 — RED `official.Option`, 변이 a140-a-budget-name·a140-a-runtime-name·a140-c-official-with | yes(RED 원인) | 변이 CAUGHT |
| B7 | 116 키 순회 | 자신 | yes | yes |
| B8 | 117 키를 못 봄 | 자신 — 변이 a140-f-no-accountbudget·a140-f-alias-runtime | 해당 없음 | 변이 CAUGHT |

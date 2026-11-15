# Branch Test Map: `TestNoMutatingOrAccountSurfaceIsNamed` (tools/a128-jev-shadow-probe)

시험 함수라 정상 경로는 이 함수 자신이 실행함. 실패 갈래는 사본 변이(`../../mutation/`)로만 탐. 행 번호는 편집 뒤 `ast.json` 기준.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 107 소스 파일 순회 | 자신 | yes | yes |
| B2 | 110 비선택자 노드 건너뜀 | 자신 | yes | yes |
| B3 | 113 금지 이름 | 자신 — 변이 a128-d-forbidden(`client.PlaceOrder`) | 해당 없음(변이에서만) | 변이 CAUGHT |
| B4 | 116 `official.*` 봄 | 자신 | yes | yes |
| B5 | 118 `official` 목록 밖 | 자신 — RED 에서 `official.Option`, 변이 a128-c-official-with(`official.WithBaseURL`) | yes(RED 원인) | 변이 CAUGHT |
| B6 | 122 `tossconfig.*` 봄 | 자신 | yes | yes |
| B7 | 124 `tossconfig` 목록 밖 | 없음(무편집 갈래, 이번에 재지 않음) | — | — |
| B8 | 128 예산 패키지 봄 | 자신(main.go 149·154·164) | 갈래 없음(편집 전) | yes |
| B9 | 130 예산 이름 목록 밖 | 자신 — 변이 a128-a-budget-name(`accountbudget.Critical`)·a128-a-runtime-name(`accountbudgetruntime.Config`) | 갈래 없음 | 변이 CAUGHT |
| B10 | 137 official·tossconfig 못 봄 | 없음(무편집 갈래) | — | — |
| B11 | 140 예산 패키지 못 봄 | 자신 — 변이 a128-f-no-accountbudget·a128-f-alias-runtime | 갈래 없음 | 변이 CAUGHT |

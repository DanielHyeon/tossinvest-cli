# Function Logic Map: `TestA165ExpiredRetryWindowDoesNotCancelSettlement`

- Source: `internal/execgw/a165_cancel_refusal_boundary_test.go` (85-104)
- Revision: D5 4판 test pre-edit, `source_sha256` `70ec4addd02ca8554550d44194212fde8293419e9940383111f6cd8b1c9f8a6e`
- AST evidence: `ast.json` — AST 분기 3, 반환 0, 호출 9.

## 기존 의미와 수정 이유

기존 시험은 첫 broker 호출의 실제 25ms 지연을 `Window: 10ms`보다 큰 것으로 보고 첫 AMBIGUOUS/IN_DOUBT를 기대함. 이는 D5가 금지한 wall-clock 판정임. 4판에서는 Fake Clock이 0에서 retry sleep 1ms로만 진행하므로 두 번째 cancel은 허용되고 CONFIRMED가 되어야 함.

## 분기와 시험 계약

| 분기 | 조건 | 기존 기대 | D5 4판 기대 |
|---|---|---|---|
| B1 | 호출 결과 | IN_DOUBT | CONFIRMED |
| B2 | broker 호출 수 | 1 | 2 |
| B3 | journal 정착 | IN_DOUBT | CONFIRMED |

실제 지연은 전송 상한의 부모 context가 아니라 Clock 창 판정을 바꾸지 않는다는 회귀 시험으로 이름·기대값을 바꿈.

# Branch Test Map: `Gateway.checkSymbolFree`

- Source: `internal/execgw/gateway.go`

> Test 열은 그 분기를 지나는 시험(현존 · 통과)이다. 「미진입」 분기의 인용은 그 함수를 도는 시험이며 그 갈래 자체의 증명이 아니다(진입 실측 열이 정본).
> RED 는 새 a094 시험에 한해 변이 원장(`analysis/implementation/mutation-ledger.md`)이 잰다.

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:802` `if err != nil {` | 예 | `TestTransportOutcomeTable` | n/a | yes |
| B2 | `:805` `if len(unsettled) > 0 {` | 예 | `TestA094AnotherIntentInFlightStopsTheArmReleaseLoop` · `TestGatewayRefusesFailClosedBranchesBeforeDispatch` | n/a | yes |
| B3 | `:811` `if !plan.raisesExposure {` | 예 | `TestTransportOutcomeTable` | n/a | yes |
| B4 | `:815` `if err != nil {` | 아니오 | `TestTransportOutcomeTable` | n/a | yes |
| B5 | `:818` `for _, rec := range unresolved {` | 예 | `TestTransportOutcomeTable` | n/a | yes |
| B6 | `:820` `if err != nil {` | 아니오 | `TestTransportOutcomeTable` | n/a | yes |
| B7 | `:823` `if same {` | 아니오 | `TestTransportOutcomeTable` | n/a | yes |

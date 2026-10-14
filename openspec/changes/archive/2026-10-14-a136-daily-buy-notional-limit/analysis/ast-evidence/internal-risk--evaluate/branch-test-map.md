# Branch Test Map: `Evaluate`

- Source: `internal/risk/chain.go` (131-144) — base 7b073f55, a136 무편집 계획

| Branch | Anchor | Scenario | Test | a136 |
|---|---|---|---|---|
| B1 | 132 `preflight` 실패 | 사용 불가 입력 | `TestUnusableInputsRefuseRatherThanDefault` · `TestPreflightAndReductionReportDistinctSteps` | 영향 없음 |
| B2 | 135 `Side == SideSell` | 매도는 reduction 만 | `TestReductionPassesEveryEntryOnlyGate` · `TestReductionIsNeverCostGated` | 근거로 인용(매도 비차단) |
| B3 | 138 `range entryChain` | 고정 순서 | `TestChainOrderIsFixed` | 영향 없음 |
| B4 | 139 rung 거부 | 첫 실패 정지 | `TestFirstFailureStopsTheChain` | 영향 없음 |

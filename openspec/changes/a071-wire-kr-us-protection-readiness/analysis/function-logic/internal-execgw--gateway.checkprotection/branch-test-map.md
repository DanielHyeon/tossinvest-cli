# Branch Test Map: Gateway.checkProtection

- Source: `internal/execgw/protection.go` (89-110); current — HEAD `648df8ef`
- 시험 칸은 측정값: 무태그 `go test -c -cover` 바이너리를 시험 함수마다 따로 돌린 커버 프로필에서 그 분기 본문 블록을 실행한 시험(`analysis/harness/51_matrix.sh`). `_test.go` 함수와 base 함수는 계측 밖이라 그 사실을 행에 적음.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 90:2 | `if !plan.raisesExposure {` | `TestReductionNeverReadsReadinessProvider` · `TestACancelNeedsNoReservation` (+20) | 5.1 에서 재실행 안 함 | 블록 90.26-92.3 을 시험 22개가 실행, 전부 PASS |
| B2 | if at 93:2 | `if g.protectionCheckForTest != nil {` | `TestKRDriftDoesNotBlockValidUSAndDispatchDriftCallsNoBroker` · `TestARelaxationClearsTheLatch` (+82) | 5.1 에서 재실행 안 함 | 블록 93.37-95.3 을 시험 84개가 실행, 전부 PASS |
| B3 | if at 96:2 | `if g.protectionReadiness == nil {` | `TestARaisingMutationIsRefusedWhileProtectionIsUnwired` | 5.1 에서 재실행 안 함 | 블록 96.34-98.3 을 시험 1개가 실행, 전부 PASS |
| B4 | if at 100:2 | `if !ok {` | `TestNonIntegralOrUnsafeProtectionQuantityStopsBeforeProviderAndBroker` | 5.1 에서 재실행 안 함 | 블록 100.9-102.3 을 시험 1개가 실행, 전부 PASS |
| B5 | if at 106:2 | `if refusal != nil {` | `TestProductionDefaultRefusesKRAndUSBuyBeforeBroker` | 5.1 에서 재실행 안 함 | 블록 106.20-108.3 을 시험 1개가 실행, 전부 PASS |

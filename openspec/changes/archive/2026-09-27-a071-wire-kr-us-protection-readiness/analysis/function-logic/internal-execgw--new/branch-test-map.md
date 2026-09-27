# Branch Test Map: New

- Source: `internal/execgw/gateway.go` (148-189); current — HEAD `648df8ef`
- 시험 칸은 측정값: 무태그 `go test -c -cover` 바이너리를 시험 함수마다 따로 돌린 커버 프로필에서 그 분기 본문 블록을 실행한 시험(`analysis/harness/51_matrix.sh`). `_test.go` 함수와 base 함수는 계측 밖이라 그 사실을 행에 적음.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | switch at 149:2 | `switch {` | `TestKRDriftDoesNotBlockValidUSAndDispatchDriftCallsNoBroker` · `TestNonIntegralOrUnsafeProtectionQuantityStopsBeforeProviderAndBroker` (+110) | 5.1 에서 재실행 안 함 | 블록 148.42-149.9 을 시험 112개가 실행, 전부 PASS |
| B2 | case at 150:2 | `case opts.Journal == nil:` | 없음 | 5.1 에서 재실행 안 함 | 블록 150.27-151.106 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B3 | case at 152:2 | `case opts.Trading == nil:` | 없음 | 5.1 에서 재실행 안 함 | 블록 152.27-153.66 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B4 | case at 154:2 | `case strings.TrimSpace(opts.AccountRef) == "":` | 없음 | 5.1 에서 재실행 안 함 | 블록 154.48-155.69 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B5 | if at 158:2 | `if protectionCheckForTest == nil && !opts.forceReadinessAdapterForTest {` | `TestACancelNeedsNoReservation` · `TestAPreMintedIntentIDIsTheOneRecorded` (+103) | 5.1 에서 재실행 안 함 | 블록 158.73-160.3 을 시험 105개가 실행, 전부 PASS |
| B6 | if at 179:2 | `if g.clk == nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 179.18-181.3 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B7 | if at 182:2 | `if g.source == "" {` | 없음 | 5.1 에서 재실행 안 함 | 블록 182.20-184.3 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B8 | if at 185:2 | `if g.newID == nil {` | `TestKRDriftDoesNotBlockValidUSAndDispatchDriftCallsNoBroker` · `TestNonIntegralOrUnsafeProtectionQuantityStopsBeforeProviderAndBroker` (+110) | 5.1 에서 재실행 안 함 | 블록 185.20-187.3 을 시험 112개가 실행, 전부 PASS |

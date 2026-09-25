# Branch Test Map: ReadinessAdapter.Check

- Source: `internal/protection/readiness_adapter.go` (118-155); current — HEAD `648df8ef`
- 시험 칸은 측정값: 무태그 `go test -c -cover` 바이너리를 시험 함수마다 따로 돌린 커버 프로필에서 그 분기 본문 블록을 실행한 시험(`analysis/harness/51_matrix.sh`). `_test.go` 함수와 base 함수는 계측 밖이라 그 사실을 행에 적음.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 119:2 | `if adapter == nil \|\| adapter.provider == nil \|\| adapter.seal != adapterSeal(adapter) {` | 없음 | 5.1 에서 재실행 안 함 | 블록 119.87-121.3 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B2 | switch at 123:2 | `switch request.Market {` | `TestReadinessAdapterFailsClosedForMissingProviderAndDefaultSnapshot` · `TestReadinessAdapterRejectsInvalidMarketWithoutCallingProvider` | 5.1 에서 재실행 안 함 | 블록 122.2-123.24 을 시험 2개가 실행, 전부 PASS |
| B3 | case at 124:2 | `case "kr":` | `TestReadinessAdapterFailsClosedForMissingProviderAndDefaultSnapshot` | 5.1 에서 재실행 안 함 | 블록 124.12-125.46 을 시험 1개가 실행, 전부 PASS |
| B4 | case at 126:2 | `case "us":` | `TestReadinessAdapterFailsClosedForMissingProviderAndDefaultSnapshot` | 5.1 에서 재실행 안 함 | 블록 126.12-127.46 을 시험 1개가 실행, 전부 PASS |
| B5 | case at 128:2 | `default:` | `TestReadinessAdapterRejectsInvalidMarketWithoutCallingProvider` | 5.1 에서 재실행 안 함 | 블록 128.10-129.134 을 시험 1개가 실행, 전부 PASS |
| B6 | if at 131:2 | `if request.OrderType == "" \|\| request.Quantity == 0 {` | 없음 | 5.1 에서 재실행 안 함 | 블록 131.54-133.3 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B7 | if at 135:2 | `if err != nil {` | `TestReadinessAdapterFailsClosedForMissingProviderAndDefaultSnapshot` | 5.1 에서 재실행 안 함 | 블록 135.16-137.3 을 시험 1개가 실행, 전부 PASS |
| B8 | if at 146:2 | `if !decision.Allowed {` | `TestReadinessAdapterFailsClosedForMissingProviderAndDefaultSnapshot` | 5.1 에서 재실행 안 함 | 블록 146.23-148.3 을 시험 1개가 실행, 전부 PASS |
| B9 | if at 151:2 | `if previous.Valid() && (previous.market != checkpoint.market \|\| previous.generation != checkpoint.generat...` | 없음 | 5.1 에서 재실행 안 함 | 블록 151.156-153.3 을 무태그 패키지 시험 어느 것도 실행하지 않음 |

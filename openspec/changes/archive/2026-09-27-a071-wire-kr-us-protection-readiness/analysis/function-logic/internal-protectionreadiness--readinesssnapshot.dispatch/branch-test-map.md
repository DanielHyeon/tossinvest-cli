# Branch Test Map: ReadinessSnapshot.Dispatch

- Source: `internal/protectionreadiness/dispatch.go` (37-90); current — HEAD `648df8ef`
- 시험 칸은 측정값: 무태그 `go test -c -cover` 바이너리를 시험 함수마다 따로 돌린 커버 프로필에서 그 분기 본문 블록을 실행한 시험(`analysis/harness/51_matrix.sh`). `_test.go` 함수와 base 함수는 계측 밖이라 그 사실을 행에 적음.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 39:2 | `if snapshot.release != ReadinessRelease {` | 없음 | 5.1 에서 재실행 안 함 | 블록 39.42-42.3 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B2 | if at 43:2 | `if scope.AccountID == "" \|\| scope.ProfileID == "" \|\| !validMarket(scope.Market) \|\| now.IsZero() {` | 없음 | 5.1 에서 재실행 안 함 | 블록 43.98-46.3 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B3 | switch at 48:2 | `switch scope.Market {` | `TestCorruptKRMarketSnapshotDoesNotInvalidateSealedUSMarket` · `TestDispatchDecisionBindsExactScopeExpiryGenerationAndProvenance` (+3) | 5.1 에서 재실행 안 함 | 블록 47.2-48.22 을 시험 5개가 실행, 전부 PASS |
| B4 | case at 49:2 | `case MarketKR:` | `TestCorruptKRMarketSnapshotDoesNotInvalidateSealedUSMarket` · `TestDispatchDecisionBindsExactScopeExpiryGenerationAndProvenance` (+3) | 5.1 에서 재실행 안 함 | 블록 49.16-51.70 을 시험 5개가 실행, 전부 PASS |
| B5 | if at 51:3 | `if snapshot.krSeal != marketVerdictSeal(snapshot.release, verdict) {` | `TestCorruptKRMarketSnapshotDoesNotInvalidateSealedUSMarket` · `TestDispatchRejectsCorruptAndFutureIssuedSnapshots` | 5.1 에서 재실행 안 함 | 블록 51.70-54.4 을 시험 2개가 실행, 전부 PASS |
| B6 | case at 55:2 | `case MarketUS:` | `TestCorruptKRMarketSnapshotDoesNotInvalidateSealedUSMarket` · `TestDispatchDecisionBindsExactScopeExpiryGenerationAndProvenance` (+1) | 5.1 에서 재실행 안 함 | 블록 55.16-57.70 을 시험 3개가 실행, 전부 PASS |
| B7 | if at 57:3 | `if snapshot.usSeal != marketVerdictSeal(snapshot.release, verdict) {` | 없음 | 5.1 에서 재실행 안 함 | 블록 57.70-60.4 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B8 | if at 63:2 | `if verdict.State != Wired \|\| verdict.Code != RefusalNone {` | `TestDispatchDecisionBindsExactScopeExpiryGenerationAndProvenance` | 5.1 에서 재실행 안 함 | 블록 63.59-65.3 을 시험 1개가 실행, 전부 PASS |
| B9 | if at 67:2 | `if provenance.AccountID != scope.AccountID \|\| provenance.ProfileID != scope.ProfileID \|\| provenance.Ser...` | `TestDispatchDecisionBindsExactScopeExpiryGenerationAndProvenance` · `TestDispatchRejectsQuantityOrderSessionTriggerReplaceAndCapabilitySubstitution` | 5.1 에서 재실행 안 함 | 블록 74.167-77.3 을 시험 2개가 실행, 전부 PASS |
| B10 | if at 79:2 | `if provenance.IssuedAt.IsZero() \|\| provenance.ExpiresAt.IsZero() \|\| now.Before(provenance.IssuedAt) {` | 없음 | 5.1 에서 재실행 안 함 | 블록 79.102-82.3 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B11 | if at 83:2 | `if !now.Before(provenance.ExpiresAt) {` | `TestDispatchDecisionBindsExactScopeExpiryGenerationAndProvenance` | 5.1 에서 재실행 안 함 | 블록 83.39-86.3 을 시험 1개가 실행, 전부 PASS |

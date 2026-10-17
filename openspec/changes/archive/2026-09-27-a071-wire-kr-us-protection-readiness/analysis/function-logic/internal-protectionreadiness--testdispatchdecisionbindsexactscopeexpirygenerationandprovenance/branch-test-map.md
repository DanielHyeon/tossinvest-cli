# Branch Test Map: TestDispatchDecisionBindsExactScopeExpiryGenerationAndProvenance

- Source: `internal/protectionreadiness/dispatch_test.go` (8-33); current — HEAD `648df8ef`
- 시험 칸은 측정값: 무태그 `go test -c -cover` 바이너리를 시험 함수마다 따로 돌린 커버 프로필에서 그 분기 본문 블록을 실행한 시험(`analysis/harness/51_matrix.sh`). `_test.go` 함수와 base 함수는 계측 밖이라 그 사실을 행에 적음.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 16:2 | `if !decision.Allowed \|\| decision.Generation == 0 \|\| decision.SnapshotID == "" {` | `TestDispatchDecisionBindsExactScopeExpiryGenerationAndProvenance` | 5.1 에서 재실행 안 함 | 시험 자신; TestDispatchDecisionBindsExactScopeExpiryGenerationAndProvenance PASS (시험별 실행) |
| B2 | if at 19:2 | `if decision.Provenance.AccountID != "acct" \|\| decision.Provenance.ProfileID != "production" \|\| decision...` | `TestDispatchDecisionBindsExactScopeExpiryGenerationAndProvenance` | 5.1 에서 재실행 안 함 | 시험 자신; TestDispatchDecisionBindsExactScopeExpiryGenerationAndProvenance PASS (시험별 실행) |
| B3 | if at 24:2 | `if got := snapshot.Dispatch(wrongAccount, readinessNow); got.Allowed \|\| got.Code != RefusalScopeMismatch {` | `TestDispatchDecisionBindsExactScopeExpiryGenerationAndProvenance` | 5.1 에서 재실행 안 함 | 시험 자신; TestDispatchDecisionBindsExactScopeExpiryGenerationAndProvenance PASS (시험별 실행) |
| B4 | if at 27:2 | `if got := snapshot.Dispatch(scope, decision.Provenance.ExpiresAt); got.Allowed \|\| got.Code != RefusalExpi...` | `TestDispatchDecisionBindsExactScopeExpiryGenerationAndProvenance` | 5.1 에서 재실행 안 함 | 시험 자신; TestDispatchDecisionBindsExactScopeExpiryGenerationAndProvenance PASS (시험별 실행) |
| B5 | if at 30:2 | `if got := snapshot.Dispatch(DispatchScope{AccountID: "acct", ProfileID: "production", Market: MarketKR}, re...` | `TestDispatchDecisionBindsExactScopeExpiryGenerationAndProvenance` | 5.1 에서 재실행 안 함 | 시험 자신; TestDispatchDecisionBindsExactScopeExpiryGenerationAndProvenance PASS (시험별 실행) |

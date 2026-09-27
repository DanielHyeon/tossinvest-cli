# Branch Test Map: TestReadinessAdapterFailsClosedForMissingProviderAndDefaultSnapshot

- Source: `internal/protection/readiness_adapter_test.go` (23-45); current — HEAD `648df8ef`
- 시험 칸은 측정값: 무태그 `go test -c -cover` 바이너리를 시험 함수마다 따로 돌린 커버 프로필에서 그 분기 본문 블록을 실행한 시험(`analysis/harness/51_matrix.sh`). `_test.go` 함수와 base 함수는 계측 밖이라 그 사실을 행에 적음.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 25:2 | `if _, err := NewReadinessAdapter(nil, "acct", "production"); err == nil {` | `TestReadinessAdapterFailsClosedForMissingProviderAndDefaultSnapshot` | 5.1 에서 재실행 안 함 | 시험 자신; TestReadinessAdapterFailsClosedForMissingProviderAndDefaultSnapshot PASS (시험별 실행) |
| B2 | if at 30:2 | `if err != nil {` | `TestReadinessAdapterFailsClosedForMissingProviderAndDefaultSnapshot` | 5.1 에서 재실행 안 함 | 시험 자신; TestReadinessAdapterFailsClosedForMissingProviderAndDefaultSnapshot PASS (시험별 실행) |
| B3 | if at 34:2 | `if refusal == nil \|\| refusal.Code != protectionreadiness.RefusalMissingEvidence \|\| checkpoint.Valid() {` | `TestReadinessAdapterFailsClosedForMissingProviderAndDefaultSnapshot` | 5.1 에서 재실행 안 함 | 시험 자신; TestReadinessAdapterFailsClosedForMissingProviderAndDefaultSnapshot PASS (시험별 실행) |
| B4 | if at 37:2 | `if provider.calls != 1 {` | `TestReadinessAdapterFailsClosedForMissingProviderAndDefaultSnapshot` | 5.1 에서 재실행 안 함 | 시험 자신; TestReadinessAdapterFailsClosedForMissingProviderAndDefaultSnapshot PASS (시험별 실행) |
| B5 | if at 42:2 | `if _, refusal = adapter.Check(context.Background(), ReadinessRequest{Market: "kr", OrderType: "LIMIT", Quan...` | `TestReadinessAdapterFailsClosedForMissingProviderAndDefaultSnapshot` | 5.1 에서 재실행 안 함 | 시험 자신; TestReadinessAdapterFailsClosedForMissingProviderAndDefaultSnapshot PASS (시험별 실행) |

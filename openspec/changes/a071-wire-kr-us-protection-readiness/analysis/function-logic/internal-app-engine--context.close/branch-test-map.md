# Branch Test Map: Context.Close

- Source: `internal/app/engine/engine.go` (606-616); current — HEAD `648df8ef`
- 시험 칸은 측정값: 무태그 `go test -c -cover` 바이너리를 시험 함수마다 따로 돌린 커버 프로필에서 그 분기 본문 블록을 실행한 시험(`analysis/harness/51_matrix.sh`). `_test.go` 함수와 base 함수는 계측 밖이라 그 사실을 행에 적음.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 607:2 | `if c == nil {` | 없음 | 5.1 에서 재실행 안 함 | 블록 607.14-609.3 을 무태그 패키지 시험 어느 것도 실행하지 않음 |
| B2 | if at 612:2 | `if j != nil {` | `TestProductionEngineAssemblesPairedUnwiredReadinessProvider` · `TestProtectionStorageFailureCannotPreventSafetyRuntimeAssembly` (+52) | 5.1 에서 재실행 안 함 | 블록 612.14-614.3 을 시험 54개가 실행, 전부 PASS |

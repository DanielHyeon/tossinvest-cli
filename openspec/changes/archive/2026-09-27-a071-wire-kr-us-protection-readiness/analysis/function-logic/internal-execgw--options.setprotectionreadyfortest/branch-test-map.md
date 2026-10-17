# Branch Test Map: Options.SetProtectionReadyForTest

- Source: `internal/execgw/export_test.go` (39-42); base `775c37cb` (HEAD 에 함수 없음)
- 시험 칸은 측정값: 무태그 `go test -c -cover` 바이너리를 시험 함수마다 따로 돌린 커버 프로필에서 그 분기 본문 블록을 실행한 시험(`analysis/harness/51_matrix.sh`). `_test.go` 함수와 base 함수는 계측 밖이라 그 사실을 행에 적음.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | branchless happy path at 39:1 | `func (o *Options) SetProtectionReadyForTest() {` | `TestNoShippedFileClaimsProtection` · `TestReductionNeverReadsReadinessProvider` | 5.1 에서 재실행 안 함 | 대체 시험 PASS at HEAD 648df8ef (시험별 실행). 171739a4 가 스칼라 WIRED 세터를 지우고 export_test.go `init` 의 비스칼라 시험 하네스로 대체함 |

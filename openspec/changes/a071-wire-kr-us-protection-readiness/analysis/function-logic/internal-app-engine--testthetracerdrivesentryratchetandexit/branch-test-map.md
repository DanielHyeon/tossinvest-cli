# Branch Test Map: TestTheTracerDrivesEntryRatchetAndExit

- Source: `internal/app/engine/tracer_test.go` (116-144); base `775c37cb` (HEAD 에 함수 없음)
- 시험 칸은 측정값: 무태그 `go test -c -cover` 바이너리를 시험 함수마다 따로 돌린 커버 프로필에서 그 분기 본문 블록을 실행한 시험(`analysis/harness/51_matrix.sh`). `_test.go` 함수와 base 함수는 계측 밖이라 그 사실을 행에 적음.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 126:2 | `if err != nil {` | `TestScalarStartupReadinessCannotAuthorizeTheTracer` | 5.1 에서 재실행 안 함 | 대체 시험 PASS at HEAD 648df8ef (시험별 실행). 171739a4 가 같은 자리(116행)에서 이 시험을 스칼라 준비 상태가 tracer 를 승인하지 못함을 보이는 시험으로 교체함 |
| B2 | if at 129:2 | `if report.EntryOrderID == "" {` | `TestScalarStartupReadinessCannotAuthorizeTheTracer` | 5.1 에서 재실행 안 함 | 대체 시험 PASS at HEAD 648df8ef (시험별 실행). 171739a4 가 같은 자리(116행)에서 이 시험을 스칼라 준비 상태가 tracer 를 승인하지 못함을 보이는 시험으로 교체함 |
| B3 | if at 132:2 | `if !report.Closed {` | `TestScalarStartupReadinessCannotAuthorizeTheTracer` | 5.1 에서 재실행 안 함 | 대체 시험 PASS at HEAD 648df8ef (시험별 실행). 171739a4 가 같은 자리(116행)에서 이 시험을 스칼라 준비 상태가 tracer 를 승인하지 못함을 보이는 시험으로 교체함 |
| B4 | if at 135:2 | `if report.Proposals == 0 {` | `TestScalarStartupReadinessCannotAuthorizeTheTracer` | 5.1 에서 재실행 안 함 | 대체 시험 PASS at HEAD 648df8ef (시험별 실행). 171739a4 가 같은 자리(116행)에서 이 시험을 스칼라 준비 상태가 tracer 를 승인하지 못함을 보이는 시험으로 교체함 |
| B5 | if at 138:2 | `if report.Outcome == nil {` | `TestScalarStartupReadinessCannotAuthorizeTheTracer` | 5.1 에서 재실행 안 함 | 대체 시험 PASS at HEAD 648df8ef (시험별 실행). 171739a4 가 같은 자리(116행)에서 이 시험을 스칼라 준비 상태가 tracer 를 승인하지 못함을 보이는 시험으로 교체함 |
| B6 | if at 141:2 | `if report.Outcome.InitialQuantity != "1" {` | `TestScalarStartupReadinessCannotAuthorizeTheTracer` | 5.1 에서 재실행 안 함 | 대체 시험 PASS at HEAD 648df8ef (시험별 실행). 171739a4 가 같은 자리(116행)에서 이 시험을 스칼라 준비 상태가 tracer 를 승인하지 못함을 보이는 시험으로 교체함 |

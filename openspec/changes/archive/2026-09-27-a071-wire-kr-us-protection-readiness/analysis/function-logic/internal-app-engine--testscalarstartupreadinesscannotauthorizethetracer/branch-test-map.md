# Branch Test Map: TestScalarStartupReadinessCannotAuthorizeTheTracer

- Source: `internal/app/engine/tracer_test.go` (116-128); current — HEAD `648df8ef`
- 시험 칸은 측정값: 무태그 `go test -c -cover` 바이너리를 시험 함수마다 따로 돌린 커버 프로필에서 그 분기 본문 블록을 실행한 시험(`analysis/harness/51_matrix.sh`). `_test.go` 함수와 base 함수는 계측 밖이라 그 사실을 행에 적음.

| Branch | AST anchor | Scenario (source text) | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 122:2 | `if !errors.Is(err, engine.ErrTracerRefused) \|\| !strings.Contains(strings.ToLower(err.Error()), "protectio...` | `TestScalarStartupReadinessCannotAuthorizeTheTracer` | 5.1 에서 재실행 안 함 | 시험 자신; TestScalarStartupReadinessCannotAuthorizeTheTracer PASS (시험별 실행) |
| B2 | if at 125:2 | `if report.EntryOrderID != "" \|\| report.Closed \|\| report.Outcome != nil {` | `TestScalarStartupReadinessCannotAuthorizeTheTracer` | 5.1 에서 재실행 안 함 | 시험 자신; TestScalarStartupReadinessCannotAuthorizeTheTracer PASS (시험별 실행) |

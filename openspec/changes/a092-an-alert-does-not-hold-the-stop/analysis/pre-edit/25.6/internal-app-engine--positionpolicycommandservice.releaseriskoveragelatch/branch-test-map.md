# Branch Test Map: `PositionPolicyCommandService.ReleaseRiskOverageLatch`

- Source: `internal/app/engine/risk_relaxation_command.go`; **편집 전** 측정 — `analysis/harness/coverage-pre-a066-relax.json`(연결 워크트리 `81934b46`, `./internal/app/engine` 시험 14개 `TestA066|PositionPolicyCommand` 를 하나씩).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 99:2 | `relaxationRepo` 오류 | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B2 | if at 103:2 | `relaxationAuditor` 오류 | `TestA066LatchReleaseRefusedWithoutAnEngineAuditLog` | 편집 전(기준선) | 블록 103.16-105.3, PASS |
| B3 | if at 107:2 | `relaxationOperator` 오류 | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B4 | if at 117:2 | `repo.ReleaseRiskOverageLatch` 오류 | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |

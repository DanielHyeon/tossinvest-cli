# Branch Test Map: `PositionPolicyCommandService.ReleaseEntryLossLock`

- Source: `internal/app/engine/risk_relaxation_command.go`; **편집 전** 측정 — `analysis/harness/coverage-pre-a066-relax.json`(연결 워크트리 `81934b46`, `./internal/app/engine` 시험 14개 `TestA066|PositionPolicyCommand` 를 하나씩).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 66:2 | `relaxationRepo` 오류(원장이 해제 면 없음) | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B2 | if at 70:2 | `relaxationAuditor` 오류(audit 로그 없음) | `TestA066RelaxationRefusedWithoutAnEngineAuditLog` | 편집 전(기준선) | 블록 70.16-72.3, PASS |
| B3 | if at 74:2 | `relaxationOperator` 오류(운영자 이름 없음) | `TestA066RelaxationRequestRefusals` | 편집 전(기준선) | 블록 74.16-76.3, PASS |
| B4 | if at 84:2 | `repo.ReleaseEntryLossLock` 오류 | `TestA066JournalRefusalsCrossTheWireAsRefusals` · `TestA066RelaxationRequestRefusals` | 편집 전(기준선) | 블록 84.16-86.3, PASS |

# Branch Test Map: `PositionPolicyCommandService.ReleaseEntryLossLock`

- Source: `internal/app/engine/risk_relaxation_command.go`; **편집 뒤** 측정 — `analysis/harness/coverage-post-a066-relax.json`(연결 워크트리 `0e4f26af`, `./internal/app/engine` 시험 20개를 하나씩). 편집 전 번들은 `analysis/pre-edit/25.6/`에 보존.
- 재번호 없음(분기 넷, 줄만 +6).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 72:2 | relaxationRepo 오류 | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개(편집 전과 같음) |
| B2 | if at 76:2 | audit 로그 없음 | `TestA066RelaxationRefusedWithoutAnEngineAuditLog` | 해당 없음(분기 불변) | 블록 76.16-78.3, PASS |
| B3 | if at 80:2 | 운영자 이름 없음 | `TestA066RelaxationRequestRefusals` | 해당 없음(분기 불변) | 블록 80.16-82.3, PASS |
| B4 | if at 90:2 | 원장 해제 오류 | `TestA066JournalRefusalsCrossTheWireAsRefusals` · `TestA066RelaxationRequestRefusals` | 해당 없음(분기 불변) | 블록 90.16-92.3, PASS |

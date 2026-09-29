# Branch Test Map: `notifyRelaxation`

- Source: `internal/app/engine/risk_relaxation_command.go` (170-196); **편집 뒤** 측정 — `analysis/harness/coverage-post-a066-relax.json`(연결 워크트리 `0e4f26af`, `./internal/app/engine` 시험 20개를 하나씩). 편집 전 번들은 `analysis/pre-edit/25.6/`에 보존.
- 재번호: 새 B1(:173, nil 기록자) · 편집 전 B1(:177) → B2(:190).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 173:2 | 기록자 없음 → 통지 안 됨 | `TestA092RelaxationWithoutANotifierIsNotNotified` | 편집 전: 필드 없음(컴파일 RED) · 변이 R01 · R02 | 블록 173.20-177.3, PASS |
| B2 | if at 190:2 | 기록 실패 → 완화됨 · 통지 실패 + 진입 잠금 | `TestA066ReleaseStandsWhenTheNoticeFails` · `TestA092RelaxationNoticeFailureLatchesEntries` | 변이 R15 | 블록 190.16-193.3, PASS |

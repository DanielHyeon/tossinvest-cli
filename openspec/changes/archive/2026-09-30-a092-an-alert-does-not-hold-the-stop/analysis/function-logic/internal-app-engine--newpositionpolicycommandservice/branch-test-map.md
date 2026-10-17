# Branch Test Map: `NewPositionPolicyCommandService`

- Source: `internal/app/engine/position_policy_command.go` (93-112); **편집 뒤** 측정 — `analysis/harness/coverage-post-a066-relax.json`(연결 워크트리 `0e4f26af`, `./internal/app/engine` 시험 20개를 하나씩). 편집 전 번들은 `analysis/pre-edit/25.6/`에 보존.
- 재번호: B1 :91 → :94 · B2 :94 → :97 · **B3 :108 새 분기**.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 94:2 | 원장 없는 Context | `TestPositionPolicyCommandServiceRequiresEngineOwnedJournal` | 해당 없음(분기 불변) | 블록 94.40-96.3, PASS |
| B2 | if at 97:2 | 시계 기본값 | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개(편집 전과 같음) |
| B3 | if at 108:2 | 알림기 있음 → 기록자 | `TestA092CommandServiceTakesTheEngineNotifier` 외 12개 | 편집 전: 필드 없음(컴파일 RED) · 변이 R08 · R09 | 블록 108.26-110.3을 시험 13개가 실행, PASS |

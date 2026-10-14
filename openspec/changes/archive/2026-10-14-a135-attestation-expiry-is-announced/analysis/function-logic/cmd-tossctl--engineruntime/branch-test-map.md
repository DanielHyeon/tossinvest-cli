# Branch Test Map: `engineRuntime`

base `421bf53c` 대비 분기 변화 0. 분기 귀속은 a090 아카이브 번들의 측정을 인용함
(`openspec/changes/archive/2026-09-30-a090-an-unobserved-position-is-counted/analysis/function-logic/cmd-tossctl--engineruntime/branch-test-map.md`).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:658` 체결 감지 생성 실패 | `TestEngineRuntimeB1IsStructurallyUnreachableWithTheHardcodedNilHintPath` (`cmd/tossctl/engine_runtime_branch_test.go:134`) — a135 무편집 | no | yes |
| B2 | `:667` 대사 드라이버 생성 실패 | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` (`cmd/tossctl/engine_runtime_branch_test.go:49`) | no | yes |
| B3 | `:682` exit 관측자 생성 실패 | 같음 | no | yes |
| B4 | `:687` 복구 생성 실패 | 같음 | no | yes |
| B5 | `:691` 전략 진입 외곽 생성 실패 | 없음(a092 측정: 시험 0) — a135 무편집 | no | no |
| B6 | `:703` 배달 실행자 생성 실패 | 없음(a092 측정: 시험 0) — a135 무편집 | no | no |

## a135 편집(분기 밖) — 끝 return 의 보조 실행자 집합

| Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|
| 생산 조립의 보조 실행자 집합 = `alert-delivery` · `normal-alert-relay` · `attestation-watch`(집합 동일성) | `TestProductionRuntimeStartsExactlyOneAlertDeliverer` (`cmd/tossctl/a098_the_engine_registers_a_sender_test.go:20`) | yes(`analysis/red/red.md` — 편집 전 두 이름) | yes |
| 감독 루프 집합은 넷 그대로, 감시는 그 안에 없음 | `TestTheAlertDelivererIsNotASupervisedLoop` (같은 파일) | no | yes |
| 감시 실행자 모양: 이름 · 자기 정지 이벤트 · `OnStop` 없음 | `TestA135ProductionExecutorShape` (`internal/app/engine/a135_attestation_watch_internal_test.go`) | yes(빌드 실패) | yes |

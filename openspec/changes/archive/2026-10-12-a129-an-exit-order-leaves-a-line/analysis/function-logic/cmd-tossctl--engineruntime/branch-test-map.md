# Branch Test Map: `engineRuntime`

**편집 뒤**: 분기 불변, B3~B6 좌표 +2(B3 `:676` · B4 `:681` · B5 `:685` · B6 `:697`) — 아래 표의 좌표는 편집 전. 진입 귀속은 a090 아카이브 BTM(같은 분기 구조)과 현재 트리의 시험 이름으로 확인했다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:652` 체결 감지 생성 실패 | 없음(a092 · a090 측정: 시험 0) — a129 무편집 | no | no |
| B2 | `:661` 대사 드라이버 생성 실패 | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` (`cmd/tossctl/engine_runtime_branch_test.go:49`) | no | yes |
| B3 | `:674` exit 관측자 생성 실패 | 같음 | no | yes |
| B4 | `:679` 복구 생성 실패 | 같음 | no | yes |
| B5 | `:683` 전략 진입 외곽 생성 실패 | 없음 — a129 무편집 | no | no |
| B6 | `:695` 배달 실행자 생성 실패 | 없음 — a129 무편집 | no | no |

## RED → GREEN (a129)

- 이 함수의 `ExitObserverOptions` 리터럴에 `ExitOrderLog: logger` 가 있고 `Log` 가 없음(go/parser 구조 핀,
  `cmd/tossctl/a129_engine_runtime_exit_order_log_test.go`). a090 의 `TestA090EngineRuntimeWiresOnlyTheDedicatedUnobservedLogger` 는 계속 green.

- RED: `analysis/red/red-1.txt`(구조 핀 FAIL, base). GREEN: `TestA129EngineRuntimeWiresTheDedicatedExitOrderLogger` PASS + `TestA090EngineRuntimeWiresOnlyTheDedicatedUnobservedLogger` PASS. 변이 M2 CAUGHT.

# Branch Test Map: `NewRuntime`

Source: `internal/app/engine/runtime.go` (183-232). AST 기준 분기 12 / 이탈 7 /
defers 0 / go_statements 0.

거부 분기는 전부 `internal/app/engine/runtime_test.go`
`TestTheRuntimeRefusesWiringItCannotSupervise:450`의 표 하나가 덮는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:184` 루프가 없다 | `TestTheRuntimeRefusesWiringItCannotSupervise` — `"no loops"` | no | yes |
| B2 | `:189` 루프 순회 | 아래 B4~B9가 대표 | — | — |
| B3 | `:191` switch 진입 | 같은 위 | — | — |
| B4 | `:192` 이름이 공백 | 같은 테스트 — `"a loop with no name"` | no | yes |
| B5 | `:195` 이름 중복 | 같은 테스트 — `"two loops with one name"` | no | yes |
| B6 | `:197` `Run`이 nil | 같은 테스트 — `"a loop with no Run"` | no | yes |
| B7 | `:201` **`Health == nil` → 나머지 검증 건너뜀** | `TestAGracefulCancelStopsEveryLoopAndRaisesNoCritical:127` 등 — `Health` 없는 루프로 구성한 모든 테스트가 이 갈래를 지난다 | no | yes |
| B8 | `:204` trigger가 열거에 없다 | 같은 표 — `"a health source with no trigger"`, `"a trigger the journal does not enumerate"` | no | yes |
| B9 | `:209` 승격 루프인데 계좌가 없다 | 같은 표 — `"an escalating loop with no account"` | no | yes |
| B10 | `:222` `Clock`이 nil → `clock.System()` | **없음** — 시계를 안 준 구성으로 시간 의존 동작을 관측하는 테스트가 없다 | no | no |
| B11 | `:225` `Threshold <= 0` → 기본값 | `TestFiveConsecutiveFailuresEscalateOnceAndTheLoopKeepsRunning:276`이 기본 임계 5에 의존한다 | no | yes |
| B12 | `:228` `HealthInterval <= 0` → 기본값 | **없음** — 간격을 안 주고 폴링 주기를 관측하는 테스트가 없다(`TestTheSupervisorPollsHealthWhileTheLoopsRun:405`은 간격을 준다) | no | no |

## a092가 이 함수에 대해 지는 것

편집하지 않으므로 이 함수에 대한 새 RED는 없다. **a092가 지는 것은 새 행 하나를
이 검증표에 추가로 통과시키는 것**이다: 배달 루프가 B4~B9를 전부 지나
`NewRuntime`이 오류 없이 돌아와야 한다.

- **§6.0 R17-10**(배달 루프가 감독 아래 있다)이 그것을 관측한다. 그 테스트는
  `NewRuntime`의 분기가 아니라 **배달 루프를 포함한 배선이 통과하는지**를 보므로
  위 표에 행을 만들지 않는다.

미테스트 B10·B12는 기본값 대입 갈래이고 실패 모드가 없다.
a092가 편집하지 않으므로 여기서 만들지 않는다
(`not-applicable`: 이 change는 이 함수를 편집하지 않는다).

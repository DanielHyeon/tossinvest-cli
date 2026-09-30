# Branch Test Map: `Runtime.Run`

Source: `internal/app/engine/runtime.go` (261-332). AST 기준 분기 4 / 이탈 3 /
defers 3 / go_statements 2.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:262` `Recover`가 배선돼 있다 → 루프보다 먼저 돈다 | `internal/app/engine/runtime_test.go TestRecoveryRunsBeforeAnyLoopStarts:488` | no | yes |
| B2 | `:267` **복구 실패 → 루프를 하나도 안 띄운다** | `TestAnIncompleteRecoveryStartsNothing:529` | no | yes |
| B3 | `:277` 루프 fan-out → 첫 정지가 전부를 내린다 | `TestAGracefulCancelStopsEveryLoopAndRaisesNoCritical:127`, `TestALoopReturningOnItsOwnStopsEverythingAndIsCritical:171` | no | yes |
| B4 | `:304` 참 — 취소로 인한 정지 → `nil`, critical 없음 · 거짓 — 루프가 스스로 반환 → `ErrLoopFailed` + critical | 참: `TestAGracefulCancelStopsEveryLoopAndRaisesNoCritical:127` · 거짓: `TestALoopReturningOnItsOwnStopsEverythingAndIsCritical:171`, `TestALoopThatSimplyReturnsIsAlsoAFailure:215`, `TestALoopFailingDuringAShutdownIsStillReported:239` | no | yes |

B4의 참·거짓 갈래를 한 행에 담았다. `check_analysis.py`가 분기 ID 중복을
거부하므로 한 분기는 한 행이고, 갈래는 행 안에서 나눈다.

## a092가 이 함수에 대해 지는 것

편집하지 않으므로 이 함수에 대한 새 RED는 없다. **그러나 이 함수가 배달 루프에
거는 두 제약은 a092가 진다** — 둘 다 위 표의 분기가 아니라 배달 루프 쪽 속성이다.

- **R17-8** — best-effort 알림도 관측 루프를 붙잡지 않는다. 배달 루프의 `Run`이
  전송 실패로 반환하지 않는지를 포함해서 본다.
- **R17-10** — 배달 루프가 `SupervisedLoop`으로 등록돼 감독을 받고,
  ctx 취소에 반응해 반환한다(그래야 `wg.Wait:300`이 풀린다).

**`TestALoopThatSimplyReturnsIsAlsoAFailure:215`가 이미 경고를 적어 두고 있다**:
nil 오류로 조용히 반환하는 루프도 실패로 다룬다. 배달 루프가 그 함정에 빠지면
엔진 전체가 내려가고, 그것은 알림 실패보다 무거운 결과다.

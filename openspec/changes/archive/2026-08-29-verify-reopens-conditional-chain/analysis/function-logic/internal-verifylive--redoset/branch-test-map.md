# Branch Test Map: `RedoSet`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `range Steps()` — catalogue 순서로 돌려준다 | `TestRedoSetTakesFailedAndSkippedStepsOnly` | no — 기존 동작 | yes |
| B2 | `if !ok` — 시도된 적 없는 단계·승인 줄은 대상이 아니다 | `TestRedoSetOnAnEmptyRecordIsEmpty`, `TestRedoSetIgnoresTheApprovalLine` | no — 기존 동작 | yes |
| B3 | `RedoableVerdict(...) \|\| subjectLost(...)` — **이 change가 우변을 추가한 분기**. 하위 경로는 아래 | `TestRedoSetReopensARegisterWhoseConditionalIsGone` | **yes** | yes |

B3 하위 경로 (분기 id는 하나이고 조건이 OR 두 개다):

| 경로 | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| 좌변 true | `fail`·`skipped`는 대상이다 | `TestRedoSetTakesFailedAndSkippedStepsOnly`, `TestRedoSetReadsTheNewestVerdict` | no — 기존 동작 | yes |
| 우변 true | **대상이 사라진 등록 단계는 다시 열린다** | `TestRedoSetReopensARegisterWhoseConditionalIsGone` | **yes** | yes |
| 둘 다 false | 통과한 단계는 그대로 닫혀 있다 | `TestRedoSetNeverOffersAPassedStep` | no — 기존 테스트, 수정 후에도 통과 | yes |
| 둘 다 false | 정상 종료한 체인은 다시 열리지 않는다 | `TestRedoSetLeavesACompletedChainClosed` | no — 좁음 고정용 | yes |
| 둘 다 false | 조건주문이 아직 살아 있으면 열지 않는다 | `TestRedoSetDoesNotReopenWhileTheConditionalIsAlive` | no — 좁음 고정용 | yes |
| 둘 다 false | `deferred` 의존 단계만으로는 열지 않는다 | `TestRedoSetDoesNotReopenForADeferredDependentAlone` | no — 좁음 고정용 | yes |
| 둘 다 false | 조건주문을 남기지 않은 pass는 열지 않는다 | `TestRedoSetDoesNotReopenAPassThatLeftNothingBehind` | no — 좁음 고정용 | yes |

콘솔 측 배선(표 ↔ 집합 일치, design.md D3):

| 대상 | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| `pageFuncs.inRedo` + `templates.go` | 표의 행이 버튼이 실행할 집합과 정확히 같다 | `TestTheRemeasureTableNamesExactlyTheStepsTheButtonWouldRun` | **yes** | yes |
| 같음 | 재개통된 등록 단계가 표에 보인다 | `TestTheRemeasureTableShowsTheReopenedRegister` | **yes** | yes |

RED 실행 기록 (구현 전, `go test ./internal/verifylive/`):

```
--- FAIL: TestRedoSetReopensARegisterWhoseConditionalIsGone (0.00s)
    redo_test.go:170: the chain cannot be measured again from the console:
    [conditional-persist conditional-modify conditional-cancel]
```

RED 실행 기록 (콘솔 수정만 되돌린 상태, `go test ./internal/console/ -run TestTheRemeasureTable`):
버튼은 `대상: conditional-register, conditional-persist, conditional-modify, conditional-cancel`
4단계를 실행하는데 표에는 `conditional-persist`·`conditional-modify`·`conditional-cancel`
3행만 렌더됐다 — 체인을 여는 바로 그 행이 조작자가 승인 전에 읽는 표에서 빠져 있었다.

GREEN 실행 기록: `go test ./... -count=1` → 3723 passed in 57 packages (구현 전 3712).

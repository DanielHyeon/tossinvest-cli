# Function Logic Map: `Console.handleStart`

- Source: `internal/console/pages.go` (revision: current)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `console-click-approval`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| mode 폼 값 | redo 또는 그 외 | 요청 | redo면 재측정 집합을 기록에서 계산 |
| 증거 기록 | JSONL | 디스크 | 읽기 실패는 안내 후 중단 — 전송 0건 |
| Pending 집합 | 판정이 없는 단계 | BuildProgress | 비어 있으면 이어하기를 시작하지 않는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if r.PostFormValue("mode") == startModeRedo {` (internal/console/pages.go:168, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestResumeWithNothingPendingStartsNoRun, TestResumeStaysOfferedWhileAStepIsPending, TestAnOrdinaryStartIsStillNotARemeasure, TestARemeasureAskedForWithNothingToRedoSendsNothing |
| B2 | `} else if snap := c.readVerify(); snap.Present && len(snap.Pending) == 0 {` (internal/console/pages.go:180, else) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestResumeWithNothingPendingStartsNoRun, TestResumeStaysOfferedWhileAStepIsPending, TestAnOrdinaryStartIsStillNotARemeasure, TestARemeasureAskedForWithNothingToRedoSendsNothing |
| B3 | `if err != nil {` (internal/console/pages.go:170, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestResumeWithNothingPendingStartsNoRun, TestResumeStaysOfferedWhileAStepIsPending, TestAnOrdinaryStartIsStillNotARemeasure, TestARemeasureAskedForWithNothingToRedoSendsNothing |
| B4 | `if len(set) == 0 {` (internal/console/pages.go:174, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestResumeWithNothingPendingStartsNoRun, TestResumeStaysOfferedWhileAStepIsPending, TestAnOrdinaryStartIsStillNotARemeasure, TestARemeasureAskedForWithNothingToRedoSendsNothing |
| B5 | `} else if snap := c.readVerify(); snap.Present && len(snap.Pending) == 0 {` (internal/console/pages.go:180, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestResumeWithNothingPendingStartsNoRun, TestResumeStaysOfferedWhileAStepIsPending, TestAnOrdinaryStartIsStillNotARemeasure, TestARemeasureAskedForWithNothingToRedoSendsNothing |
| B6 | `if len(snap.Redo) > 0 {` (internal/console/pages.go:187, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestResumeWithNothingPendingStartsNoRun, TestResumeStaysOfferedWhileAStepIsPending, TestAnOrdinaryStartIsStillNotARemeasure, TestARemeasureAskedForWithNothingToRedoSendsNothing |
| B7 | `if _, err := c.startRun(redo); err != nil {` (internal/console/pages.go:193, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestResumeWithNothingPendingStartsNoRun, TestResumeStaysOfferedWhileAStepIsPending, TestAnOrdinaryStartIsStillNotARemeasure, TestARemeasureAskedForWithNothingToRedoSendsNothing |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Console.redoSet | 재측정 대상을 기록에서 계산(폼 금지) | 에러는 안내로 종결 | AST callees |
| Console.readVerify | 기록 상태(Present·Pending·Redo) 판독 | 읽기 실패는 Present=false로 안전 방향 | AST callees |
| Console.startRun | run 시작 | errProcessSpent·errRunInProgress는 안내로 종결 | AST callees |

## State mutations and fallbacks

- run을 생성하고 고루틴을 띄운다 — 단, 위 가드를 통과했을 때만.
- 새로 추가된 분기: 기록이 있고 Pending이 비었으면 run을 만들지 않고 안내한다(무동작 run 방지).
- 재측정 대상 계산은 계속 기록에서만 한다 — 폼이 단계를 지명할 수 없다(무변경).

## Safety conclusion

- Safe edit boundary: 무동작 방지 분기는 redo 경로를 건드리지 않아야 하고, awaiting-restart(비terminal)를 Pending으로 세는 의미를 유지해야 한다. 그러지 않으면 조건주문 존속 측정의 이어하기가 막힌다.
- High-risk impact: no — 시작만 한다. 승인 게이트는 여전히 그 뒤에 있다.

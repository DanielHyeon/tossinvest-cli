# Function Logic Map: `TestARemeasureRunsOnlyTheFailedStepAndStillAsksForANewApproval`

- Source: `internal/console/remeasure_test.go` (revision: current)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `console-click-approval`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 테스트 입력 | 테스트가 만든 하네스 상태 | 테스트 코드 | 실패는 t.Error/t.Fatal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if len(view.Redo) != 1 \|\| view.Redo[0] != verifylive.StepOrderCancel {` (internal/console/remeasure_test.go:139, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B2 | `if view.Batch == nil {` (internal/console/remeasure_test.go:142, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B3 | `if n := h.broker.mutationCount(); n != 0 {` (internal/console/remeasure_test.go:145, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B4 | `for _, m := range view.Batch.Plan.Mutations {` (internal/console/remeasure_test.go:149, range) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B5 | `if m.Step != verifylive.StepOrderCancel {` (internal/console/remeasure_test.go:150, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B6 | `if got := len(h.broker.placements()); got != 1 {` (internal/console/remeasure_test.go:158, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B7 | `for _, o := range final.Summary.Outcomes {` (internal/console/remeasure_test.go:163, range) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B8 | `if o.AlreadySettled {` (internal/console/remeasure_test.go:164, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B9 | `if !ran[verifylive.StepOrderCancel] {` (internal/console/remeasure_test.go:170, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B10 | `for _, step := range verifylive.Steps() {` (internal/console/remeasure_test.go:173, range) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B11 | `if step.ID == verifylive.StepOrderCancel {` (internal/console/remeasure_test.go:174, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B12 | `if !settled[step.ID] {` (internal/console/remeasure_test.go:177, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 하네스 헬퍼(newHarness·post·waitFor…) | 시나리오 구동 | 테스트 전용 | AST callees |

## State mutations and fallbacks

- 테스트 함수다 — httptest 서버와 fake broker만 사용하고 실계좌·실 파일 시스템 밖 side effect가 없다.
- 이 change에서의 변경은 승인 방식(타이핑→클릭)에 맞춘 시나리오 갱신이다.

## Safety conclusion

- Safe edit boundary: 주장(assertion)이 약해지지 않을 것 — 모든 거부 경로는 계속 'mutating broker call 0건'을 확인한다.
- High-risk impact: no — 테스트 코드이며 프로덕션 경로에 포함되지 않는다.

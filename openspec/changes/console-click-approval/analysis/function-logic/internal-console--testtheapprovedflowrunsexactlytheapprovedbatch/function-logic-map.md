# Function Logic Map: `TestTheApprovedFlowRunsExactlyTheApprovedBatch`

- Source: `internal/console/console_test.go` (revision: current)
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
| B1 | `if len(view.Batch.Plan.Mutations) == 0 {` (internal/console/console_test.go:470, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B2 | `if strings.Contains(page, view.Batch.Nonce) {` (internal/console/console_test.go:474, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B3 | `if !strings.Contains(shown, flatten(view.Batch.Summary())) {` (internal/console/console_test.go:482, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B4 | `for _, m := range view.Batch.Plan.Mutations {` (internal/console/console_test.go:488, range) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B5 | `if !strings.Contains(shown, flatten(m.HeadlineKO())) {` (internal/console/console_test.go:489, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B6 | `if !strings.Contains(shown, flatten(m.EndsKO)) {` (internal/console/console_test.go:492, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B7 | `if n := h.broker.mutationCount(); n == 0 {` (internal/console/console_test.go:500, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B8 | `for _, p := range h.broker.placements() {` (internal/console/console_test.go:503, range) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B9 | `if p.Symbol != "005930" \|\| !strings.EqualFold(p.Side, "buy") \|\| p.Quantity != 1 {` (internal/console/console_test.go:504, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B10 | `if final.Err != "" && strings.Contains(final.Err, verifylive.ErrOutsidePlan.Error()) {` (internal/console/console_test.go:508, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B11 | `if approval.Verdict != verifylive.VerdictPass {` (internal/console/console_test.go:513, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B12 | `if got, want := observation(approval, "approval.plan_digest"), view.Batch.Plan.Digest(); got != want {` (internal/console/console_test.go:516, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |

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

# Function Logic Map: `Runner.Run`

- Source: `internal/verifylive/runner.go` (revision: current)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `verify-survives-already-processing`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 이 함수의 입력 | 시그니처가 정의한 범위 | 호출자 | 범위 밖 값은 정규화되거나 거부된다 |
| 증거 기록의 Outstanding | 이 도구가 만들고 취소되지 않은 객체 | capability-verify*.jsonl | 기록이 없으면 대상도 없다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if halt, err := r.approveBatch(ctx); err != nil \|\| halt != "" {` (internal/verifylive/runner.go:269, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B2 | `if outcome, err, stop := r.cleanup(ctx); outcome.Step != "" \|\| err != nil {` (internal/verifylive/runner.go:279, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B3 | `if outcome.Step != "" {` (internal/verifylive/runner.go:280, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B4 | `if stop {` (internal/verifylive/runner.go:283, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B5 | `if outcome.Reason == "" {` (internal/verifylive/runner.go:286, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B6 | `for _, step := range Steps() {` (internal/verifylive/runner.go:294, range) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B7 | `if err := ctx.Err(); err != nil {` (internal/verifylive/runner.go:295, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B8 | `if settled, verdict := r.settled(step.ID); settled {` (internal/verifylive/runner.go:302, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B9 | `if reason, skip := r.preflight(step); skip {` (internal/verifylive/runner.go:313, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B10 | `} else {` (internal/verifylive/runner.go:315, else) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B11 | `if err := r.recorder.Append(entry); err != nil {` (internal/verifylive/runner.go:321, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B12 | `if sr.verdict == VerdictAwaitingRestart {` (internal/verifylive/runner.go:330, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B13 | `if errors.Is(sr.abort, ErrNotATerminal) {` (internal/verifylive/runner.go:336, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B14 | `if errors.Is(sr.abort, ErrOutsidePlan) {` (internal/verifylive/runner.go:342, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B15 | `if sr.abort != nil && errors.Is(sr.abort, context.Canceled) {` (internal/verifylive/runner.go:351, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B16 | `if leftovers := undeliberate(summary.Outstanding); len(leftovers) > 0 {` (internal/verifylive/runner.go:360, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST callees 참조 | 단계 본문 실행 직후 sweep을 부른다. 승인 게이트·판정·중단 규칙은 무변경이다. | 호출자 계약을 따른다 | ast.json |

## State mutations and fallbacks

- 단계 본문 실행 직후 sweep을 부른다. 승인 게이트·판정·중단 규칙은 무변경이다.
- 승인 게이트·계획 인가(Plan.Authorises)·노출 상한·1주 규칙은 이 함수 밖이며 무변경이다.

## Safety conclusion

- Safe edit boundary: sweep은 preflight로 건너뛴 단계에는 돌지 않는다(보낸 것이 없다). 위치가 dispatch 앞으로 가면 아직 만들지 않은 것을 정리하려 든다.
- High-risk impact: yes — 실계좌에 나가는 요청의 목록·순서를 결정한다.

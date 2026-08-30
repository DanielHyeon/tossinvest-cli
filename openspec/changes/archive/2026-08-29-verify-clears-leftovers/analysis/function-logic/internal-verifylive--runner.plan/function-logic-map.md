# Function Logic Map: `Runner.Plan`

- Source: `internal/verifylive/plan.go` (revision: current)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `verify-clears-leftovers`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 이 함수의 입력 | 시그니처가 정의한 범위 | 호출자 | 범위 밖 값은 정규화되거나 거부된다 |
| 증거 기록의 Outstanding | 이 도구가 만들고 취소되지 않은 객체 | capability-verify*.jsonl | 기록이 없으면 대상도 없다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, line := range r.planCleanup() {` (internal/verifylive/plan.go:537, range) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B2 | `for _, step := range Steps() {` (internal/verifylive/plan.go:542, range) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B3 | `if settled, verdict := r.settled(step.ID); settled {` (internal/verifylive/plan.go:543, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B4 | `if step.Mutates {` (internal/verifylive/plan.go:544, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B5 | `if reason, skip := r.preflightStatic(step, passed); skip {` (internal/verifylive/plan.go:553, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B6 | `if step.Mutates {` (internal/verifylive/plan.go:554, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B7 | `if !step.Mutates {` (internal/verifylive/plan.go:560, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B8 | `if !ok {` (internal/verifylive/plan.go:566, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B9 | `for _, line := range lines {` (internal/verifylive/plan.go:575, range) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST callees 참조 | 계획이 catalogue보다 먼저 정리 줄을 싣는다. 줄의 출처는 `r.planCleanup()`이고 대상은 증거 기록의 Outstanding이다. 단계별 계획 생성 로직은 무변경이다. | 호출자 계약을 따른다 | ast.json |

## State mutations and fallbacks

- 계획이 catalogue보다 먼저 정리 줄을 싣는다. 줄의 출처는 `r.planCleanup()`이고 대상은 증거 기록의 Outstanding이다. 단계별 계획 생성 로직은 무변경이다.
- 승인 게이트·계획 인가(Plan.Authorises)·노출 상한·1주 규칙은 이 함수 밖이며 무변경이다.

## Safety conclusion

- Safe edit boundary: 정리 줄이 목록에서 빠지면 정리 취소가 ErrOutsidePlan으로 거절된다(안전 방향). 반대로 대상이 없을 때 줄이 생기면 사람이 존재하지 않는 객체의 취소를 승인하게 된다 — TestNoLeftoversMeansNoCleanupLines가 그 방향을 막는다. 순서는 의미가 있다: 잔여물이 노출 상한을 채우고 있는 동안 아래 단계는 아무것도 보낼 수 없다.
- High-risk impact: yes — 실계좌에 나가는 요청의 목록·순서를 결정한다.

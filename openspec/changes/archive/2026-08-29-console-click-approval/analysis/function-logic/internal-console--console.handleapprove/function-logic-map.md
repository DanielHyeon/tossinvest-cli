# Function Logic Map: `Console.handleApprove`

- Source: `internal/console/pages.go` (revision: current)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `console-click-approval`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 세션/핸드오프 자격 | 유효한 쿠키 | session0 미들웨어(무변경) | 없으면 403 — 이 핸들러에 도달하지 않는다 |
| CSRF 토큰 | 이 프로세스의 토큰 | mutating 미들웨어(무변경) | 틀리면 403 — 도달하지 않는다 |
| 대기 중 배치 | run.pending != nil && decided != nil | runState.confirmBatch | 없으면 '승인을 기다리는 배치가 없다'로 거부 |
| 승인 창 | Batch.ExpiresAt 이전 | verifylive.Batch | 만료면 ErrConfirmationExpired를 전달 — 전송 0건 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if run == nil {` (internal/console/pages.go:214, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestApprovalIsOneClickWithNothingTyped, TestAnExpiredApprovalSendsNothing, TestAWrongCSRFTokenSendsNothing, TestAMissingSessionOnTheApprovalSendsNothing, TestTheApprovalWindowIsJudgedByVerifylive |
| B2 | `if !view.Awaiting \|\| view.Batch == nil {` (internal/console/pages.go:219, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestApprovalIsOneClickWithNothingTyped, TestAnExpiredApprovalSendsNothing, TestAWrongCSRFTokenSendsNothing, TestAMissingSessionOnTheApprovalSendsNothing, TestTheApprovalWindowIsJudgedByVerifylive |
| B3 | `if view.Batch.Expired(c.now()) {` (internal/console/pages.go:225, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestApprovalIsOneClickWithNothingTyped, TestAnExpiredApprovalSendsNothing, TestAWrongCSRFTokenSendsNothing, TestAMissingSessionOnTheApprovalSendsNothing, TestTheApprovalWindowIsJudgedByVerifylive |
| B4 | `if !run.deliver(answer) {` (internal/console/pages.go:228, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestApprovalIsOneClickWithNothingTyped, TestAnExpiredApprovalSendsNothing, TestAWrongCSRFTokenSendsNothing, TestAMissingSessionOnTheApprovalSendsNothing, TestTheApprovalWindowIsJudgedByVerifylive |
| B5 | `if answer != nil {` (internal/console/pages.go:232, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestApprovalIsOneClickWithNothingTyped, TestAnExpiredApprovalSendsNothing, TestAWrongCSRFTokenSendsNothing, TestAMissingSessionOnTheApprovalSendsNothing, TestTheApprovalWindowIsJudgedByVerifylive |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| Console.currentRun | 현재 run 조회 | 없음 | AST callees |
| runState.snapshot | 일관된 복사본 | 없음 | AST callees |
| verifylive.Batch.Expired | 창 판정 위임 — 두 번째 시계 규칙 금지 | 없음 | AST callees + static guard |
| runState.deliver | 판정을 runner에게 전달 | 이미 닫힌 창이면 false | AST callees |

## State mutations and fallbacks

- run.decided 채널로 승인/거부를 1회 전달한다. 두 번째 제출은 deliver가 false를 반환해 '창이 닫혔다'로 끝난다.
- nonce 폼 값을 더 이상 읽지 않는다 — 승인은 이 POST 자체다.
- 계좌·journal·게이트에 직접 닿는 경로는 없다(무변경).

## Safety conclusion

- Safe edit boundary: 세션·CSRF·대기 배치·만료 네 조건. 이 중 하나라도 없이 deliver(nil)에 도달하면 사람 승인 없는 LIVE 요청이 가능해진다(§0.1).
- High-risk impact: yes — 실계좌 mutation 직전의 승인 레일이다. 승인 이후 레일(Plan.Authorises·상한·취소)은 이 함수 밖이며 무변경.

# Function Logic Map: `Runner.writeBanner`

- Source: `internal/verifylive/runner.go` (revision: current)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `console-click-approval`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| r.confirmEach | bool | Options | true면 단계별 확인 문구(무변경) |
| r.approvalChannel | 채널 상수 | New | 콘솔 클릭이면 '화면의 클릭'으로 안내 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if r.confirmEach {` (internal/verifylive/runner.go:347, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestTheApprovalScreenAsksForNoTypedString(console), TestApprovalIsOneClickWithNothingTyped(console) |
| B2 | `if r.approvalChannel == ApprovalChannelConsoleClick {` (internal/verifylive/runner.go:357, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestTheApprovalScreenAsksForNoTypedString(console), TestApprovalIsOneClickWithNothingTyped(console) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| fmt.Fprintf | 진행 출력에 배너를 쓴다 | 없음(io.Writer) | AST callees |

## State mutations and fallbacks

- 출력만 한다. 계좌·기록·계획에 영향이 없다.
- 콘솔에서는 이 출력이 웹 화면의 진행 로그가 되므로, 문구가 실제 승인 방식과 일치해야 한다.

## Safety conclusion

- Safe edit boundary: 배너는 실제로 만나게 될 게이트를 말해야 한다. 잘못된 안내는 사용자가 없는 절차를 찾게 만든다.
- High-risk impact: no — 순수 안내 출력.

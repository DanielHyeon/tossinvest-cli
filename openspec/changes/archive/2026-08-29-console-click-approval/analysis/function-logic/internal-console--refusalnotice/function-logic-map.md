# Function Logic Map: `refusalNotice`

- Source: `internal/console/pages.go` (revision: current)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `console-click-approval`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| err | nil·만료·그 외 | handleApprove | 만료와 그 외를 구분해 안내 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `switch {` (internal/console/pages.go:256, switch) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestAnExpiredApprovalSendsNothing |
| B2 | `case err == nil:` (internal/console/pages.go:257, case) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestAnExpiredApprovalSendsNothing |
| B3 | `case strings.Contains(err.Error(), verifylive.ErrConfirmationExpired.Error()):` (internal/console/pages.go:259, case) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestAnExpiredApprovalSendsNothing |
| B4 | `default:` (internal/console/pages.go:261, case) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | TestAnExpiredApprovalSendsNothing |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| strings.Contains | 만료 판별 | 없음 | AST callees |

## State mutations and fallbacks

- 없음 — 문자열만 만든다.
- 문구를 타이핑 전제에서 클릭 전제로 고쳤다(동작 무변경).

## Safety conclusion

- Safe edit boundary: 어느 갈래든 '아무것도 전송되지 않았다'를 말해야 한다.
- High-risk impact: no.

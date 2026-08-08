# Function Logic Map: `classifyMutation`

- Source: `internal/execgw/classify.go` (`21`–`79`)
- Qualified: `classifyMutation`
- AST evidence: `ast.json` (`source_sha256` 808e462cd6f9f136…)
- Risk scan: `risk-pattern-report.md`
- 분기 7 · return 5 · 호출 16

**역할.** 브로커 호출 하나의 결과를 원장의 dispatch 분류로 바꾼다. **순서가 설계다** — 로컬 거부 → 브로커의 서술적 거절 → status.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `err` | 브로커/로컬 오류 | 호출자 | B1이 nil이면 Acked |
| `result` | `domain.MutationResult` | 브로커 응답 | B1에서 broker order id 추출 |
| `send` | 전송 진행도 | dispatch 추적기 | B5에서 `ClassifyHTTPMutation`에 넘어간다 |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/... -count=1 -covermode=set`의 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:22` `if err == nil {` | `brokerOrderID`, `string` | :23 | 예 |
| B2 | if | `:32` `if reason, refused := policyRefusal(err); refused {` | `err.Error`, `policyRefusal`, `string` | :33 | 예 |
| B3 | if | `:46` `if reason, refused := ClassifyBrokerRefusal(err); refused {` | `ClassifyBrokerRefusal` | — | 예 |
| B4 | if | `:49` `if errors.As(err, &branch) && branch.Source == trading.BranchSourcePostPrepareConfirmation {` | `err.Error`, `errors.As`, `string` | :54 | 아니오 |
| B5 | if | `:64` `if status, known := statusOf(err); known {` | `journal.ClassifyHTTPMutation`, `statusOf` | — | 예 |
| B6 | if | `:67` `if outcome.Detail == "" {` | `err.Error` | — | 아니오 |
| B7 | else | `:69` `} else {` | `err.Error`, `journal.ClassifyHTTPMutation`, `reasonForClass` | :73, :78 | 예 |

## Calls and live bindings

`policyRefusal`(B2) · `ClassifyBrokerRefusal`(B3) · `errors.As`(B4) · `statusOf`(B5) · `journal.ClassifyHTTPMutation`(B5 안) · `reasonForClass`.

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

없다.

## Safety conclusion

- **Safe edit boundary**: **a094가 바꾸는 것은 이 함수가 아니라 B3이 부르는 `classifyRefusalBody`다.** 이 함수의 분기 구조와 순서는 그대로 둔다 — 그 순서가 이미 옳다(B3 주석: *the meaning comes from the answer and not from how far the bytes got*).
- **High-risk impact**: yes — 주문 분류의 최상위 진입점.

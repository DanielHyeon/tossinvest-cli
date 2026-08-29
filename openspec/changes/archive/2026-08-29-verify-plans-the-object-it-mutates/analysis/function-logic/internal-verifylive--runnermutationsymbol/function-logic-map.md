# Function Logic Map: `Runner.mutationSymbol`

- Source: `internal/verifylive/runner.go`
- Function: `internal/verifylive/runner.go:Runner.mutationSymbol`
- AST evidence: `ast.json` — 아래 분기 id·행 번호·호출은 여기서 읽었다
- Risk scan: `risk-pattern-report.md`
- Change: `verify-plans-the-object-it-mutates`

이 change의 중심. 한 단계의 라이브 요청이 **무엇에 대한 것인지** 답하는 유일한 자리다. 계획(plan.go)이 승인 줄을 쓸 때와 preflight가 시장을 검사할 때가 모두 여기를 읽고, 그 답은 단계 본문이 실제로 쓰는 대상과 같아야 한다. 2026-07-29에 같지 않았고 실행이 두 번 멈췄다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| step Step | 카탈로그 항목 | Steps() | 선언이 없으면 기존 경로(probe 종목)로 떨어진다 |
| r.liveConditional() | 기록의 outstanding 조건주문 | 기록(append-only) | 없으면 holdingSymbol로 폴백 |
| r.holdingSymbol | 이 실행의 보유 종목 | --holding-symbol 또는 계좌 | 비면 빈 문자열 — 호출부가 계획에서 제외한다 |
| r.symbol | 매수측 probe 종목 | --symbol | 무변경 경로 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` (line 572) — `if step.ActsOnConditional {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B2 | `if` (line 573) — `if _, symbol, ok := r.liveConditional(); ok && strings.TrimSpace(symbol) != "" {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B3 | `if` (line 578) — `if step.NeedsHolding {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.liveConditional` | ast.json calls (line 573) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `strings.TrimSpace` | ast.json calls (line 573) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |

라이브 바인딩 없음 — 이 함수는 브로커·네트워크를 직접 호출하지 않는다. 라이브 요청은
`mutate.go`가 이 파일의 판정을 통과한 뒤에만 보낸다.

## State mutations and fallbacks

- 없음 — 순수 조회. 기록도 계좌도 바꾸지 않는다.

## Safety conclusion

- Safe edit boundary: `ActsOnConditional` 분기를 맨 앞에 둔다. 그 아래 `NeedsHolding`·기본 경로는 바이트 단위로 종전과 같다.
- High-risk impact: yes — 승인 목록의 각 줄이 무엇에 대한 것인지를 결정한다. 틀리면 인가가 거절되고(현재 실패 방향) 요청은 전송되지 않는다.

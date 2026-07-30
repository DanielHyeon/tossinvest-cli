# Function Logic Map: `Runner.preflightStatic`

- Source: `internal/verifylive/runner.go`
- Function: `internal/verifylive/runner.go:Runner.preflightStatic`
- AST evidence: `ast.json` — 아래 분기 id·행 번호·호출은 여기서 읽었다
- Risk scan: `risk-pattern-report.md`
- Change: `verify-plans-the-object-it-mutates`

> **base revision** — 이 함수는 이 change가 수정하지 않았다. diff 문맥에만 걸렸고, AST는 `base-commit.txt`의 소스에서 뽑았다.

**이 change가 수정하지 않은 함수다.** base revision으로 고정한다. 바로 아래 `mutationSymbol`에 주석과 분기를 더하면서 diff 문맥에 걸렸을 뿐이고, 본문은 한 바이트도 바뀌지 않았다. 여기 있는 이유는 이 함수가 `mutationSymbol`의 두 호출부 중 하나이기 때문이다 — 시장 검사가 이제 조건주문의 종목으로 수행되며, 그것이 그 검사가 원래 물으려던 질문이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| step Step | 카탈로그 항목 | Steps() | deferred는 건너뛰지 않는다 |
| passed func(StepID) bool | 의존 단계 통과 여부 | 계획 시점은 willRun 포함 | 미통과면 skip |
| r.holdingSymbol | 보유 종목 | 계좌 | NeedsHolding 단계는 비면 skip |
| r.mutationSymbol(step) | 대상 종목 | runner.go | 다른 시장이면 skip |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` (line 531) — `if step.Deferred != "" {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B2 | `if` (line 537) — `if step.OptIn != "" && !r.optedIn(step) {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B3 | `range` (line 540) — `for _, dep := range step.DependsOn {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B4 | `if` (line 541) — `if !passed(dep) {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B5 | `if` (line 545) — `if step.NeedsHolding && r.holdingSymbol == "" {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B6 | `if` (line 550) — `if symbol := r.mutationSymbol(step); step.Mutates && !SameMarket(MarketOf(symbol), r.market) {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.optedIn` | ast.json calls (line 537) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `fmt.Sprintf` | ast.json calls (line 538) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `passed` | ast.json calls (line 541) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `r.mutationSymbol` | ast.json calls (line 550) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `SameMarket` | ast.json calls (line 550) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `MarketOf` | ast.json calls (line 550) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |

라이브 바인딩 없음 — 이 함수는 브로커·네트워크를 직접 호출하지 않는다. 라이브 요청은
`mutate.go`가 이 파일의 판정을 통과한 뒤에만 보낸다.

## State mutations and fallbacks

- 없음 — 판정만 한다.

## Safety conclusion

- Safe edit boundary: 무변경(base revision). 이 change는 이 함수의 입력 하나가 더 정확해지게 했을 뿐이다.
- High-risk impact: yes — mutating 단계가 자기 시장 밖 종목에 주문을 내지 못하게 막는다.

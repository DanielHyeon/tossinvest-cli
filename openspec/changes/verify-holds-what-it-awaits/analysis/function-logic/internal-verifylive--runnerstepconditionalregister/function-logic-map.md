# Function Logic Map: `Runner.stepConditionalRegister`

- Source: `internal/verifylive/steps.go`
- Function: `internal/verifylive/steps.go:Runner.stepConditionalRegister`
- AST evidence: `ast.json` — 아래 분기 id·행 번호·호출은 여기서 읽었다
- Risk scan: `risk-pattern-report.md`
- Change: `verify-holds-what-it-awaits`

SINGLE + MARKET 매도 조건주문을 등록하고 되읽어 측정한다(task 2.5). 이 change의 편집은 **끝의 표시 한 줄**이다 — `markDeliberate`가 `markHeld`가 되면서 gate(`conditional-cancel`)와 사슬을 함께 찍는다. 그 gate는 이 함수가 이미 따르던 정책을 명시적으로 적은 것이라 판정이 바뀌지 않는다. 사슬은 기록에 이미 있으면 물려받는데, **멱등 재생이 같은 식별자를 돌려주기 때문이다**(M14·M35 실측).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| r.holdingSymbol | 보유 종목 | 계좌 또는 --holding-symbol | 없으면 상위가 skip |
| id string | 브로커가 준 conditionalOrderId | POST 응답 | 빈 값이면 위쪽에서 fail |
| r.chainOf(...) | 기존 사슬 | 기록 | 비면 새 토큰을 만든다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` (line 558) — `if err != nil {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B2 | `if` (line 561) — `if sellable < MinQuantity {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B3 | `if` (line 568) — `if err != nil {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B4 | `if` (line 572) — `if err != nil {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B5 | `if` (line 587) — `if err != nil {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B6 | `switch` (line 600) — `switch {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B7 | `case` (line 601) — `case isGateError(replayErr):` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B8 | `case` (line 603) — `case replayErr != nil:` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B9 | `case` (line 606) — `case replayID == id:` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B10 | `case` (line 608) — `default:` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B11 | `if` (line 611) — `if replayID != "" {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B12 | `if` (line 614) — `if err := r.cancelConditional(ctx, sr, replayID, symbol,` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B13 | `if` (line 622) — `if co, err := r.readConditional(ctx, sr, id); err == nil {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B14 | `else` (line 629) — `} else {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B15 | `if` (line 642) — `if err != nil {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B16 | `else` (line 644) — `} else {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |
| B17 | `if` (line 652) — `if chain == "" {` | 없음 | 분기 계속 또는 반환 | 아래 참조 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.readSellable` | ast.json calls (line 557) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `sr.skip` | ast.json calls (line 562) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `fmt.Sprintf` | ast.json calls (line 562) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `trim` | ast.json calls (line 563) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `r.marketFrame` | ast.json calls (line 567) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `FarStopTrigger` | ast.json calls (line 571) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `MarketOf` | ast.json calls (line 571) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `newToken` | ast.json calls (line 581) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `r.expireDate` | ast.json calls (line 582) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `r.createConditional` | ast.json calls (line 586) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `sr.observe` | ast.json calls (line 590) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `r.sessionLabel` | ast.json calls (line 595) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `r.replayCreateConditional` | ast.json calls (line 599) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `isGateError` | ast.json calls (line 601) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `truncateError` | ast.json calls (line 605) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `sr.created` | ast.json calls (line 612) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `r.now` | ast.json calls (line 612) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `r.cancelConditional` | ast.json calls (line 614) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `r.readConditional` | ast.json calls (line 622) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `orDash` | ast.json calls (line 624) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `orNone` | ast.json calls (line 627) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `readRetry` | ast.json calls (line 636) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `r.broker.ConditionalOrders` | ast.json calls (line 639) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `len` | ast.json calls (line 641) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `strconv.FormatBool` | ast.json calls (line 646) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `containsConditional` | ast.json calls (line 646) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `r.chainOf` | ast.json calls (line 651) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |
| `sr.markHeld` | ast.json calls (line 655) | error 계약은 호출부에서 처리 — 이 함수는 브로커를 직접 부르지 않는다 | ast.json |

라이브 바인딩 없음 — 이 함수는 브로커·네트워크를 직접 호출하지 않는다. 라이브 요청은
`mutate.go`가 이 파일의 판정을 통과한 뒤에만 보낸다.

## State mutations and fallbacks

- 브로커에 조건주문을 만들고(승인 목록 경유) `sr.artifacts`에 기록한 뒤 표시한다.

## Safety conclusion

- Safe edit boundary: 표시 한 줄. 등록·조회·관측·멱등 재생 경로는 무변경이며 기존 단계 테스트가 그대로 통과한다.
- High-risk impact: yes — 라이브 조건주문을 만든다. 다만 이 change는 그 요청을 건드리지 않는다.

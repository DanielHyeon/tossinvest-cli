# Function Logic Map: `ExitObserver.clearTheSymbol`

- Source: `internal/app/engine/exitloop.go` (`1334`–`1392`)
- Qualified: `ExitObserver.clearTheSymbol`
- AST evidence: `ast.json` (`source_sha256` 6625c92061d5b05f…)
- Risk scan: `risk-pattern-report.md`
- 분기 9 · return 4 · 호출 16

**역할.** 청산과 충돌할 미체결 주문을 책에서 내리고, 종목이 깨끗해졌는지 답한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `live` | 치울 후보 목록 | **`Journal.LiveOrdersForSymbol` — 저널이다** | **a094가 바꾸는 지점** |
| `withPending` | 자기 방향 미체결까지 치우는가 (= `CancelPendingFirst`) | `record` `:1141` | B3·B8 |
| `order.Side` | 주문 방향 | 목록 | B3이 `buy`를 판별한다 |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/... -count=1 -covermode=set`의 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:1337` `if err != nil {` | `fmt.Errorf` | :1338 | 아니오 |
| B2 | range | `:1341` `for _, order := range live {` | `strings.EqualFold`, `strings.TrimSpace` | — | 예 |
| B3 | if | `:1343` `if !buy && !withPending {` | `costs.Market`, `fmt.Sprintf`, `o.opts.Issuer.IssueReduction`, `strings.ToLower`, `strings.ToUpper`, `strings.TrimSpace` | — | 아니오 |
| B4 | if | `:1358` `if err != nil {` | `floatOf` | — | 아니오 |
| B5 | if | `:1364` `if qerr != nil \|\| perr != nil {` | `o.opts.Submit.Cancel` | — | 아니오 |
| B6 | if | `:1379` `if err != nil \|\| out.State != journal.StateConfirmed {` | — | — | 예 |
| B7 | if | `:1383` `if !clear {` | — | :1384 | 예 |
| B8 | if | `:1386` `if withPending && m.state.Pending() {` | `m.state.Pending` | — | 예 |
| B9 | if | `:1387` `if err := o.release(ctx, m, journal.ProposalCancelled); err != nil {` | `o.release` | :1388, :1391 | 아니오 |

## Calls and live bindings

`o.opts.Journal.LiveOrdersForSymbol`(B1 앞) · `strings.EqualFold`(B2 안) · 취소 제출(B6 앞) · `o.release`(B9).

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

**브로커 취소 주문**(B6 앞) · 제안 해제(B9). 신규 주문도 정정도 내지 않는다.

## Safety conclusion

- **Safe edit boundary**: **a094가 바꾸는 것은 `live`의 원천 하나다.** B3의 술어(`!buy && !withPending → continue`), B6의 확정 규칙, B7의 `!clear → 제출 안 함`은 **전부 그대로 둔다.** 새 권한이 아니라 기존 권한의 눈을 넓히는 것이다(design D2).
- **High-risk impact**: yes — 브로커 취소를 내고, 그 실패가 손절 제출을 막는다.

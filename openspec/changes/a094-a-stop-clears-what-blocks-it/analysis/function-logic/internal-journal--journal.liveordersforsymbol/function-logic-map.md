# Function Logic Map: `Journal.LiveOrdersForSymbol`

- Source: `internal/journal/fills.go` (`1849`–`1927`)
- Qualified: `Journal.LiveOrdersForSymbol`
- AST evidence: `ast.json` (`source_sha256` 2d0c2175810b5067…)
- Risk scan: `risk-pattern-report.md`
- 분기 7 · return 6 · 호출 15

**역할.** 이 계좌·시장·종목에서 아직 살아 있는, **엔진이 낸** 주문을 돌려준다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `accountRef/market/symbol` | 조회 범위 | 호출자 | B1이 계좌 정체를 먼저 지킨다 |
| `mutation_attempts ⋈ intents` | 원천 | **원장. 브로커가 아니다** | **엔진 밖 주문은 0행** |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/... -count=1 -covermode=set`의 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:1851` `if err := j.guardTrackedFillIdentity(ctx, accountRef); err != nil {` | `j.db.QueryContext`, `j.guardTrackedFillIdentity`, `normaliseMarket`, `normaliseSymbol`, `string` | :1852 | 예 |
| B2 | if | `:1894` `if err != nil {` | `fmt.Errorf`, `rows.Close` | :1895 | 아니오 |
| B3 | for | `:1900` `for rows.Next() {` | `rows.Next` | — | 예 |
| B4 | if | `:1902` `if err := rows.Scan(&o.OrderID, &o.IntentID, &o.AccountRef, &o.Market, &o.TradingDay,` | `append`, `fmt.Errorf`, `rows.Scan` | :1905 | — |
| B5 | if | `:1909` `if err := rows.Err(); err != nil {` | `fmt.Errorf`, `rows.Err` | :1910 | 예 |
| B6 | range | `:1913` `for i := range out {` | `j.ResolveCurrentOrderIDScoped` | — | 예 |
| B7 | if | `:1921` `if err != nil {` | — | :1922, :1926 | 아니오 |

## Calls and live bindings

`j.guardTrackedFillIdentity`(B1) · `j.db.QueryContext` · `rows.Scan`(B4) · lineage 해소(B6 안).

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

없다 — 읽기 전용 질의다.

## Safety conclusion

- **Safe edit boundary**: a094는 이 함수의 **의미를 바꾸지 않는다** — 여기는 계속 '엔진이 낸 주문'을 답한다. 브로커 미체결은 **호출자(`clearTheSymbol`)가 합류**시킨다. 이 함수에 브로커 조회를 넣으면 원장 질의가 네트워크에 의존하게 되고, 그 실패가 원장 읽기의 실패가 된다.
- **High-risk impact**: yes — 이 결과가 무엇을 취소할지 정한다.

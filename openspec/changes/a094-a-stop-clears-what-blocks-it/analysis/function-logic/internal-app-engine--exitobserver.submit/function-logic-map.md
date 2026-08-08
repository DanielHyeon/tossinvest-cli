# Function Logic Map: `ExitObserver.submit`

- Source: `internal/app/engine/exitloop.go` (`1237`–`1312`)
- Qualified: `ExitObserver.submit`
- AST evidence: `ast.json` (`source_sha256` 6625c92061d5b05f…)
- Risk scan: `risk-pattern-report.md`
- 분기 11 · return 9 · 호출 24

**역할.** 확정된 청산 제안을 주문으로 만들어 제출하고, 결과에 따라 제안을 걸어 두거나 해제한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `proposal` | 레벨과 동작 | `exitpolicy` | B2가 0수량을 먼저 거른다 |
| `m.position` | 포지션 | 원장 | B4가 exit intent를 붙인다 |
| `out.State` | 제출 결과 상태 | `execgw` | **B7~B10이 이것으로 갈린다** |
| `out.Reason` | 거절 사유 | `execgw` | B9가 `ReasonSymbolInFlight`를 본다 |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/... -count=1 -covermode=set`의 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:1240` `if err != nil {` | — | :1241 | 아니오 |
| B2 | if | `:1243` `if isZeroQuantity(submitQuantity) {` | `costs.Market`, `fmt.Sprintf`, `isZeroQuantity`, `o.opts.Issuer.IssueReduction`, `o.release`, `strings.ToLower`, `strings.TrimSpace` | :1248 | 예 |
| B3 | if | `:1263` `if err != nil {` | `err.Error`, `o.alertProposalRefused`, `o.release` | :1265 | 아니오 |
| B4 | if | `:1272` `if err := o.opts.Journal.AttachExitIntent(ctx, m.position.ID, intentID); err != nil {` | `fmt.Errorf`, `o.opts.Journal.AttachExitIntent`, `o.sellIntent` | :1273 | 예 |
| B5 | if | `:1277` `if err != nil {` | `o.alertRefused`, `o.opts.Submit.Place`, `o.release` | :1279 | 아니오 |
| B6 | switch | `:1287` `switch {` | — | — | — |
| B7 | case | `:1288` `case out.State == journal.StateConfirmed:` | `o.log`, `string` | :1295 | 예 |
| B8 | case | `:1296` `case out.State == journal.StateInDoubt \|\| out.State == journal.StateUnresolvedInDoubt:` | — | :1300 | 예 |
| B9 | case | `:1301` `case out.Reason == execgw.ReasonSymbolInFlight:` | `o.noteDelay`, `o.release` | :1303 | 아니오 |
| B10 | case | `:1304` `default:` | — | — | 예 |
| B11 | if | `:1306` `if detail == "" && err != nil {` | `err.Error`, `o.alertProposalRefused`, `o.release` | :1310 | 아니오 |

## Calls and live bindings

`o.applyFloor`(B1 앞) · `isZeroQuantity`(B2) · `o.release`(B2·B5·B9·B10) · `o.opts.Issuer.IssueReduction` · `o.opts.Journal.AttachExitIntent`(B4) · `o.sellIntent`(B5 앞) · **`o.opts.Submit.Place`(B6 앞 — 유일한 브로커 mutation)** · `o.log` · `o.alertRefused`/`o.alertProposalRefused`.

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

`AttachExitIntent`(원장) · `Place`(**브로커 주문**) · `release`(제안 해제). B8은 **아무것도 하지 않고 반환한다** — 제안이 걸린 채 남는 것이 그 결과다.

## Safety conclusion

- **Safe edit boundary**: a094는 **B8의 판단을 바꾸지 않는다** — 팔렸을지 모르는 주문 위에 두 번째 매도를 올리지 않는 것은 옳다. 바꾸는 것은 (1) R1이 409를 B8이 아니라 **B10**으로 보내는 것, (2) R4가 `:1281`의 `PlaceRequest`에 `Baseline`을 싣는 것이다. **B8 자체는 그대로다.**
- **High-risk impact**: yes — 손절 주문이 실제로 나가는 자리. §0.3 직접 적용.

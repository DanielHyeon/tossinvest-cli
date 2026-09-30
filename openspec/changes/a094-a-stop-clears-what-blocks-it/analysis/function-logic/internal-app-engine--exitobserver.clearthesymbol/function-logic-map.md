# Function Logic Map: `ExitObserver.clearTheSymbol`

- Source: `internal/app/engine/exitloop.go` (`1482`–`1574`)
- Qualified: `ExitObserver.clearTheSymbol`
- AST evidence: `ast.json` (`source_sha256` 9b7d9a800069b100…) — 구현 로트(2026-09-30) 편집 뒤. 편집 전 AST 는 `analysis/implementation/pre-edit/`
- Risk scan: `risk-pattern-report.md`
- 분기 17

**역할.** 청산과 충돌할 미체결 주문을 치우고, 보호 청산을 제출해도 되는지(와 왜 아닌지)를 돌려준다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `live` | 원장의 미체결 목록(엔진 귀속 주문만) | `Journal.LiveOrdersForSymbol` | 읽기 실패는 오류 반환(B1, 종전) |
| `unsettled` | 같은 종목 미종결 attempt | `Submit.UnsettledOnSymbol` = `Gateway.unsettledFor`(주문 경로와 같은 판정) | 읽기 실패는 오류 반환(B2) |
| `withPending` | `CancelPendingFirst` | `record` | B5 · B14 |
| `m.state.PendingIntentID` | 무장된 발의의 intent | `exit_states` | 비면 해제하지 않음(B15) |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 `go test ./internal/<pkg>/ -count=1 -covermode=set` 프로파일(2026-09-30, `analysis/harness/flm_tables.py`)에서 그 줄로 시작하는 블록의 count.

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:1485` `if err != nil {` | 예 |
| B2 | if | `:1489` `if err != nil {` | 아니오 |
| B3 | if | `:1492` `if len(unsettled) > 0 {` | 예 |
| B4 | range | `:1498` `for _, order := range live {` | 예 |
| B5 | if | `:1500` `if !buy && !withPending {` | 예 |
| B6 | if | `:1503` `if !buy {` | 예 |
| B7 | if | `:1506` `if err != nil {` | 아니오 |
| B8 | if | `:1510` `if waiting {` | 예 |
| B9 | if | `:1527` `if err != nil {` | 아니오 |
| B10 | if | `:1533` `if qerr != nil \|\| perr != nil {` | 아니오 |
| B11 | if | `:1548` `if err != nil \|\| out.State != journal.StateConfirmed {` | 예 |
| B12 | if | `:1552` `if !buy {` | 예 |
| B13 | if | `:1557` `if !res.cleared {` | 예 |
| B14 | if | `:1560` `if withPending && m.state.Pending() {` | 예 |
| B15 | if | `:1561` `if strings.TrimSpace(m.state.PendingIntentID) == "" {` | 아니오 |
| B16 | if | `:1566` `if err != nil {` | 아니오 |
| B17 | if | `:1569` `if !released {` | 예 |

## Calls and live bindings

`LiveOrdersForSymbol` · `Submit.UnsettledOnSymbol` · `Journal.ConfirmedCancelOf`(매도마다) · `Issuer.IssueReduction` · `floatOf` · `workingOrderPrice`(빈 가격 0) · `Submit.Cancel`(유일한 브로커 mutation — 취소만) · `Journal.ReleaseClearedExitProposal`. 새 브로커 **조회**는 0 — 추가된 호출은 전부 원장 읽기다.

## State mutations and fallbacks

브로커 취소(B11 앞) · 발의 해제(B16 뒤 — 원장 판정이 허락할 때만). 신규 · 정정 주문 없음.

## Safety conclusion

- a094 D−4.3 · D−4.7 · D−3.2. 제출을 여는 `cleared=true` 는 (1) 같은 종목 미종결 0, (2) 치운 주문이 전부 매수이거나 치울 것이 없고, (3) 발의가 없거나 원장 판정이 해제를 허락했을 때만이다. 매도 취소 접수 · 확정 취소 대기 매도 · park/모호/접수 대기 발의는 전부 미완료. High-risk: yes — 그 판정이 손절 제출을 연다/막는다.

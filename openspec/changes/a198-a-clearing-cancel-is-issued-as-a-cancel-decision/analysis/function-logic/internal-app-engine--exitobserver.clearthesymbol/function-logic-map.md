# Function Logic Map: `ExitObserver.clearTheSymbol`

- Source: `internal/app/engine/exitloop.go:1649-1743`
- AST evidence: `ast.json` (Pre-Edit HEAD `e1eaf720c`, source_sha256 `bc81b5fd…`, 분기 17 · return 8 · 호출 25)
- Risk scan: `risk-pattern-report.md` (적중 1: exitloop.go:2046 panic — 이 함수 밖)
- 판: **pre-edit**. 편집 범위(design D1·D2 (가-B)·D4, Manager 결정 2026-11-23): B9 앞의 결정 발행 호출(L1684-1695)을 관측기 안에서 만든 `execgw.Issuer{Journal: o.opts.Journal, Clock: o.clk, AccountRef: o.opts.AccountRef}` 의 `IssueReduction`(종류 CANCEL, 방향·수량 = 주문)으로, B9·B11 의 실패 갈래에 새 함수 `noteClearCancelFailure` 호출 추가(B10 은 기록 없음, D4 재결정). 판정(`res.cleared=false; continue`)과 다른 분기는 무편집.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `m.position`(Market·Symbol·ID) | 관리 중 포지션 | 원장 exit 상태 | — |
| `withPending` | `snapshot.CancelPendingFirst` | record L1395 | 거짓이면 매도는 대상 아님(B8) |
| `live` | 엔진 귀속 CONFIRMED PLACE/AMEND 중 종결 증거 없는 주문(`LiveOrdersForSymbol`, fills.go:1854). `Quantity` = intent 수량(원/정정 후, 잔량 아님) | 원장 | 읽기 실패 → 오류 반환(B1) |
| `unsettled` | 같은 시장·종목의 RECORDED·DISPATCH_STARTED·ACKED·IN_DOUBT attempt(`Submit.UnsettledOnSymbol` = 게이트웨이 `unsettledFor`) | 원장 | 읽기 실패 → 오류 반환(B2) |
| `o.opts.Issuer` | 운영 = `RiskGuardian`(exitwiring.go:339) — 편집 뒤 이 함수에서는 안 씀(submit 만) | 조립 | 발행 실패 → 그 주문 미치움(B9) |
| `o.opts.Journal`·`o.clk`·`o.opts.AccountRef` | 편집 뒤 취소 결정 발행기의 재료. 운영에서 `o.clk` = 게이트웨이 시계(claim-check 「D2 선행 확인」) | NewExitObserver | — |
| `o.opts.Submit` | 운영 = `*execgw.Gateway` | 조립 | 오류·비확정 → 미치움(B11) |
| 불변식 | 하나라도 확정 취소에 실패하면 cleared=false(정본). 매도 취소 접수는 치움이 아님. 매수 취소 접수는 치움. 같은 종목 미종결 attempt 가 있으면 미완료 | exit-policy 정본 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 L1652 | LiveOrdersForSymbol 오류 | 없음 | `clearResult{}`, 오류 | 기존 |
| B2 L1656 | UnsettledOnSymbol 오류 | 없음 | `clearResult{}`, 오류 | 기존 없음(선택 실행 미덮음) |
| B3 L1659 | 같은 종목 미종결 ≥1 | 없음 | `{countable: !onlyEngineCancelsInFlight(...)}` — cleared=false | 기존 + a198 §4-6(실제 IN_DOUBT 취소) |
| B4 L1665 | 미체결 주문 순회 | — | — | 기존 |
| B5 L1667 | 매도 주문 | 원장 읽기 ConfirmedCancelOf | — | 기존 + a198 §4-3 |
| B6 L1671 | ConfirmedCancelOf 오류 | `warnEpisode`(운영 무기록 — opts.Log nil), cleared=false | continue | 기존 없음 |
| B7 L1676 | 엔진 취소가 이미 CONFIRMED | cleared=false, awaitingClose | continue | 기존(직접 쓴 행) + a198 §4-3(게이트웨이가 쓴 행) |
| B8 L1680 | 매도 && !withPending | — | continue(대상 아님) | 기존 |
| B9 L1696 | **결정 발행 오류** — 편집 대상(D1 발행 · D4 기록) | 지금: cleared=false 만 | continue | 기존 없음 · a198 §4-5 |
| B10 L1702 | 수량·가격 해석 실패 | cleared=false 만(D4 기록 대상 여부 미정) | continue | 기존 없음(a087 핀은 헬퍼 값만) |
| B11 L1717 | **취소 오류 또는 CONFIRMED 아님** — 편집 대상(D4) | 지금: cleared=false 만 | continue | 기존(가짜 제출자 cancelFails) · a198 §4-2·4-4·4-6 |
| B12 L1721 | 취소 접수된 매도 | cleared=false, awaitingClose | — | 기존 + a198 §4-3 |
| B13 L1726 | 미치움 | — | `res` | 기존 |
| B14 L1729 | withPending && 발의 있음 | — | — | 기존 |
| B15 L1730 | 발의 intent 없음 | — | `{countable:true}` | 기존 없음(선택 실행) |
| B16 L1735 | ReleaseClearedExitProposal 오류 | — | 오류 | 기존 없음 |
| B17 L1738 | 해제 안 됨 | — | `{countable, awaitingClose: verdict==AwaitingClose}` | 기존 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Journal.LiveOrdersForSymbol` | 치울 대상 | 오류 → B1 반환 | CodeGraph callees + AST L1650 |
| `Submit.UnsettledOnSymbol` | 주문 경로와 같은 미종결 판정(a094 D−4.7) | 오류 → B2 | AST L1655 |
| `onlyEngineCancelsInFlight`·`clearTargets` | 계수 제외 판정 | 순수 | AST L1662 |
| `Journal.ConfirmedCancelOf` | 매도 취소 접수 여부(kind=CANCEL, CONFIRMED, target_order_id, 취소 intent 의 거래일 = 주문 거래일) | 오류 → B6 | AST L1670 |
| `Issuer.IssueReduction` | **지금 = 매도·KindPlace 결정**(RiskGuardian B2·L614) → 취소에 키가 실려 Prepare 거절 | 오류 → B9 | AST L1684, probe-output.txt |
| `floatOf`·`workingOrderPrice` | 요청 수량·가격 | 오류 → B10 | AST L1700-1701 |
| `Submit.Cancel` | 게이트웨이 취소(claim-check (a) 의 가드 열둘) | 오류·비확정 → B11. 엔진 정책 409 재전송 0 | AST L1706 |
| `Journal.ReleaseClearedExitProposal` | 발의 해제 판정 | 오류 → B16 | AST L1734 |
| `warnEpisode` | B6 의 기록 | **운영 무기록**(`o.log` L1963 opts.Log nil) | exit_held_proposal.go:336 |
| live config | `ctx` 의 계좌 예산 우선순위(record L1392-1394, 보호 청산 = Critical) | a166 admission 대기 가능 | record AST B10 |

## State mutations and fallbacks

- 함수 자신은 메모리 상태를 바꾸지 않음 — `clearResult` 값만 만듦. 부작용은 결정 행 기록(Issuer)·취소 attempt 기록과 브로커 취소(Submit.Cancel)·발의 해제(ReleaseClearedExitProposal).
- 폴백 없음: 모든 실패는 「이 주문 미치움 → 그 주기 보호 청산 미제출」로 수렴(보수 방향).

## Safety conclusion

- Safe edit boundary: L1684-1695(발행 호출 인자·발행기) 와 B9·B11 본문에 `noteClearCancelFailure` 호출 한 줄씩. `res.cleared=false; continue` 와 B1-B8·B12-B17 은 무편집.
- High-risk impact: **yes** — 손절 제출을 막는 판정 함수이고, 이 편집으로 브로커 취소가 처음 나간다(claim-check §2-1 재취소 위험 포함).

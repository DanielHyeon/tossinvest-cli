# Function Logic Map: `ExitObserver.clearTheSymbol`

- Source: `internal/app/engine/exitloop.go:1651-1748`
- AST evidence: `ast.json` (**post-edit**, 작업 트리 — source_sha256 `6ffb1ec1…`, 분기 17 · return 8 · 호출 30). Pre-edit 판은 HEAD `e1eaf720c` 의 `analysis/cited-ast/ExitObserver_clearTheSymbol.ast.json`(sha `bc81b5fd…`).
- Risk scan: `risk-pattern-report.md`
- 재번호(옛/새 ast 분기를 difflib 으로 정렬): B1~B8 동일 · 옛 B10(수량·가격 해석 실패) → **새 B9**(발행 앞으로 이동) · 옛 B9(결정 발행 실패) → **새 B10** · B11~B17 동일. 분기 수 17 → 17.
- 편집(design D1·D2 (가-B)·D4, Manager 결정 2026-11-23): ① 수량·가격 해석을 결정 발행 앞으로 옮김(최대 수량을 결정에 싣기 위함 — 해석 실패 시 결정 행을 남기지 않는 것 외에 판정 동일) ② 결정 발행을 `RiskGuardian.IssueReduction`(매도·PLACE) 에서 관측기 안의 `execgw.Issuer{Journal, Clock: o.clk, AccountRef}.IssueReduction`(CANCEL, 주문 방향·수량)으로 ③ 새 B10·B11 에 `noteClearCancelFailure` ④ 취소 접수 확정 뒤 `noteClearCancelConfirmed`. `res.cleared` 판정·`continue`·다른 분기 무편집.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `m.position`(Market·Symbol·ID) | 관리 중 포지션 | 원장 exit 상태 | — |
| `withPending` | `snapshot.CancelPendingFirst` | record | 거짓이면 매도는 대상 아님(B8) |
| `live` | 엔진 귀속 CONFIRMED PLACE/AMEND 중 종결 증거 없는 주문(`LiveOrdersForSymbol`). `Quantity` = intent 수량(잔량 아님) | 원장 | 읽기 실패 → 오류 반환(B1) |
| `unsettled` | 같은 시장·종목의 RECORDED·DISPATCH_STARTED·ACKED·IN_DOUBT attempt | 원장(`Submit.UnsettledOnSymbol`) | 읽기 실패 → 오류 반환(B2) |
| `o.opts.Journal`·`o.clk`·`o.opts.AccountRef` | 취소 결정 발행기의 재료. 운영에서 `o.clk` = 게이트웨이 시계 | NewExitObserver | 발행 실패 → 그 주문 미치움(B10) |
| `o.opts.Submit` | 운영 = `*execgw.Gateway` | 조립 | 오류·비확정 → 미치움(B11) |
| `o.opts.ExitOrderLog` | 운영 = 엔진 로거(cmd/tossctl/engine.go:680), nil 이면 실패 줄 없음 | 조립 | 기록만, 판정 무영향 |
| 불변식 | 하나라도 확정 취소에 실패하면 cleared=false. 매도 취소 접수는 치움이 아님. 매수 취소 접수는 치움. 같은 종목 미종결 attempt 가 있으면 미완료 | exit-policy 정본 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 L1654 | LiveOrdersForSymbol 오류 | 없음 | `clearResult{}`, 오류 | 기존 |
| B2 L1658 | UnsettledOnSymbol 오류 | 없음 | `clearResult{}`, 오류 | 기존 없음 |
| B3 L1661 | 같은 종목 미종결 ≥1 | 없음 | `{countable: !onlyEngineCancelsInFlight}` | a198 InDoubt·Acked |
| B4 L1667 | 미체결 순회 | — | — | 기존 |
| B5 L1669 | 매도 주문 | ConfirmedCancelOf 읽기 | — | a198 ArmedSell |
| B6 L1673 | ConfirmedCancelOf 오류 | warnEpisode(운영 무기록), cleared=false | continue | 기존 없음 |
| B7 L1678 | 엔진 취소가 이미 CONFIRMED | cleared=false, awaitingClose | continue | a198 ArmedSell(게이트웨이가 쓴 행) |
| B8 L1682 | 매도 && !withPending | — | continue | 기존 |
| B9 L1688 | 수량·가격 해석 실패(옛 B10) | cleared=false | continue | 기존 없음 |
| B10 L1704 | **취소 결정 발행 실패**(옛 B9) | `noteClearCancelFailure(…, "", cancel_decision_failed)`, cleared=false | continue | a198 AnUnrecordableCancelDecision |
| B11 L1720 | **취소 오류 또는 CONFIRMED 아님** | `noteClearCancelFailure(…, out.State, out.Reason)`, cleared=false | continue | a198 ARefusedCancel · AnInDoubtCancel |
| (B11 거짓) | 접수 확정 | `noteClearCancelConfirmed` | — | a198 AWorkingEngineBuy |
| B12 L1726 | 접수된 매도 취소 | cleared=false, awaitingClose | — | a198 ArmedSell |
| B13 L1731 | 미치움 | — | `res` | 기존 |
| B14 L1734 | withPending && 발의 있음 | — | — | 기존 |
| B15 L1735 | 발의 intent 없음 | — | `{countable:true}` | 기존 없음 |
| B16 L1740 | ReleaseClearedExitProposal 오류 | — | 오류 | 기존 없음 |
| B17 L1743 | 해제 안 됨 | — | `{countable, awaitingClose}` | 기존 · a198 ArmedSell |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Journal.LiveOrdersForSymbol` | 치울 대상 | 오류 → B1 | AST |
| `Submit.UnsettledOnSymbol` | 주문 경로와 같은 미종결 판정 | 오류 → B2 | AST |
| `Journal.ConfirmedCancelOf` | 매도 취소 접수 여부 | 오류 → B6 | AST |
| `floatOf`·`workingOrderPrice` | 요청·결정 수량, 요청 가격 | 오류 → B9 | AST |
| `execgw.Issuer.IssueReduction`(Kind CANCEL) | 취소 결정(위험 감소, 키 없음, preimage 방향 = 주문 방향) | 오류 → B10 | AST, a198 시험 |
| `Submit.Cancel` | 게이트웨이 취소(사전 단건 조회 + 취소 요청, 엔진 409 재전송 0) | 오류·비확정 → B11 | AST |
| `noteClearCancelFailure`·`noteClearCancelConfirmed` | 실패 줄·표시(a198_clear_cancel_log.go) | 기록만, recover 로 쓰기 패닉 격리 | AST |
| `Journal.ReleaseClearedExitProposal` | 발의 해제 | 오류 → B16 | AST |

## State mutations and fallbacks

- 관측기 메모리: `clearCancelLogged`(실패 줄 중복 억제 — 새 함수가 nil 이면 만듦, 접수 확정 시 그 주문 항목 삭제). 그 밖에 이 함수는 `clearResult` 값만 만듦.
- 부작용: 취소 결정 행(원장), 취소 attempt(원장) · 브로커 사전 조회 + 취소(게이트웨이), 발의 해제, 실패 줄(ExitOrderLog).
- 폴백 없음: 모든 실패는 「이 주문 미치움 → 그 주기 보호 청산 미제출」로 수렴.

## Safety conclusion

- Edit boundary(실제): 해석 두 줄 이동, 발행 호출 교체, 실패 기록 호출 2, 확정 표시 삭제 호출 1. 판정 대입·분기 조건 무편집.
- High-risk impact: **yes** — 이 편집으로 브로커 취소가 처음 나간다. 남은 위험은 design §2-5·§6.

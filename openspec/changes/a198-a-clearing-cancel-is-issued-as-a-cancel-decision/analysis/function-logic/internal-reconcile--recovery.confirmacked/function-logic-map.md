# Function Logic Map: `Recovery.confirmAcked`

- Source: `internal/reconcile/acked_boot.go:48-90`(post-edit)
- AST evidence: `ast.json` (**post-edit**, 4c.7 뒤 — source_sha256 `ac42ae8e…`, 분기 9). Pre-edit 판은 HEAD `a7639378e`(sha `4919ea34…`, 분기 6 — 아래 표의 번호, `analysis/cited-ast/d2_Recovery_confirmAcked.ast.json`).
- **재번호(pre → post)**: B1 동일(L50) · **새 B2 L56 `Kind == CANCEL && 번호 있음` · 새 B3 L58 `Resume` 성공 · 새 B4 L62 쓰기 실패**(D9) · 옛 B2~B6 → 새 B5~B9(조건·반환 무변). 아래 표는 pre-edit 번호다(BTM 은 post-edit 번호).
- Risk scan: `risk-pattern-report.md` (적중 0)
- 편집 범위(design D9, Manager 결정 2026-11-24): B1 뒤·B2 앞에 갈래 하나 — 종류 CANCEL 이고 기록 번호가 비어 있지 않으면 `Resume` → `ResolveConfirmed(기록 번호, ReasonBootAckedCancelCompleted, 상세)`, 성공이면 true(알림 0), 쓰기 실패면 명명 알림·false. 그 밖 분기 무편집.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `rec` | `Recovery.Run` 이 넘긴 ACKED attempt(recovery.go:279) | 원장 `PendingAttempts` | — |
| `rec.Kind` | PLACE · CANCEL · AMEND | 원장 | PLACE 외는 오늘 B2(알림만) |
| `rec.BrokerOrderID` | ACKED 면 `MarkAcked` 가 쓴 번호(dispatch.go:181). CANCEL 은 계약상 취소가 새로 받은 식별자(대상 번호 아님) | 원장 | 비면 B3 |
| 불변식 | 이 경로의 실패는 복구를 실패시키지 않음(false + 알림) · 바이트 일치 외 추론으로 ACKED 발주를 종결하지 않음(정본) | 정본 order-execution 「IN_DOUBT 해소」 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 L47 | intent 읽기 실패 | 알림 | false | 기존 없음 |
| B2 L51 | 종류 ≠ PLACE | 알림(「CANCEL/AMEND … stays ACKED」) | false | TestA094AnAckedCancelOrNumberlessPlaceIsOnlyNamed(cancel 행) |
| B3 L55 | 기록 번호 없음 | 알림 | false | 기존 없음 |
| B4 L59 | `ConfirmPlacedOrder` 실패 | 알림(본문 없는 상태 코드) | false | TestA094AnUnconfirmedAckedPlaceStaysAndIsNamed · TestA094TheAckedAlertKeyIsStableAcrossRestarts · TestA094TheAckedAlertWithholdsTheBrokerBody |
| B5 L64 | `Resume` 성공 | `ResolveConfirmed` | — | TestA094AnAckedPlaceIsConfirmedByItsRecordedNumber |
| B6 L68 | 확정 쓰기 실패 | 알림(「ledger write failed」) | false | TestA094AFailedConfirmWriteLeavesTheAckAndNamesIt |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Journal.LookupIntent` | 종목(알림·확인) | 오류 → B1 | AST |
| `alertAcked` | 명명 critical(창 0, key = attempt) | 기록 실패는 입구가 진입을 잠금 | AST |
| `execgw.ConfirmPlacedOrder` | 발주 번호 바이트 일치 읽기 | 3s 시한, 실패 → B4 | AST · TestA094BootAndPostPlaceShareOneConfirmation |
| `Journal.Resume` · `Attempt.ResolveConfirmed` | ACKED → CONFIRMED | 실패 → B6 | AST |

## State mutations and fallbacks

- 원장: ACKED → CONFIRMED(B5 경로만). 그 밖은 알림 기록뿐. 폴백 없음 — 실패는 ACKED 유지 + 알림, 복구는 계속.
- High-risk: yes(기동 복구·원장 전이).

## Safety conclusion

- Edit boundary(실제): B1 뒤에 갈래 하나(B2~B4) — CANCEL + 기록 번호면 `Resume`·`ResolveConfirmed(기록 번호, boot_acked_cancel_completed)`, 실패면 기존과 같은 명명 알림·false. 옛 분기(B5~B9) 무편집, 브로커 호출 0 추가.
- 귀결: ACKED 취소는 디스패치가 하려던 쓰기 그대로 CONFIRMED — 매도는 종결 증거 대기, 매수는 D8 로 치움(정상 CONFIRMED 와 같음). 실패는 ACKED 유지·알림·복구 성공(복구를 실패시키지 않음 — 불변식 그대로).
- 남은 위험: 접수 확정은 취소 완료 증명이 아님(CANCEL_REJECTED 가능성은 정상 경로와 같음). design §6.

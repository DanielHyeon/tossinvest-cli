# a087 tasks 3.6 — MARKET 보호 청산의 부분체결 잔량

- 측정 순간: 2026-10-22, HEAD `0686c0f0a`. AST: `ast/`(`tools/logic-map`, 생성 시 `source_sha256` 일치). 브로커 거동은 실측 0
  (§5.1 은 1주 전량 `체결완료` 한 건 — 부분체결을 관측한 적 없음).
- **판정: 체결 감지가 잔량을 「닫는」 방법은 종결 관측뿐이고, 잔량의 재제출은 청산 루프가 종결 뒤에만 함. 잔량이 브로커에서
  종결되지 않고 남는 동안에는 시간 상한·지연 경보가 없음 → 갭, `issues.md` I-R9.**

## 잔량이 남는 동안 (주문 비종결)

| 단계 | 근거 |
| --- | --- |
| 체결 감지는 그 주문을 매 주기 읽음 — 비종결 스냅숏은 추적 집합에 남음 | `Journal.TrackedFillOrders`(`internal/journal/fills.go:1543`) 질의의 `f.terminal = 0`; 종결 판정은 `brokerstate.State.IsTerminal`(`derive.go:94-102` B2: FILLED·CANCELLED·CANCELLED_PARTIALLY_FILLED·REPLACED·REJECTED·REJECTED_PARTIALLY_FILLED) — PARTIAL_FILLED 는 비종결 |
| 체결분만 포지션에 반영되고 발의는 무장 유지 | `journal.ApplyExitFill`(`apply_hook.go:971-1041`) B11(`:1009` `fill.Terminal && state.Pending() && … PendingIntentID == where.IntentID`) — 비종결이면 `ResolvePending` 안 함. 기존 시험 `TestANonTerminalFillDoesNotResolveTheProposal`(`exit_state_test.go:572`) |
| 같은 보호 행동의 재발의는 억제됨 | `exitpolicy.EvaluateRatchet` B17(`ratchet.go:423` `PendingAction == ActionBaselineBreach` → `SuppressedPending`, 발의 없음) · `EvaluateLadder` B26(`ladder.go:441` `PendingAction == ActionLadderStop`) |
| 억제된 판정은 청소·제출·지연 시계에 닿지 않음 | `ExitObserver.record`: `orderable := snapshot.Orderable && !proposal.Zero()`(`exitloop.go:1314`) — 억제 결정은 발의가 비어 orderable 거짓. `noteDelay`(`:1910`)의 호출 자리는 `record` B11(`:1387` 청소 실패)·`submit` B10(`:1567` 같은 종목 진행 중) 둘뿐 |
| 무장 발의 알림은 엔진 **취소**가 확정된 매도만 셈 | `ExitObserver.noteHeldProposal`(`exit_held_proposal.go:126-173`) → `noteAwaitingClose`(`:177-196`) B2(`:183` `!found`) — `ConfirmedCancelOf` 가 없으면 반환 — 엔진이 취소하지 않은 부분체결 잔량은 대상 아님 |

## 잔량이 브로커에서 끝나면 (CANCELED + canceledAt, 또는 FILLED)

| 단계 | 근거 |
| --- | --- |
| 체결 감지가 id 로 읽어 종결 스냅숏 `CANCELLED_PARTIALLY_FILLED` 를 남기고, 체결 수량은 늘지 않으며, 추적 집합에서 빠짐 | 신규 시험 `TestA087MarketSellRemainderIsTrackedUntilTheBrokerEndsIt`(`internal/filldetect/a087_market_exit_fill_pin_test.go`) — 변이 M11(비종결을 추적에서 뺌) CAUGHT |
| 종결 관측이 발의를 풂(`PROPOSAL_FILLED` 이벤트) | `ApplyExitFill` B11·B12(`apply_hook.go:1009-1013`) — 조건에 행동 종류가 없으므로 보호·익절 같음. 기존 시험 `TestAPartialSellMovesTheTakenFractionAndResolvesTheProposal`(`exit_state_test.go:478`, 익절 발의로 같은 갈래) |
| 남은 포지션은 다음 관측에서 다시 판정 → 이탈 지속이면 새 보호 발의, 수량은 그때의 투영 수량 | 억제 조건(위 B17·B26)이 풀림; `record` 의 수량 = `snapshot.ProjectedQuantity`(`exitloop.go:1313`). **종단 시험 없음** — 「잔량 종결 → 다음 주기 재발의 수량 = 잔량」을 엔진 수준에서 묶은 시험은 이 로트가 만들지 않았음(미검증) |

## 갭 (→ `issues.md` I-R9)

MARKET 보호 청산이 부분체결된 뒤 잔량이 브로커에 **비종결로 남으면**(예: 매수 호가 공백 — design D8 1행) 그 포지션의 보호는 그 주문
하나에 묶이고, 재발의·재가격·지연 경보가 모두 없음. KR 시장가 잔량을 브로커가 어떻게 처리하는지(즉시 취소/잔존)는 [미측정].
재가격 에스컬레이션은 무소유 장부(tasks 「후속 change」 2행)이고, 이 change 는 production 을 바꾸지 않음.

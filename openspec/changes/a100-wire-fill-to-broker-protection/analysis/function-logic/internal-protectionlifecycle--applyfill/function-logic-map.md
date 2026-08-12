# Function Logic Map: `applyFill`

- Source: `internal/protectionlifecycle/lifecycle.go` (L235-271)
- AST evidence: `ast.json` — 분기 7, return 7, 호출 12
- Risk scan: `risk-pattern-report.md`

> **이 함수가 a100의 진입점이다.** 설계 문서 §8.2의 1단계("체결 delta 수신")가 여기다.
> 현재 외부 non-test importer가 0이므로 프로덕션에서 한 번도 호출되지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `state` | seal 유효 | `validState(state)` | 무효면 `RefusalInvalidState` |
| `key` | 존재하는 position | `state.positions` | `mutablePosition` 에러 (단 `RefusalInvalidObservation`은 통과시킨다) |
| `fill.FillID`, `fill.Fingerprint` | `validIdentity` | 호출부 | 불일치 시 `RefusalInvalidObservation` |
| `fill.BrokerOrderID` | 비어 있지 않고 `position.Observed.BrokerOrderID`와 **정확히 일치** | 브로커 관측 | 불일치 시 거부 |
| `fill.Quantity` | `0 < q <= Observed.Quantity` **and** `<= Holdings` | 브로커 관측 | 초과 시 `RefusalFillExceeded` |

**불변식 1**: 모든 실패 경로가 `FillResult{PreserveExit: true}`를 돌려준다 — 체결 처리에 실패해도
**청산 경로는 보존된다.** 안전 불변식 §0-3의 구현.
**불변식 2**: 성공 경로도 `PreserveExit: true`다. 즉 이 함수는 청산을 절대 막지 않는다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error |
|---|---|---|---|
| B1 (L237) | `err != nil && errorCode(err) != RefusalInvalidObservation` | 없음 | `(state, {PreserveExit}, err)` |
| B2 (L240) | `!validState(state)` | 없음 | `RefusalInvalidState` |
| B3 (L244) | fill 식별자 무효 또는 `BrokerOrderID` 불일치 | 없음 | `RefusalInvalidObservation` |
| B4 (L247) | `fill.FillID`가 이미 존재 | (B5 미해당 시) `latchPosition(RefusalConflictingFill)` | `RefusalConflictingFill` — **latch 발생** |
| B5 (L248) | 기존 fill과 수량·fingerprint 동일 | 없음 | `{Duplicate: true, PreserveExit: true}, nil` — 멱등 |
| B6 (L254) | `Quantity == 0` 또는 claim 초과 | 없음 | `RefusalFillExceeded` |
| B7 (L264) | `position.Observed.Quantity == 0` | `Status/Phase → BrokerFilled/Terminal`, `HasPending=false` | (계속) |
| fall-through (L270) | 정상 | 아래 "State mutations" 전부 | `{Applied: true, PreserveExit: true}, nil` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry 계약 | Evidence |
|---|---|---|---|
| `mutablePosition` | position 취득 + capability 검사 | error. `RefusalInvalidObservation`은 **의도적으로 통과** | AST L236 |
| `validState` | seal 검증 | bool | AST L240 |
| `validIdentity` | fill 식별자 검증 | bool, 2회 호출 | AST L244 |
| `latchPosition` | 충돌 fill에 latch | 새 state 반환 | AST L251 |
| `cloneState` / `next.reseal` | 불변 state 갱신 | 순수 | AST L257 / L269 |

## State mutations and fallbacks

성공 경로에서만 mutation이 일어나며 전부 `cloneState` 사본 위에서 수행된 뒤 `reseal`된다.

- `position.Fills[fill.FillID] = {Quantity, Fingerprint}`
- `Holdings -= fill.Quantity`
- `Observed.Quantity -= fill.Quantity`, `Desired.Quantity = Observed.Quantity`
- `LifecycleRevision++`
- B7이면 `Observed.Status/Desired.Status = BrokerFilled`, `Phase = Terminal`, `HasPending = false`
- `EntryOpen`을 **5개 조건의 논리곱**으로 재계산 —
  `Phase == Active` ∧ `EntryLatch == ""` ∧ `marketLatches[Market] == ""` ∧
  `Observed.Status == BrokerActive` ∧ `Observed.Quantity + OtherSellClaims == Holdings`

**마지막 조건이 설계 문서 §8.2의 4단계(보호 수량 = 보유 수량)에 해당한다.**
불일치하면 `EntryOpen = false`가 되어 진입이 닫힌다 — §8.2의 5단계가 이미 구현돼 있다.

fallback 없음. 모든 실패는 fail-closed다.

## Safety conclusion

- Safe edit boundary: **본문을 편집할 이유가 없다.** a100은 이 함수를 *호출하는* 경로를 만든다.
  단 아래 테스트 공백을 먼저 메워야 프로덕션 배선이 정당화된다.
- High-risk impact: **yes** — 체결 감지와 보호 수량 정합의 핵심.

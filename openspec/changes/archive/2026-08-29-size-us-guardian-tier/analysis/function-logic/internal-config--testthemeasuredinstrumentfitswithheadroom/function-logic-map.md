# Function Logic Map: `TestTheMeasuredInstrumentFitsWithHeadroom`

- Source: `internal/config/limits_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Narrative context: `../../function-logic-map.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `measuredShare` = 300 | 관측값 | `verify-execution-capability/measurements.md` M49 (2026-07-30, US 정규장, TSLA 299.88–299.94) | 상한이 이보다 낮으면 `t.Fatalf` |
| 최소 헤드룸 50% | design D1 하한 논증 | SINGLE+MARKET은 체결가를 모른다 | 미달은 `t.Errorf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | USD 상한 조회 오류 | 없음 | `t.Fatalf` | `TestAnUnregisteredCurrencyFailsClosed` |
| B2 | 상한 < 관측 1주가 | 없음 | `t.Fatalf` | 자기 자신 |
| B3 | 헤드룸 < 50% | 없음 | `t.Errorf` | 자기 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `GuardianCeiling` | USD 주문 상한 | 미등록 통화는 error | CodeGraph + AST |

## State mutations and fallbacks

- 상태 변경 없음. 신규 테스트다.
- 이 테스트는 실측을 코드에 **인용**한다. `risk-management`의 provenance 요구가 TossOS 실측 출처에 요구하는 것이 정확히 그것이다 — 서술이 아니라 식별자다.
- **이 테스트는 시장을 보지 않는다.** `measuredShare`는 2026-07-30 스냅샷으로 고정된 상수이므로, 종목 가격이 올라도 여기는 조용하다. 빨개지는 경우는 하나뿐이다 — 누군가 USD 주문 상한을 낮춰 기록된 관측 위의 여유를 없앨 때. 즉 이것은 시세 감시가 아니라 **되돌림 방지**이고, 가격 변동에 따른 티어 재검토는 사람이 해야 한다.
- fallback 없음.

## Safety conclusion

- Safe edit boundary: 50% 임계. 낮추려면 시장가 체결의 불확실성을 다시 논증해야 한다.
- High-risk impact: yes — 상한을 되돌리는 변경을 막는 하한이다.

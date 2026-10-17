# Function Logic Map: `StrategyEntrySupervisor.runMarket`

- Source: `internal/app/engine/strategy_entry_supervisor.go` (894-980, **편집 전**, HEAD `f504cc52`, `source_sha256` 9cb510c1…)
- AST evidence: `ast.json` — branches 17 · returns 11
- a137 은 **이 함수를 편집하지 않는 것이 목표**다. 일일 손실 거절이 이 함수의 오류 갈래(B12~B17)에 닿지 않도록 `strategyDispatchCycle.dispatch` 가
  오류가 아닌 타입 결과로 주기를 끝낸다(proposal-freeze P1-3). 이 맵은 그 판단의 근거이고, 「어느 갈래도 잠그지 않음」 시험의 대상 목록이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `worker.effective` | 서명 활성화된 레인이면 참(`StrategyMarketWorker.Effective`, :511) | 감독자 조립 | — |
| `worker.descriptor.RefreshesAuthority` | 생산 조립에서 참(:397 · :442 · :513) | 감독자 조립 | — |
| `refreshOnly` | `= !effective && RefreshesAuthority`(:907) — **서명 활성화된 레인은 거짓** | 위 둘 | — |
| 주기 결과 `err` | `nil` · 일반 오류 · 중앙 무결성 오류 | `invokeBoundedStrategyCycle` → `strategyDispatchCycle.dispatch` | 아래 갈래 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | a137 관련 |
|---|---|---|---|---|
| B1 (895) | 시작 대기 select | — | ctx 취소 → return(897) | 무관 |
| B2 (900) · B3 (901) | 루프 · 큐/취소 select | — | ctx 취소 → return(903) | 무관 |
| B4 (910) | 권한 만료 | `latchMarket(ErrStrategyAuthorityExpired)` | B5 · B6 · B7 | 무관 |
| B5 (912) | 잠금 실패 | `signalCentral` | return(914) | 무관 |
| B6 (916) · B7 (917) | 재시작 대기 실패 | `signalCentral` 또는 return | 918 · 921 | 무관 |
| B8 (925) | 평가 불허 | — | continue | 무관 |
| B9 (929) | 주기 포기 | `markAbandoned` | — | 무관 |
| B10 (932) | 취소 | — | return(933) | 무관 |
| B11 (935) | `err == nil` | — | continue | **a137 목표 갈래** — 계좌 진입 차단은 여기로 와야 함 |
| B12 (938) | `refreshOnly` | `recordSwallowedCycleError` | B13 이 아니면 continue | 오늘 서명 없는 레인의 일일 손실 거절이 여기서 삼켜짐(F7) |
| B13 (956) | B12 안: 중앙 무결성 · 진입 차단 불가 | `signalCentral` | return(958) | 분류 불변 시험(P2-1) |
| B14 (962) | 중앙 무결성(effective) | `signalCentral` | return(964) | 분류 불변 시험(P2-1) |
| (966) | 그 밖의 오류(effective) | **`latchMarket` — durable 레인 잠금** | — | 서명 활성화된 레인의 일일 손실 거절이 여기서 잠금(P1-3) |
| B15 (967) | 잠금 실패 | `signalCentral` | return(969) | — |
| B16 (971) · B17 (972) | 재시작 대기 실패 | `signalCentral` 또는 return | 973 · 976 | — |

Exact AST return positions: 897 · 903 · 914 · 918 · 921 · 933 · 958 · 964 · 969 · 973 · 976

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `invokeBoundedStrategyCycle` | 주기 1회(시간 상한) | `(err, abnormal, cancelled, abandoned)` | AST calls |
| `recordSwallowedCycleError` | refreshOnly 오류 계수 | 오류 없음 | :949 |
| `latchMarket` | 레인 잠금(원장 durable, a112 5.3.3) | 실패 → `signalCentral` | :966 · :1099 |
| `isCentralStrategyIntegrity` · `blockEntryOnCentralIntegrity` · `signalCentral` | 중앙 무결성 분류 | — | :956 · :962 |

## Safety conclusion

- a137 의 요구: 일일 손실 한도 거절과 그 뒤 운영 모드 투영 거절(ENTRY_BLOCKED → `ObserveStrategyEntryGate`)은 `dispatch` 가 **nil 오류 + 타입 결과**로
  돌려 B11 로 온다. B12 · (966) 어느 쪽에도 닿지 않음을 refreshOnly · effective 두 조립 각각의 시험으로 고정한다.
- High-risk impact: 이 함수 무편집. 손절 · 청산 루프는 이 감독자 밖(엔진 런타임)이다.

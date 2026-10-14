# Function Logic Map: `strategyDispatchCycle.dispatch`

- Source: `internal/app/engine/strategy_dispatch_cycle.go` (75-222, **편집 전**, HEAD `f504cc52`, `source_sha256` 72021ef3…)
- AST evidence: `ast.json` — branches 22 · returns 21
- 생산 호출자: `dispatchStrategyMarketHandoffs`(strategy_market_handoff_delivery.go:44)의 전달 몸통 하나 → `deliverEachStrategyHandoff` → 감독자 주기 → `runMarket`

**역할.** 첫 다리 결과 하나를 읽기 전용 경계(스케줄 · FX · 보호 · 활성화 수명 · 진입 관문 · 소유자) → 첫 다리 승인(q_final 원자 예약) → 임대 발급 · 점유 →
Gateway 제출까지 몬다. 오류는 전부 `(Outcome{}, err)` 로 올라가 `runMarket` 의 오류 갈래(B12 삼킴 · (966) 레인 잠금)로 간다.

## Branches and early returns (a137 관련만 상세)

| Branch | Condition | Return/error | a137 |
|---|---|---|---|
| B1~B9 (78~140) | 결과 검증 · 배선 · 스케줄/FX · 의도 · 보호 · 활성화 하한 · 시계 · 임대 상한 | 오류 | 무변경 |
| B10 (144) | `ObserveStrategyEntryGate` 오류(진입 관문 — 운영 모드 투영 · 기동 복원 실패 latch · RECONCILE latch 등. 모드 투영과 기동 복원 실패는 같은 `ReasonOperatingModeBlocked`) | 오류 → 레인 잠금 가능(effective) | **편집(좁게)**: 차단 모드 행의 trigger 가 `ModeTriggerDailyLossLimit` 임이 **타입으로** 실린 거절만 「계좌 진입 차단」 결과. 그 밖은 바이트 단위로 무변경 |
| B11 (148) | 소유자 조회 오류 | 오류 | 무변경 |
| B12 (152) | `admitted.Code != ADMITTED` | B13 | **편집**: `DAILY_LOSS_LIMIT` 이면 「계좌 진입 차단」 타입 결과 |
| B13 (155) | `admitted.cause != nil` | `%w` 로 원인 보존(156) / 아니면 `Detail` 문자열(158) | admit 이 모든 precheck 거절에 cause 를 실으므로 158 은 원인 없는 거절만 남음 |
| B14~B22 (161~216) | 결정 세대 · 위험 세대 · 활성화 수명 · 임대 발급/점유 · 의도 · 최종 재검증 | 오류 | 무변경 |

Exact AST return positions: 79 · 82 · 88 · 94 · 98 · 116 · 137 · 141 · 145 · 149 · 156 · 158 · 162 · 175 · 181 · 194 · 200 · 204 · 206 · 217 · 220

## 「계좌 진입 차단」 결과의 운반

- `dispatch` 는 그 결과를 오류가 아닌 타입으로 돌려주고, 전달 루프는 그 주기의 나머지 범위 전달을 멈춘 뒤 nil 을 돌려준다 → `runMarket` B11.
- 운반 형태(타입 결과 vs 표지 오류를 전달 루프가 nil 로 바꿈)는 편집 시 정하고 이 맵을 갱신한다. 어느 쪽이든 `errors.As` 로만 판정한다(문구 금지).
- 마지막 반환(220, `PlaceClaimedStrategy`)의 `ReasonStrategyDispatchFenced` 중 원인이 trigger 일일 손실 모드 변경인 것도 같은 결과(재확인 (c)) — Gateway 가 펜스 원인을 타입으로 실어야 함(gateway.go:712-716 은 지금 문구만).
- 같은 주기에 앞서 건너뛴 `*strategyScopeRefusal` 이 있으면 전달 루프는 지금처럼 `errors.Join(skipped...)` — 혼합 경우는 오류로 끝남(명시 · 시험).

## Safety conclusion

- 진입을 더 일찍 끝내는 방향만. 주문 제출 · 임대는 차단 결과에서 일어나지 않는다(B12 · B10 이 그 앞).
- 매도 · 손절은 이 함수를 지나지 않는다.

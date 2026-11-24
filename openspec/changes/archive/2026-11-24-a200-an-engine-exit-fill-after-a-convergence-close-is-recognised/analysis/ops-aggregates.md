# a200 운영 집계 (Manager, 2026-11-24 — 읽기 전용, 건수·종류·날짜만. 계좌·종목·수량·가격·주문 번호 없음)

출처: 운영 저널을 `mode=ro` 로 열어 집계(엔진 정지 중이라 저널은 정적). 스키마 `user_version` 36.

## 대사 차단 이력 전체 (`reconcile_states`)
| 원인 | 해제 원인 | 행 수 | 기간 |
|---|---|---|---|
| ATTRIBUTION_FAILED | (미해제) | 1 | 2026-11-20 |
| QUANTITY_MISMATCH | ADJUSTMENT_APPLIED | 113 | 2026-08-06 ~ 2026-11-20 |
| QUANTITY_MISMATCH | OPERATOR | 20 | 2026-07-31 ~ 2026-11-16 |

- `ATTRIBUTION_FAILED` 1건의 evidence 토큰: `X59` · `SELL_ON_CLOSED`. 미해제.
- 지금 미해제 행은 그 1건뿐(미해제 `QUANTITY_MISMATCH` 0 — W2·W3 고착이 지금 운영에 걸려 있지는 않다).
- `ORDER_TERMINAL_CONFIRMED` 해제는 0건(a192 반영 2026-11-17 이후 아직 발동 없음).

## 포지션·청산
- 닫힌 포지션 119, 성과(`trade_outcomes`) 행 61.
- `exit_events` action 별(행 수 / 포지션 수): `ADJUSTMENT_CLOSED` 34/34 · `COMPLETED` 72/72 · `PROPOSAL_FILLED` 111/81.
- `ADJUSTMENT_CLOSED` 포지션의 월별 수: 2026-07 1 · 08 17 · 10 5 · 11 11.
- `ADJUSTMENT_CLOSED` 포지션 34 중 `PROPOSAL_FILLED` 도 있는 것 8.
- 그중 **`ADJUSTMENT_CLOSED` 뒤(같은 초 포함, 이벤트 id 기준)에 `PROPOSAL_FILLED` 가 온 것 2** — 2026-11-16 1, 2026-11-20 1. 두 포지션 모두 성과 행 0.

## 읽는 법
- 이 경합(조정이 닫은 뒤 엔진 매도 체결 도착)은 운영에서 2회 관측됐다. 11-20 은 차단(`ATTRIBUTION_FAILED`)을 남겼고, 11-16 은 남기지 않았다(증거 문서의 W6 과 맞는 모양 — 확정하려면 그 시각의 수량 불일치 행을 대조해야 하며 이 집계는 하지 않았다).
- 두 번 모두 그 청산의 성과가 기록되지 않았다(F15).

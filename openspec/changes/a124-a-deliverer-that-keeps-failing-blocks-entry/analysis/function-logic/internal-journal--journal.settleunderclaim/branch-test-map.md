# Branch Test Map: `Journal.settleUnderClaim`

freeze 단계 — 시험 이름·도달은 tasks 1.4 에서 `covermode=set` 으로 잰다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 토큰 없는 정산 거절 | a099 (1.4 에서 이름 확정) | no(회귀 핀) | 미측정 |
| B3 | UPDATE 오류 → 오류, 롤백 | 1.4 — 없으면 a124 RED 2.10 이 실행자 쪽에서 덮는다 | ? | ? |
| B5 | 적용 → `Applied` | a099 정산 시험 | no | 미측정 |
| 종단 | 임차 상실 / 이미 정산 / 행 없음 | a099 | no | 미측정 |
| **a124 RED** | 적용 시 `Attempts` = 커밋된 증가 후 값, 나열 뒤 다른 발송자가 올린 값까지 반영 | tasks 2.7 | 예정 | 예정 |
| B2 | BeginTx 오류 | 1.4 에서 도달만 기록 — 이 change 가 채우지 않는다 | ? | ? |
| B4 | RowsAffected 오류 | 1.4 에서 도달만 기록 — 이 change 가 채우지 않는다 | ? | ? |
| B6 | 적용 뒤 커밋 오류 | 1.4 에서 도달만 기록 — 이 change 가 채우지 않는다 | ? | ? |
| B7 | explainSettleTx 오류 | 1.4 에서 도달만 기록 — 이 change 가 채우지 않는다 | ? | ? |
| B8 | 설명 뒤 커밋 오류 | 1.4 에서 도달만 기록 — 이 change 가 채우지 않는다 | ? | ? |

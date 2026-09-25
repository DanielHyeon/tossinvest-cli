# Branch Test Map: `Notifier.Acknowledge`

a124 는 이 함수를 편집하지 않는다 — 대조 근거 번들. 분기 도달은 tasks 1.4 에서 잰다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 운영자 공백 거절 | obs 승인 시험 (1.4 에서 이름 확정) | no(회귀 핀) | 미측정 |
| B4 | 전부 승인 (ids 비면 나열) | `TestPersistentDeliveryFailureBlocksEntries` | no | 미측정 |
| B6 | 전부 승인 (나열 행 적재) | `TestPersistentDeliveryFailureBlocksEntries` | no | 미측정 |
| B7 | 전부 승인 (행마다 승인) | `TestPersistentDeliveryFailureBlocksEntries` | no | 미측정 |
| B10 | 미전달 0 → 해제 / 남으면 유지 | `TestAcknowledgeWhileStillPendingKeepsTheBlock` · `TestAcknowledgeCannotClearTheGateMidSend` | no | 미측정 |
| **a124 RED** | 실행자의 늦은 실패와의 두 인터리빙 — 승인이 이긴다 | tasks 2.8 · 2.9 | 예정 | 예정 |
| B2 | 무원장 | 1.4 에서 도달만 기록 | ? | ? |
| B3 | 무원장 + 게이트 → 해제 | 1.4 에서 도달만 기록 | ? | ? |
| B5 | 나열 오류 | 1.4 에서 도달만 기록 | ? | ? |
| B8 | 승인 오류(NotFound 아님) | 1.4 에서 도달만 기록 | ? | ? |
| B9 | 미전달 수 오류 | 1.4 에서 도달만 기록 | ? | ? |

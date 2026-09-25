# Branch Test Map: `alertDeliverer.cycle`

proposal 단계 — 시험은 **이름만** 매핑했고 분기 도달은 재지 않았다(tasks 1.4 에서 `go test -covermode=set` 으로 잰다).
a124 가 더하는 행은 「a124 RED」로 표시한다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 원장 나열 실패 → 사이클 오류, 루프 지속 | `a098_the_outbox_gets_emptied_test.go` (파일 단위) | no(회귀 핀) | 미측정(1.4) |
| B2 | 배치의 행마다 `deliverOne` | `a098_one_cycle_takes_a_batch_test.go` | no(회귀 핀) | 미측정(1.4) |
| B3 | 행 사이 취소 → 남은 행 미처리, `nil` | a098 취소 전파 시험 (1.4 에서 이름 확정) | no(회귀 핀) | 미측정(1.4) |
| 종단 | 배치 소진 → `nil` | 위와 같음 | — | 미측정(1.4) |
| **a124 RED** | 죽은 행 batch 개 + 새 critical 행 → 다음 사이클이 새 행을 먼저 시도 | tasks 2.4 | 예정 | 예정 |
| **a124 RED** | 한도에 이른 행이 잔여 자리로 밀려도 사라지지 않는다 | tasks 2.5 | 예정 | 예정 |

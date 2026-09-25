# Branch Test Map: `alertDeliverer.deliverOne`

proposal 단계 — 시험은 **파일 이름**으로 매핑했고 분기 도달은 재지 않았다(tasks 1.4 에서 `covermode=set` 으로
잰 뒤 함수 이름으로 바꾼다). a124 가 더하는 행은 「a124 RED」.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 임차 요청 오류 | a098 (1.4 에서 이름 확정) | no(회귀 핀) | 미측정 |
| B3 | 임차 획득 → 발행 | `a098_the_outbox_gets_emptied_test.go` | no | 미측정 |
| B4·B5 | 남이 쥔 행은 건너뛰고 임차당 한 번만 말한다 | `a098_two_senders_one_row_test.go` | no | 미측정 |
| B6 | 나열 뒤 정착된 행 | a098 | no | 미측정 |
| B7 | 만료 임차 탈취 | `a098_two_senders_one_row_test.go` | no | 미측정 |
| B8 | publisher 없음 → 반납, `PENDING` 유지 | a098 | no | 미측정 |
| B9 | 전송 실패 → `attempts+1`, 반납 | `a098_the_backlog_does_not_delay_protection_test.go` | no | 미측정 |
| B10 | 실패 기록 자체 실패 | 1.4 에서 확인 — 없으면 a124 가 더한다 | ? | ? |
| B11 | 발행 뒤 정착 실패 → 임차 유지 | a098 | no | 미측정 |
| B12 | 남이 정착한 행 | `a098_the_operator_reads_and_acknowledges_test.go` | no | 미측정 |
| **a124 RED** | B9 로 `attempts` 가 한도에 이른다 → `Gate.Block(ReasonAlertUndelivered)` + `EscalateOperatingMode(CRITICAL_ALERT_UNDELIVERED)`, 동기 경로는 한 번도 시도하지 않은 구성 | tasks 2.1 | 예정 | 예정 |
| **a124 RED** | B8 도 같은 한도 규칙을 탄다(publisher 없음 = 실패 시도) — 열린 결정 D2 | tasks 2.1 | 예정 | 예정 |
| **a124 RED** | 승인으로 미전달 0 → 차단 해제, 모드는 남는다 | tasks 2.2 | 예정 | 예정 |
| **a124 RED** | 한도 도달 뒤 재시작 → 기동 복원이 다시 잠그고 모드는 원장에 있다 | tasks 2.3 | 예정 | 예정 |

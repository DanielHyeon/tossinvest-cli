# Branch Test Map: `restoreAlertEntryLatch`

a124 는 이 함수를 편집하지 않는다 — 대조 근거 번들. 분기 도달은 tasks 1.4 에서 잰다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 계수 오류 → 기동 거부 | `a098_restart_does_not_release_the_gate_test.go` (파일 단위) | no(회귀 핀) | 미측정 |
| B2 | 미전달 0 → 열림 | 같음 | no | 미측정 |
| 종단 | 미전달 > 0 → 잠금 복원 | 같음 | no | 미측정 |
| **a124 RED** | 한도 도달 행이 있는 채 재시작 → 잠금 복원 **그리고** 모드가 원장에 남아 있다 | tasks 2.3 | 예정 | 예정 |

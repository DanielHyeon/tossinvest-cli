# Branch Test Map: `exitOrderLine.args` (pre-edit)

덮음: 이름 grep(실행 측정 미실시). a129 시험은 `a129Want` 로 부분 키 대조.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 체결 줄 | TestA129AConfirmedStopAndItsFillsEachLeaveALine (a129_exit_order_log_test.go:216) | n/a(기존) | 미측정(선택 밖) |
| B1 거짓 | 확정·거절 줄(`side SELL`·`order_type LIMIT`) | 같은 시험 :230 대조 | n/a | 미측정 |
| (예정) | 취소 줄 | a198 §4-4·4-5 | 예정 | 예정 |

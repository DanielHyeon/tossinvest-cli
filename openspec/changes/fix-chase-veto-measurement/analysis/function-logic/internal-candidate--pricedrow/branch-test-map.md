# Branch Test Map: `pricedRow`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 자격을 갖춘 순위 행 | `TestAScanReportsTheShadowRecordForEveryCodeThatHasOne`(`seen_late` 밴드 measured=1을 요구) · `wiring_test.go`의 스캔 배선 테스트들 | yes (두 필드를 지우면 밴드가 미측정으로 실패) | yes |

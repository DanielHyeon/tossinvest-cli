# Branch Test Map: `TestProductionRuntimeStartsExactlyOneAlertDeliverer`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 조립 실패 | `TestProductionRuntimeStartsExactlyOneAlertDeliverer`(자기 자신 — 실패 보고 갈래) | 해당 없음(시험의 실패 갈래) | 해당 없음 — 조립 성공 |
| B2 | 보조 실행자 집합 불일치 | `TestProductionRuntimeStartsExactlyOneAlertDeliverer` | yes — 기대 목록만 고친 상태(배선 전) FAIL, `../../implementation/red-cmd.txt` | yes — 배선 뒤 PASS, `../../implementation/green-cmd.txt` |

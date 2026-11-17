# Branch Test Map: `TestProductionRuntimeStartsExactlyOneAlertDeliverer`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 조립 실패 | 자기 자신(성공이 정상 — 이 갈래는 실패 보고) | 해당 없음(시험의 실패 갈래) | 해당 없음 |
| B2 | 보조 실행자 집합 불일치 | 자기 자신 — 기대 목록만 고친 상태(배선 전)에서 FAIL, 배선 뒤 PASS 를 2단계에서 관측 | 2단계 | 2단계 |

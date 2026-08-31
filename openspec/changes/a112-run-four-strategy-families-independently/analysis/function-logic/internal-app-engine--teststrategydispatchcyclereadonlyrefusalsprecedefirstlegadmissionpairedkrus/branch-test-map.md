# Branch Test Map: `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS`

- Source: `internal/app/engine/strategy_dispatch_cycle_test.go` (151-180); file SHA-256 `ed412100d736dfcb474a0b6c126379c383dc0495be9bdceb409808c18d76f844`.

이 함수는 시험 자신이다. 분기는 단언 실패 경로이므로 "그 분기를 도는 시험"은
이 함수 자체이고, 초록은 어느 실패 arm 도 돌지 않았다는 뜻이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 봉투 경유로만 바뀐 경로 | `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS` | n/a — 재는 대상이 바뀌지 않았다 | PASS |
| B2 | 봉투 경유로만 바뀐 경로 | `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS` | n/a — 재는 대상이 바뀌지 않았다 | PASS |
| B3 | 봉투 경유로만 바뀐 경로 | `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS` | n/a — 재는 대상이 바뀌지 않았다 | PASS |
| B4 | 봉투 경유로만 바뀐 경로 | `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS` | n/a — 재는 대상이 바뀌지 않았다 | PASS |
| B5 | 봉투 경유로만 바뀐 경로 | `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS` | n/a — 재는 대상이 바뀌지 않았다 | PASS |
| B6 | 봉투 경유로만 바뀐 경로 | `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS` | n/a — 재는 대상이 바뀌지 않았다 | PASS |
| B7 | 봉투 경유로만 바뀐 경로 | `TestStrategyDispatchCycleReadOnlyRefusalsPrecedeFirstLegAdmissionPairedKRUS` | n/a — 재는 대상이 바뀌지 않았다 | PASS |

## 반증

이 로트가 바꾼 것은 값을 만드는 방법뿐이므로, 반증은 이 파일이 아니라 봉투
타입에서 한다 — `strategy_dispatch_envelope_test.go` 의
`TestAForgedEnvelopeIsRefusedBeforeAnyGatewayCall` 와
`TestTheSameEnvelopeCannotPlaceASecondOrder`, 그리고 뮤테이션 M1~M3.

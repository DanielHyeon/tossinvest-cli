# Branch Test Map: `Notifier.notifyCritical`

a124 는 이 함수를 편집하지 않는다 — 대조 근거 번들. 분기 도달은 tasks 1.4 에서 잰다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1·B2 | 원장 없음 → 경고 + best-effort | `TestAFailedClaimWithNothingWiredStillReports` | no(회귀 핀) | 미측정 |
| B3 | 기록 실패 → 승격 시도 + 오류 반환 | `TestAClaimThatFailsAttemptsTheDurableBlock` · `TestAFailedClaimStillReturnsItsError` | no | 미측정 |
| B4 | 시도했으나 못 보냄 → 승격 | `TestADeadTransportIsStillFoundAfterASuccessfulDelivery` 계열 (1.4 에서 확정) | no | 미측정 |
| 종단 | 보냄 | `TestOneConditionIsOneSend` | no | 미측정 |

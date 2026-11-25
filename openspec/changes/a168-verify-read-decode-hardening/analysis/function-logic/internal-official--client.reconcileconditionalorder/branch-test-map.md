# Branch Test Map: `Client.ReconcileConditionalOrder`

편집 전 갈래(아래)와 편집 뒤 갈래는 GREEN 뒤 갱신함.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if 177 `strings.TrimSpace(id) == ""` | `TestReconcileReadsKeepTheirInputChecks`(a168, 특성) | 해당 없음(기존 갈래) | 편집 전 pass |
| B2 | if 181 `getAcct(…, &raw)` 오류 | `TestReconcileByIDRefusesEveryOtherOutcome` · `TestReconcileByIDRefusesATransportError` | 해당 없음(기존 갈래) | 편집 전 pass |
| B3 | if 182 `reconcileDetailSilent(err)` | `TestReconcileByIDClassifiesOnlyTheMeasuredNotFoundAsSilent` · `TestReconcileByIDRefusesEveryOtherOutcome` | 해당 없음(기존 갈래) | 편집 전 pass |
| (반환) | 187 `DecodeReconcileConditionalDetail(raw)` | `TestReconcileByIDReadsTheReceiptShapedStatement` · `TestReconcileByIDRefusesSchemaHoles` | 해당 없음(기존 갈래) | 편집 전 pass |

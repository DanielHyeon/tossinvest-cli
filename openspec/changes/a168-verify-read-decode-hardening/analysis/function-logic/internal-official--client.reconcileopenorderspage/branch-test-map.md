# Branch Test Map: `Client.ReconcileOpenOrdersPage`

편집 전 갈래(아래)와 편집 뒤 갈래는 GREEN 뒤 갱신함.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if 208 `getAcct(…, &raw)` 오류 | `TestReconcileClientReadsRefuseANullResult` · a168 `TestReconcileReadsRefuseANon200Success`(204) | 해당 없음(기존 갈래) | 편집 전 pass |
| (반환) | 211 `DecodeReconcileOrderPage(raw)` | `TestReconcileClientReadsSendTheDocumentedGetQueries` | 해당 없음(기존 갈래) | 편집 전 pass |

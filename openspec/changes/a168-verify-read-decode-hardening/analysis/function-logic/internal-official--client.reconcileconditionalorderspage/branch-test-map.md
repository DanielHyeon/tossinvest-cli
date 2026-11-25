# Branch Test Map: `Client.ReconcileConditionalOrdersPage`

편집 전 갈래(아래)와 편집 뒤 갈래는 GREEN 뒤 갱신함.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if 195 `strings.TrimSpace(status) == ""` | `TestReconcileReadsKeepTheirInputChecks`(a168, 특성) | 해당 없음(기존 갈래) | 편집 전 pass |
| B2 | if 199 `getAcct(…, &raw)` 오류 | `TestReconcileClientReadsRefuseANullResult`(봉투 null → 해독 오류는 B2 아님, 아래 반환 202) · a168 `TestReconcileReadsRefuseANon200Success`(204 빈 본문 ErrServer) | 해당 없음(기존 갈래) | 편집 전 pass |
| (반환) | 202 `DecodeReconcileConditionalPage(raw)` | `TestReconcileClientReadsSendTheDocumentedGetQueries` · `TestReconcileClientReadsRefuseANullResult` | 해당 없음(기존 갈래) | 편집 전 pass |

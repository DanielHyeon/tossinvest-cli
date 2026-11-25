# Branch Test Map: `Client.ReconcileInstrument`

편집 전 갈래(아래)와 편집 뒤 갈래는 GREEN 뒤 갱신함.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if 221 `c.get(…, &raw)` 오류 | a168 `TestReconcileReadsRefuseANon200Success`(204 빈 본문) | 해당 없음(기존 갈래) | 편집 전 pass |
| B2 | if 224 `len(raw) != 1` | `TestReconcileReadsKeepTheirInputChecks`(a168, 특성 — 두 항목) | 해당 없음(기존 갈래) | 편집 전 pass |
| (반환) | 227 `raw[0].Symbol` | `TestReconcileInstrumentReturnsTheBrokersEchoNotItsInput` | 해당 없음(기존 갈래) | 편집 전 pass |

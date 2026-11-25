# Branch Test Map: `Client.ReconcileConditionalOrdersPage`

편집 뒤(현재 AST) 갈래. RED 원문: `../../red-green/red-official-before.txt` · `red-verifylive-before.txt`(편집 전), GREEN 원문: `green-official-after.txt` · `green-verifylive-after.txt`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if 195 빈 status → 196 `ErrOrderStatusRequired` | `TestReconcileReadsKeepTheirInputChecks` | 아님(기존 갈래 — 특성) | pass |
| B2 | if 202 전송 오류(관측자 감싼 `getAcct`) → 203 오류 그대로 | `TestReconcileReadsRefuseANon200Success` `cond-204-empty` · `TestReconcileReadsJudgeTheLastAttempt` `cond-401-then-201` 은 B4 | 아님(기존 갈래) | pass |
| B3 | if 206 `DecodeReconcileConditionalPage` 오류 → 207 오류(기존 스키마 거절) | `TestReconcileClientReadsRefuseANullResult` · `TestReconcileReadsKeepTheExistingSchemaRefusalFirst` | 아님(기존 갈래 위치만 이동) | pass |
| B4 | if 210 `reconcileAccepted(attempts)` 실패 — a168 새 갈래 → 211 오류(ErrReconcileSchema) | `TestReconcileReadsRefuseAmbiguousBodies` cond 사례 · `TestReconcileReadsRefuseANon200Success` cond-201~299 · `TestReconcileReadsJudgeTheLastAttempt` cond-401-then-201 · `TestReconcileRefusesAmbiguousBrokerResponsesOverHTTP` R8a·R8b·R8c·R12 | yes — red-official-before.txt · red-verifylive-before.txt(줄 씀) | pass |
| (반환) | 213 `page, nil` → 해독 결과 | `TestReconcileReadsKeepTheirAcceptance` cond-fired-row · `TestReconcileClientReadsSendTheDocumentedGetQueries` | — | pass |

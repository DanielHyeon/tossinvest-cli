# Branch Test Map: `Client.ReconcileOpenOrdersPage`

편집 뒤(현재 AST) 갈래. RED 원문: `../../red-green/red-official-before.txt` · `red-verifylive-before.txt`(편집 전), GREEN 원문: `green-official-after.txt` · `green-verifylive-after.txt`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if 222 전송 오류 → 223 오류 그대로 | `TestReconcileReadsRefuseANon200Success` `order-204-empty` | 아님(기존 갈래) | pass |
| B2 | if 226 `DecodeReconcileOrderPage` 오류 → 227 오류 | `TestReconcileClientReadsRefuseANullResult` | 아님(기존 갈래) | pass |
| B3 | if 230 `reconcileAccepted` 실패 — a168 새 갈래 → 231 오류 | `TestReconcileReadsRefuseAmbiguousBodies` order 사례(R4·W6·W7·R10·R10b) · `TestReconcileReadsRefuseANon200Success` order-201~299 · `TestReconcileReadsJudgeTheLastAttempt` order-401-then-201 · `TestReconcileRefusesAmbiguousBrokerResponsesOverHTTP` R8b(그룹 일반 OPEN 은 정상 — 조건 OPEN 거절 사례) | yes — red-official-before.txt | pass |
| (반환) | 233 `page, nil` → 해독 결과 | `TestReconcileReadsKeepTheirAcceptance` order-open-row | — | pass |

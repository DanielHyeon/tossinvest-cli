# Branch Test Map: `Client.ReconcileConditionalOrder`

편집 뒤(현재 AST) 갈래. RED 원문: `../../red-green/red-official-before.txt` · `red-verifylive-before.txt`(편집 전), GREEN 원문: `green-official-after.txt` · `green-verifylive-after.txt`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if 177 빈 id → 178 ErrReconcileSchema | `TestReconcileReadsKeepTheirInputChecks` | 아님(기존 갈래 — 특성) | pass |
| B2 | if 184 전송·상태 오류 → B3 로 | `TestReconcileByIDRefusesEveryOtherOutcome` · `TestReconcileByIDRefusesATransportError` | 아님(기존 갈래) | pass |
| B3 | if 185 silent(404 conditional-order-not-found) → 186 Silent · 아니면 188 오류 | `TestReconcileByIDClassifiesOnlyTheMeasuredNotFoundAsSilent` · `TestReconcileByIDSilentClassificationIsUnchanged` | 아님(기존 갈래 — 불변 고정) | pass |
| B4 | if 191 `DecodeReconcileConditionalDetail` 오류 → 192 오류(a164 엄격 해독) | `TestReconcileByIDRefusesSchemaHoles` | 아님(기존 갈래) | pass |
| B5 | if 195 `reconcileAccepted` 실패 — a168 새 갈래(D1 · D2 · D6) → 196 오류 | `TestReconcileReadsRefuseAmbiguousBodies` D1·D2·D4·D5·D6·D7 · `TestReconcileReadsRefuseANon200Success` detail-201~299 · `TestReconcileReadsJudgeTheLastAttempt` detail-401-then-201 · `TestReconcileRefusesAmbiguousBrokerResponsesOverHTTP` R9·R11 | yes — red-official-before.txt · red-verifylive-before.txt(R11 줄 씀 3) | pass |
| (반환) | 198 `d, nil` → statement | `TestReconcileReadsKeepTheirAcceptance` detail · `TestReconcileByIDReadsTheReceiptShapedStatement` | — | pass |

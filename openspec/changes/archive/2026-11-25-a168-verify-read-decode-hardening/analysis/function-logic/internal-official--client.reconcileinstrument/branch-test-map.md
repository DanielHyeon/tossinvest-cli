# Branch Test Map: `Client.ReconcileInstrument`

편집 뒤(현재 AST) 갈래. RED 원문: `../../red-green/red-official-before.txt` · `red-verifylive-before.txt`(편집 전), GREEN 원문: `green-official-after.txt` · `green-verifylive-after.txt`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if 246 전송 오류 → 247 오류 | `TestReconcileReadsRefuseANon200Success` `inst-204-empty` | 아님(기존 갈래) | pass |
| B2 | if 249 `len(raw) != 1` → 250 ErrReconcileSchema | `TestReconcileReadsKeepTheirInputChecks` | 아님(기존 갈래 — 특성) | pass |
| B3 | if 253 `reconcileAccepted` 실패 — a168 새 갈래 → 254 오류 | `TestReconcileReadsRefuseAmbiguousBodies` I1 · `TestReconcileReadsRefuseANon200Success` inst-201~299 · `TestReconcileReadsJudgeTheLastAttempt` inst-401-then-201 | yes — red-official-before.txt | pass |
| (반환) | 256 `raw[0].Symbol` → 메아리 | `TestReconcileReadsKeepTheirAcceptance` inst · `TestReconcileInstrumentReturnsTheBrokersEchoNotItsInput` | — | pass |

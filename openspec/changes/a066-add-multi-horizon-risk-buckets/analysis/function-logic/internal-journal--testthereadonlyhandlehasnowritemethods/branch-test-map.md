# Branch Test Map: `TestTheReadOnlyHandleHasNoWriteMethods`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | every `*ReadOnly` method walked (for at 155) | this test | n/a (loop) | GREEN 2026-09-29 |
| B2 | unlisted method (if at 157) | this test | RED before the two names were listed (2026-09-28: "unlisted method ReadEntryLossLocks / ReadRiskOwnerLatches") | GREEN |

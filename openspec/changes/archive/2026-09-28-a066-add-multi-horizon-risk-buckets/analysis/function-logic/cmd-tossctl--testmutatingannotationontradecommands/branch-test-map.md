# Branch Test Map: `TestMutatingAnnotationOnTradeCommands`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | every leaf command walked (range at 140) | this test | n/a (loop) | GREEN 2026-09-29 |
| B2 | listed command lacks `mutating: true` (if at 143) | this test | mutation C06 (`entry-lock-release` → false) CAUGHT | GREEN |
| B3 | unlisted command declares `mutating: true` (if at 146) | this test | existing (a066 changes data only) | GREEN |

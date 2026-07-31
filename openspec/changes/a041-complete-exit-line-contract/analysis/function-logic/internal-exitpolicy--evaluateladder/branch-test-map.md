# Branch Test Map: `EvaluateLadder`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | invalid policy refuses evaluation | existing policy validation tests | existing | yes |
| B2 | state policy ID differs from executable table | existing policy mismatch tests | existing | yes |
| B3 | canonical policy identity cannot be constructed | a041 identity validation tests | no | yes |
| B4 | persisted policy version differs from executable table | a041 snapshot state mismatch test | no | yes |
| B5 | persisted policy digest differs from executable table | a041 snapshot state mismatch test | no | yes |
| B6 | invalid entry price refuses evaluation | existing unusable-input tests | existing | yes |
| B7 | invalid observed price refuses evaluation | existing unusable-input tests | existing | yes |
| B8 | invalid high-water refuses evaluation | existing unusable-input tests | existing | yes |
| B9 | invalid previous baseline refuses evaluation | existing unusable-input tests | existing | yes |
| B10 | invalid taken ratio refuses evaluation | existing ratio validation tests | existing | yes |
| B11 | active rung outside the policy table is refused | existing active-rung bounds test | existing | yes |
| B12 | a higher observation advances the watermark | existing monotonicity tests | existing | yes |
| B13 | every rung is scanned for the highest reached target | existing rung reach table tests | existing | yes |
| B14 | malformed target refuses evaluation | existing policy validation tests | existing | yes |
| B15 | only a newly reached higher rung promotes | existing jump promotion tests | existing | yes |
| B16 | any active rung contributes its protection lock | existing protected ladder tests | existing | yes |
| B17 | invalid lock calculation refuses evaluation | existing invalid policy tests | existing | yes |
| B18 | final runner rung adds a high-water trail candidate | existing runner tests | existing | yes |
| B19 | malformed runner trail refuses evaluation | existing runner validation tests | existing | yes |
| B20 | protection composition failure refuses evaluation | existing candidate composition tests | existing | yes |
| B21 | malformed composed baseline refuses evaluation | existing baseline tests | existing | yes |
| B22 | a newly reached rung updates decision-time state | existing promotion tests | existing | yes |
| B23 | completed ladder preserves stored rung and emits no proposal | existing completed test | existing | yes |
| B24 | malformed effective baseline refuses evaluation | existing baseline tests | existing | yes |
| B25 | observed price below newly composed protection takes breach priority | existing stop tests + a041 promoted-breach snapshot test | existing | yes |
| B26 | an already-pending ladder stop is suppressed | existing pending stop test | existing | yes |
| B27 | no new rung means no take-profit outcome | existing quiet ladder tests | existing | yes |
| B28 | promoted rung outcome selects final, partial, or state-only behavior | existing promotion table tests | existing | yes |
| B29 | final take-full rung proposes the complete remainder | existing final rung test + a041 one-share final test | existing | yes |
| B30 | positive intermediate ratio proposes a partial | existing intermediate rung tests | existing | yes |
| B31 | zero-ratio rung performs a state-only protection promotion | existing first-rung test | existing | yes |
| B32 | unresolved proposal suppresses a newly promoted take-profit | existing pending partial test | existing | yes |

# Branch Test Map: `LadderPolicy.Validate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | empty id/rungs refused | existing validation tests | existing | yes |
| B2 | each invalid rung refused | existing rung tests | existing | yes |
| B3 | descending target/stop refused | existing monotonicity tests | existing | yes |
| B4 | runner missing/valid/invalid | common policy runner tests | existing | yes |
| B5 | same id/version with wrong digest refused | `TestPolicyIdentityRejectsDigestCollision` | no | yes |
| B6 | rung targets must increase strictly | existing monotonicity tests | existing | yes |
| B7 | protective stop percentages must not decrease | existing monotonicity tests | existing | yes |
| B8 | configured runner trail is parsed | existing runner validation tests | existing | yes |
| B9 | malformed runner trail is refused | existing runner validation tests | existing | yes |
| B10 | runner trail must be in the open 0..100 interval | existing runner validation tests | existing | yes |

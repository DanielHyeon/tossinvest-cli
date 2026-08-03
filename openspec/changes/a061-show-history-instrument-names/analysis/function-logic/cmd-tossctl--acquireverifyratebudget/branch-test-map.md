# Branch Test Map: `acquireVerifyRateBudget`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | metadata owner outlives verification context | `TestA061VerificationWaitsForAndThenOwnsTheSameRateBudget` | yes | yes |
| B2 | profile path resolves but kernel lease acquisition fails or is canceled | `TestA061VerificationWaitsForAndThenOwnsTheSameRateBudget` | yes | yes |
| tail | metadata releases and verification acquires identical path | same test | yes | yes |

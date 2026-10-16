# Branch Test Map: `M0ExactPrerequisites`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Enumerates every catalogue step | `TestM0ExactPrerequisitesStillRejectsEveryNonTTLNonPass` | `TestM0TriggerAdmitsEveryTTLVerdictWithoutTTLMutation` | all but PASS rejected with `idempotency-ttl-edge` prior-PASS error | all admit trigger execution; zero TTL planned mutation and order POST/DELETE |
| B2 | TTL PASS, skipped, failed, refused, missing, awaiting-restart | `TestM0CLIValidationTreatsOnlyTTLAsNonPrerequisite` | skipped/fail/refused/awaiting-restart rejected | all accepted by CLI M0 validation |
| B3 | `order-cancel` missing, fail, skipped, refused, awaiting-restart | `TestM0ExactPrerequisitesStillRejectsEveryNonTTLNonPass` | existing hold behavior | remains rejected and names `order-cancel` |
| caller hold | outstanding/non-pending checkpoint/pending-create recovery | existing `TestM0RestartWithDurableParentOrChildOwnerNeverReentersTriggerMutation`, `TestM0PendingRecoveryZeroOneMultipleHoldWithoutMutation` | not changed | focused regression retained |

# Branch Test Map: `validateLegacyAuditEvent`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | missing/corrupt snapshot refuses legacy event | `TestMigrationRejectsCorruptLegacyAuditRow` | legacy audit unauthenticated | PASS |
| B2 | snapshot audit ID/actor/reason/time mismatch refuses event | same corrupt legacy matrix | plausible tamper could be signed | PASS |
| B3 | missing candidate refuses event | same corrupt legacy matrix | plausible tamper could be signed | PASS |
| B4 | malformed candidate changes refuses event | same corrupt legacy matrix | plausible tamper could be signed | PASS |
| B5 | every candidate change is compared | v2 migration test | audit digest absent | PASS |
| B6 | exact candidate change succeeds; no match fails | v2 migration and corrupt legacy tests | audit digest absent | PASS |

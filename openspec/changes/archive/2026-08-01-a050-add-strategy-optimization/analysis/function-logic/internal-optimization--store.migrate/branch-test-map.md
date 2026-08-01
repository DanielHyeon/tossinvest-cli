# Branch Test Map: `Store.migrate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | begin failure | DB fault path | n/a | n/a |
| B2 | version read failure | DB fault path | n/a | n/a |
| B3 | future schema refusal | `TestOpenRefusesNewerSchemaAndSecuresFiles` | yes | yes |
| B4 | DDL failure | migration fault path | n/a | n/a |
| B5 | version write failure | migration fault path | n/a | n/a |
| B6 | commit failure | migration fault path | n/a | n/a |
| B7 | inspect legacy payload MAC column | `TestMigrationAddsPayloadMACToEmptyLegacyCandidateSchema` | yes | yes |
| B8 | add missing legacy payload MAC column | `TestMigrationAddsPayloadMACToEmptyLegacyCandidateSchema` | yes | yes |
| B9 | alter migration failure | migration fault path | n/a | n/a |
| B10 | legacy actor column migration DDL failure aborts | migration rollback tests | actor column absent before hardening | PASS |
| B11 | pre-install trigger verification refuses drift while allowing genuinely missing legacy triggers | trigger drift/migration tests | exact verification absent | PASS |
| B12 | pre-v2 database enters digest migration | legacy digest migration test | full metadata digest absent | PASS |
| B13 | old snapshot update trigger drop failure aborts | migration rollback review | v2 migration absent | defensive branch reviewed |
| B14 | snapshot digest migration error aborts whole transaction | legacy digest tamper test | corrupt legacy row could be re-signed | PASS |
| B15 | append-only trigger installation error aborts | schema migration tests | exact triggers absent | PASS |
| B16 | post-install exact trigger verification error aborts | same-name no-op/drift test | name-only trigger could bypass protection | PASS |
| B17 | schema version write failure aborts | transaction fault-path review | versioned schema absent | defensive branch reviewed |
| B18 | commit failure leaves prior schema intact | migration rollback test | atomic versioned migration absent | PASS |
| B19 | pre-v2 database enters snapshot digest migration | v1 migration tests | full snapshot digest absent | PASS |
| B20 | old snapshot update trigger drop failure aborts | migration fault review | v2 migration absent | defensive branch reviewed |
| B21 | snapshot digest migration error aborts transaction | corrupt v1 digest test | corrupt row could be re-signed | PASS |
| B22 | pre-v3 database enters pointer/audit migration | `TestMigrationFromV2BindsControlPointerAndLegacyAuditRows` | v3 integrity absent | PASS |
| B23 | control pointer digest migration error aborts | v2 migration/corruption tests | pointer unauthenticated | PASS |
| B24 | old audit update trigger drop failure aborts | migration fault review | v3 migration absent | defensive branch reviewed |
| B25 | audit digest migration/corroboration error aborts | `TestMigrationRejectsCorruptLegacyAuditRow` | plausible tamper could be signed | PASS |
| B26 | append-only trigger installation failure aborts | schema migration tests | v3 triggers absent | PASS |
| B27 | post-install exact trigger verification failure aborts | same-name trigger drift test | drift could survive migration | PASS |
| B28 | schema v3 version write failure aborts | transaction fault review | v3 migration absent | defensive branch reviewed |
| B29 | final commit failure rolls back all v3 schema/digest changes | migration rollback test | v3 migration absent | PASS |

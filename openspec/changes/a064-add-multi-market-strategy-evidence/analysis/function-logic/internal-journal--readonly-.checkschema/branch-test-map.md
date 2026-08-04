# Branch Test Map: `ReadOnly.checkSchema`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | user-version query failure is returned | existing read-only open failure coverage | existing coverage | PASS |
| B2 | schema newer than this build is refused | existing schema-direction test | existing coverage | PASS |
| B3 | every released base table is inspected | existing read-only schema tests | existing coverage | PASS |
| B4 | base-table query result is classified | existing read-only schema tests | existing coverage | PASS |
| B5 | missing base table is accumulated | existing missing-table test | existing coverage | PASS |
| B6 | base-table inspection error is returned | existing read-only failure coverage | existing coverage | PASS |
| B7 | accumulated base-table absence is `ErrSchemaTooOld` | existing missing-table test | existing coverage | PASS |
| B8 | every released base column is inspected | existing read-only schema tests | existing coverage | PASS |
| B9 | base-column query result is classified | existing read-only schema tests | existing coverage | PASS |
| B10 | missing base column is accumulated | existing damaged-schema tests | existing coverage | PASS |
| B11 | base-column inspection error is returned | existing read-only failure coverage | existing coverage | PASS |
| B12 | accumulated base-column absence is `ErrSchemaTooOld` | existing damaged-schema tests | existing coverage | PASS |
| B13 | v20 campaign checks are version-gated | existing v19/v20 migration and read-only tests | existing coverage | PASS |
| B14 | every v20 campaign table is inspected | existing damaged-v20 test | existing coverage | PASS |
| B15 | v20 table query result is classified | existing damaged-v20 test | existing coverage | PASS |
| B16 | missing v20 table is accumulated | existing damaged-v20 test | existing coverage | PASS |
| B17 | v20 table inspection error is returned | existing read-only failure coverage | existing coverage | PASS |
| B18 | every v20 campaign column is inspected | existing damaged-v20 test | existing coverage | PASS |
| B19 | v20 column query result is classified | existing damaged-v20 test | existing coverage | PASS |
| B20 | missing v20 column is accumulated | existing damaged-v20 test | existing coverage | PASS |
| B21 | v20 column inspection error is returned | existing read-only failure coverage | existing coverage | PASS |
| B22 | damaged v20 prerequisite is `ErrSchemaTooOld` | existing damaged-v20 test | existing coverage | PASS |
| B23 | v21 evidence-lineage checks are version-gated | v20 migration preservation plus damaged-v21 tests | compile-fail contract captured | PASS |
| B24 | both snapshot reference columns are inspected | `TestOpenReadOnlyRejectsDamagedV21EvidenceLineage` | compile-fail contract captured | PASS |
| B25 | v21 column query result is classified | `TestOpenReadOnlyRejectsDamagedV21EvidenceLineage` | compile-fail contract captured | PASS |
| B26 | missing v21 column is accumulated | `TestOpenReadOnlyRejectsDamagedV21EvidenceLineage` | compile-fail contract captured | PASS |
| B27 | v21 column inspection error is returned | existing read-only failure mechanics | existing coverage | PASS |
| B28 | damaged v21 lineage is `ErrSchemaTooOld` | `TestOpenReadOnlyRejectsDamagedV21EvidenceLineage` | compile-fail contract captured | PASS |

The successful v21 path is exercised by `TestStrategyEvidenceLineagePersistsOnlyImmutableReference`; PASS.

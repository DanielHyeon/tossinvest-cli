# Branch Test Map: `ReadOnly.checkSchema`

Rows corrected 2026-09-07. The previous version cited "existing read-only failure coverage" for all six
`case err != nil` arms and "existing v19/v20 tests" for both version gates. Neither existed: every
`OpenReadOnly(` call site in the package supplied a well-formed journal at v21 or later, so no arm and
neither gate had ever executed. The tests named below are the ones that reach them; each RED column
records what was observed when the branch was mutated under `go test -overlay`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | the user-version read itself fails and is reported as such, not as a schema verdict | `TestCheckSchemaReportsEveryInspectionFailureInsteadOfSwallowingIt` | the arm swallowed: `checkSchema` reported the journal healthy | PASS |
| B2 | a journal newer than this build is refused in the other direction | existing schema-direction test in `readonly_test.go` | existing coverage | PASS |
| B3 | every released base table is inspected | `TestOpenReadOnlyRejectsV9BeforeSnapshotQueries` | existing coverage | PASS |
| B4 | a base-table query result is classified as present or absent | `TestOpenReadOnlyRejectsV9BeforeSnapshotQueries` | existing coverage | PASS |
| B5 | a missing base table is accumulated | `TestOpenReadOnlyRejectsV9BeforeSnapshotQueries` | existing coverage | PASS |
| B6 | a base-table inspection failure is returned instead of counted as absence | `TestCheckSchemaReportsEveryInspectionFailureInsteadOfSwallowingIt` | the arm swallowed: the failure became `ErrSchemaTooOld` | PASS |
| B7 | accumulated base-table absence is `ErrSchemaTooOld` | `TestOpenReadOnlyRejectsV9BeforeSnapshotQueries` | existing coverage | PASS |
| B8 | every released base column is inspected | `TestOpenReadOnlyRejectsV9BeforeSnapshotQueries` | existing coverage | PASS |
| B9 | a base-column query result is classified | `TestOpenReadOnlyRejectsV9BeforeSnapshotQueries` | existing coverage | PASS |
| B10 | a missing base column is accumulated and named | `TestOpenReadOnlyRejectsV9BeforeSnapshotQueries` | existing coverage | PASS |
| B11 | a base-column inspection failure is returned | `TestCheckSchemaReportsEveryInspectionFailureInsteadOfSwallowingIt` | the arm swallowed | PASS |
| B12 | accumulated base-column absence is `ErrSchemaTooOld` | `TestOpenReadOnlyRejectsV9BeforeSnapshotQueries` | existing coverage | PASS |
| B13 | a complete v19 journal skips the v20 campaign checks | `TestOpenReadOnlyAcceptsCompleteJournalsBelowTheVersionGates` | no test opened a complete pre-v20 journal read-only; forcing the gate true went unnoticed | PASS |
| B14 | every v20 campaign table is inspected | existing damaged-v20 test in `readonly_test.go` | existing coverage | PASS |
| B15 | a v20 table query result is classified | existing damaged-v20 test in `readonly_test.go` | existing coverage | PASS |
| B16 | a missing v20 table is accumulated | existing damaged-v20 test in `readonly_test.go` | existing coverage | PASS |
| B17 | a v20 table inspection failure is returned | `TestCheckSchemaReportsEveryInspectionFailureInsteadOfSwallowingIt` | the arm swallowed | PASS |
| B18 | every v20 campaign column is inspected | existing damaged-v20 test in `readonly_test.go` | existing coverage | PASS |
| B19 | a v20 column query result is classified | existing damaged-v20 test in `readonly_test.go` | existing coverage | PASS |
| B20 | a missing v20 column is accumulated | existing damaged-v20 test in `readonly_test.go` | existing coverage | PASS |
| B21 | a v20 column inspection failure is returned | `TestCheckSchemaReportsEveryInspectionFailureInsteadOfSwallowingIt` | the arm swallowed | PASS |
| B22 | a damaged v20 prerequisite is `ErrSchemaTooOld` | existing damaged-v20 test in `readonly_test.go` | existing coverage | PASS |
| B23 | a complete v20 journal skips the v21 evidence-lineage checks | `TestOpenReadOnlyAcceptsCompleteJournalsBelowTheVersionGates` | no test opened a complete v20 journal read-only; forcing the gate true went unnoticed | PASS |
| B24 | **each** snapshot-reference column is inspected — a journal carrying only one is refused by the name of the other | `TestOpenReadOnlyNamesTheMissingV21ColumnIndividually` | the refusal matched on the shared prefix, so deleting one column from the list changed nothing | PASS |
| B25 | a v21 column query result is classified | `TestOpenReadOnlyRejectsDamagedV21EvidenceLineage` | existing coverage | PASS |
| B26 | a missing v21 column is accumulated | `TestOpenReadOnlyRejectsDamagedV21EvidenceLineage` | existing coverage | PASS |
| B27 | a v21 column inspection failure is returned — a064's own arm | `TestCheckSchemaReportsEveryInspectionFailureInsteadOfSwallowingIt` | the arm swallowed: an injected driver failure was reported as a healthy schema | PASS |
| B28 | a damaged v21 lineage is `ErrSchemaTooOld` | `TestOpenReadOnlyRejectsDamagedV21EvidenceLineage` | existing coverage | PASS |

The successful v21 path is exercised by `TestStrategyEvidenceLineagePersistsOnlyImmutableReference`; PASS.

# Branch Test Map: `ReadOnly.checkSchema`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | pragma query failure | existing `readonly_test.go` cases | baseline | yes |
| B2 | schema is newer than reader | existing `readonly_test.go` cases | baseline | yes |
| B3 | required-table iteration | existing `readonly_test.go` cases | baseline | yes |
| B4 | v8 tables but no policy_id | `TestOpenReadOnlyRejectsV8BeforePolicyQuery` | yes | yes |
| B5 | metadata failure | malformed schema test or inherited DB error coverage | baseline | yes |
| B6 | current writer remains readable | `TestOpenReadOnlyReadsWhatTheEngineIsWriting` | baseline | yes |
| B7 | missing required table is rejected | existing `readonly_test.go` cases | baseline | yes |
| B8 | required-column iteration | `TestOpenReadOnlyRejectsV8BeforePolicyQuery` | yes | yes |
| B9 | required-column inspection result | `TestOpenReadOnlyRejectsV8BeforePolicyQuery` | yes | yes |
| B10 | missing required column is classified | `TestOpenReadOnlyRejectsV8BeforePolicyQuery` | yes | yes |
| B11 | required-column metadata failure | inherited read-only schema error contract | baseline | yes |
| B12 | missing required columns produce typed refusal | `TestOpenReadOnlyRejectsV8BeforePolicyQuery` | yes | yes |

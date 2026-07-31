# Branch Test Map: `Console.openJournal`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | empty path remains unwired | existing portfolio tests | baseline | yes |
| B2 | read-only open result is classified | existing portfolio state tests | baseline | yes |
| B3 | selected current journal opens and renders, preserving a whitespace-relative path exactly | `TestPositionsReadsOnlyTheSelectedJournal`, `TestPositionsPreservesTheExactSelectedJournalPath` | yes (whitespace-selected markers absent after path trim) | yes |
| B4 | selected file missing; no fallback | command resolver plus existing portfolio missing test | baseline | yes |
| B5 | selected journal newer than reader | existing portfolio schema test | baseline | yes |
| B6 | selected v8 journal is too old before query | `TestOpenReadOnlyRejectsV8BeforePolicyQuery` | yes | yes |
| B7 | malformed selected journal reports failure | existing portfolio invalid-journal test | baseline | yes |

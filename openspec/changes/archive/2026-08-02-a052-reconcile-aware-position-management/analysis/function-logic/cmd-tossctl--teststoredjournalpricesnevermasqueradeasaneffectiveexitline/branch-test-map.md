# Branch Test Map: `TestStoredJournalPricesNeverMasqueradeAsAnEffectiveExitLine`

| Branch | AST control path | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if` line 92 | `if projected.ExitLine.Status != "unknown" \|\| projected.ExitLine.CurrentProtection != "—" \|\|` entered and complementary path | TestStoredJournalPricesNeverMasqueradeAsAnEffectiveExitLine | a052 RED contract or pre-existing regression | verified by focused package suite |
| B2 | `if` line 96 | `if projected.StoredExitEvidence == nil \|\| projected.StoredExitEvidence.EntryPrice != "70000" \|\|` entered and complementary path | TestStoredJournalPricesNeverMasqueradeAsAnEffectiveExitLine | a052 RED contract or pre-existing regression | verified by focused package suite |

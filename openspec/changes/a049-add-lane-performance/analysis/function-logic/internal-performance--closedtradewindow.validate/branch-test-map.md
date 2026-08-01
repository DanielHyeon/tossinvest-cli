# Branch Test Map: `ClosedTradeWindow.validate`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at line 31: `if strings.TrimSpace(w.AccountRef) == "" {`; invariant: missing/corrupt/alternate path is explicit | `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations`, `TestJournalHandoffValidatesAccountAndWindowBeforeReading`, `TestJournalHandoffReaderErrorWritesNothing`, `TestJournalHandoffStoreBindsOneServerSelectedAccount` | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |
| B2 | `if` at line 34: `if w.ClosedAfter.IsZero() \|\| w.ClosedAtOrBefore.IsZero() \|\| !w.ClosedAfter.Before(w.ClosedAtOrBefore) {`; invariant: missing/corrupt/alternate path is explicit | `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations`, `TestJournalHandoffValidatesAccountAndWindowBeforeReading`, `TestJournalHandoffReaderErrorWritesNothing`, `TestJournalHandoffStoreBindsOneServerSelectedAccount` | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |

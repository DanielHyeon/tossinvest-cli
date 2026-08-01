# Branch Test Map: `TestJournalHandoffReaderErrorWritesNothing`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at line 81: `if err == nil \|\| reader.calls != 1 {`; invariant: missing/corrupt/alternate path is explicit | `TestJournalHandoffReaderErrorWritesNothing` (this regression test) | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |
| B2 | `if` at line 85: `if err := store.db.QueryRow(\`SELECT count(*) FROM performance_trades\`).Scan(&trades); err != nil \|\| trades != 0 {`; invariant: missing/corrupt/alternate path is explicit | `TestJournalHandoffReaderErrorWritesNothing` (this regression test) | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |

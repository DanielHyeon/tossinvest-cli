# Branch Test Map: `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at line 37: `if err != nil {`; invariant: missing/corrupt/alternate path is explicit | `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations` (this regression test) | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |
| B2 | `if` at line 40: `if reader.calls != 1 \|\| len(got) != 1 \|\| got[0].Markout(5).ObservationID != "existing-5" {`; invariant: missing/corrupt/alternate path is explicit | `TestJournalHandoffConsumesOneExactLineageReadAndCallerOwnedObservations` (this regression test) | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |

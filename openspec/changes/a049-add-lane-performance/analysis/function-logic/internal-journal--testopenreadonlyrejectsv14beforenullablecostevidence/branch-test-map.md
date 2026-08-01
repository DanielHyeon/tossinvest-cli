# Branch Test Map: `TestOpenReadOnlyRejectsV14BeforeNullableCostEvidence`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at line 202: `if err := v14.Close(); err != nil {`; invariant: missing/corrupt/alternate path is explicit | `TestOpenReadOnlyRejectsV14BeforeNullableCostEvidence` (this regression test) | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |
| B2 | `if` at line 206: `if !errors.Is(err, ErrSchemaTooOld) \|\| !strings.Contains(err.Error(), "trade_outcomes.cost_total") {`; invariant: missing/corrupt/alternate path is explicit | `TestOpenReadOnlyRejectsV14BeforeNullableCostEvidence` (this regression test) | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |

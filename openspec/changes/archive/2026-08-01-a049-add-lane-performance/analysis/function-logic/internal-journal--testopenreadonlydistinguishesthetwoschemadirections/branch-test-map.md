# Branch Test Map: `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | base-revision `if` at line 155: `if _, err := j.db.ExecContext(ctx, "PRAGMA user_version = 9999"); err != nil {`; hunk adjacency requires base evidence | `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections` (this regression test) | existing regression branch | package-targeted regression PASS before integration; rerun by gate |
| B2 | base-revision `if` at line 160: `if !errors.Is(err, ErrSchemaTooNew) {`; hunk adjacency requires base evidence | `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections` (this regression test) | existing regression branch | package-targeted regression PASS before integration; rerun by gate |
| B3 | base-revision `if` at line 177: `if err != nil {`; hunk adjacency requires base evidence | `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections` (this regression test) | existing regression branch | package-targeted regression PASS before integration; rerun by gate |
| B4 | base-revision `if` at line 180: `if err := j.Close(); err != nil {`; hunk adjacency requires base evidence | `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections` (this regression test) | existing regression branch | package-targeted regression PASS before integration; rerun by gate |
| B5 | base-revision `if` at line 185: `if !errors.Is(err, ErrSchemaTooOld) {`; hunk adjacency requires base evidence | `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections` (this regression test) | existing regression branch | package-targeted regression PASS before integration; rerun by gate |
| B6 | base-revision `if` at line 190: `if !strings.Contains(err.Error(), "positions") {`; hunk adjacency requires base evidence | `TestOpenReadOnlyDistinguishesTheTwoSchemaDirections` (this regression test) | existing regression branch | package-targeted regression PASS before integration; rerun by gate |

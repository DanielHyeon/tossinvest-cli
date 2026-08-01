# Branch Test Map: `openCrashChildJournalAtVersion`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `if` at line 57: `if path == "" {`; invariant: missing/corrupt/alternate path is explicit | `TestMigrationV14CommitAndUserVersionSurviveSIGKILL` | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |
| B2 | `if` at line 64: `if err != nil {`; invariant: missing/corrupt/alternate path is explicit | `TestMigrationV14CommitAndUserVersionSurviveSIGKILL` | not separately captured for this evidence refresh | package-targeted regression PASS before integration; rerun by gate |

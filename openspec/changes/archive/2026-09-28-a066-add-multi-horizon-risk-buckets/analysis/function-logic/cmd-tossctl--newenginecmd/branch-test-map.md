# Branch Test Map: `newEngineCmd`

The AST has no branches (`branches: null`, 114–127). The one path is the happy path below. It includes the a066
registration statement.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path: `engine` command built with `run`, `reconcile-resolve`, `alerts` and the three a066 5.5 commands registered | `TestEngineRiskRelaxationCommandsDeclareWhatTheyAre` (walks `newRootCmd()`; fails if any of the three is missing or mis-annotated) | mutation C06 (`entry-lock-release` declared non-mutating) CAUGHT | GREEN 2026-09-29 |

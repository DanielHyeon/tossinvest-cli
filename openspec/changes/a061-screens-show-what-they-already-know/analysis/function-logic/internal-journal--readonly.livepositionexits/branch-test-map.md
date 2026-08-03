# Branch Test Map: `ReadOnly.LivePositionExits`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | an exit-state read failure propagates | existing journal tests | n/a | pass |
| B2 | a quarantine read failure propagates rather than being swallowed | fail-closed by construction: the error is returned unchanged, and the caller's existing `journalFailed` path is covered by the existing schema/permission tests | n/a | pass |
| B3 | a positions query failure propagates | existing journal tests | n/a | pass |
| B4 | one row per live position | existing journal tests | n/a | pass |
| B5 | a scan failure propagates | existing journal tests | n/a | pass |
| B6 | an active quarantine on the current generation reaches the row; a quarantine on a previous generation does not | `TestAQuarantinedPositionIsNotDrawnAsProtected`; `TestAReleasedGenerationsQuarantineDoesNotCloseTheCurrentOne` | yes | yes |
| B7 | an exit state attaches and sets `HasExit` | existing journal tests | n/a | pass |
| B8 | a pre-v10 state resolves its legacy identity in memory | existing legacy tests | n/a | pass |
| B9 | a resolvable legacy identity is set | existing legacy tests | n/a | pass |
| B10 | an unresolvable one becomes a typed unknown reason | existing legacy tests | n/a | pass |
| B11 | a row iteration error propagates | existing journal tests | n/a | pass |

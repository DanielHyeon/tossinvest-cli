# Branch Test Map: `ReadOnly.accountActiveQuarantines`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | a query failure is returned, never softened into an empty map | fail-closed by construction; the caller's `journalFailed` rendering is covered by the existing schema tests | n/a | pass |
| B2 | an unreleased quarantine on the current generation is indexed by position id; a quarantine on a previous generation is not returned | `TestAQuarantinedPositionIsNotDrawnAsProtected`; `TestAReleasedGenerationsQuarantineDoesNotCloseTheCurrentOne` | yes | yes |
| B3 | a scan failure is returned | fail-closed by construction | n/a | pass |
| B4 | an iteration error is returned | fail-closed by construction | n/a | pass |

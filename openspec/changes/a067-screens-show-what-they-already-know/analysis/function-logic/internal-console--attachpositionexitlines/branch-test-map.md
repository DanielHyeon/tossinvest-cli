# Branch Test Map: `attachPositionExitLines`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | every row on both holdings screens is decorated once | `TestTheProtectionLineSurvivesAJudgementThatChangedNothing` | yes | yes |
| B2 | a designated-but-unmanaged holding keeps an explicitly unknown reference | `TestPendingDesignationTruthTable` (a053) | n/a | pass |
| B3 | a holding with no exit state renders reference only | `TestPositionsRenderCanonicalExitLineFixtures` | n/a | pass |
| B4 | lifecycle generation mismatch hides every stored price | `TestPositionsSuppressCrossLifecycleExitEvidence` | n/a | pass |
| B5 | an operator-released lifecycle hides them | `TestReleasedDesignatedRowDoesNotShowPendingFallback` | n/a | pass |
| B6 | released rows still show legacy raw evidence in the detail area | same | n/a | pass |
| B7 | an active quarantine closes the line while the engine runs; a previous generation's quarantine does not close the current one | `TestAQuarantinedPositionIsNotDrawnAsProtected`; `TestAReleasedGenerationsQuarantineDoesNotCloseTheCurrentOne` | yes | yes |
| B8 | a quarantined position with no canonical snapshot still says so | `TestAQuarantinedPositionIsNotDrawnAsProtected` (SEED path) | yes | yes |
| B9 | a canonical snapshot supplies the five values and its provenance | `TestTheProtectionLineSurvivesAJudgementThatChangedNothing` | yes | yes |
| B10 | an unverifiable lifecycle suppresses the line | `TestPositionsSuppressCorruptAndLifecycleUnverifiedRawEvidence` | n/a | pass |
| B11 | legacy raw evidence reaches the detail area only | `TestPositionsRenderCanonicalExitLineFixtures` | n/a | pass |

RED observed = the test failed against the pre-a067 code, verified 2026-08-03.
`n/a` marks pre-existing regression guards that were green before and after.

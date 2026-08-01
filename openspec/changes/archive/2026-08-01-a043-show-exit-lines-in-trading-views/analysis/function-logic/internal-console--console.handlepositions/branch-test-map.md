# Branch Test Map: `Console.handlePositions`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | no settings seam still renders read-only position exit lines | `TestPositionsControlsWorkUnderTheDeployedCSPWithoutScript` | yes | yes |
| B2 | settings load succeeds and annotates state without a form/control | settings label tests + `TestPositionsControlsWorkUnderTheDeployedCSPWithoutScript` | yes | yes |
| B3 | each row receives both include/exclude labels from one settings snapshot | settings precedence tests | existing | yes |

The new leaf `attachPositionExitLines` is covered by `TestPositionsRenderCanonicalExitLineFixtures` for complete, stale, unknown, and one-share states; it was not an existing-function edit and therefore has its unit/integration tests rather than a separate Function Logic Map.

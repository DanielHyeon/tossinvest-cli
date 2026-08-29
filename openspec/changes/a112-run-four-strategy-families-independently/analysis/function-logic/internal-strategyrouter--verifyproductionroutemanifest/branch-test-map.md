# Branch Test Map: `verifyProductionRouteManifest`

- Source: `internal/strategyrouter/production.go`; file SHA-256 `1175f67d72d78cc9f3ef65d505d97112382de26ea1eae89165314529dafb26d9`. AST branch positions are authoritative.
- Rows carry measured counts from Go coverage profiles, count mode.
- untagged package suite: `go test -count=1 -covermode=count ./internal/strategyrouter/`
- tagged package suite: `go test -count=1 -tags tossos_testseams -covermode=count ./internal/strategyrouter/`

Mutation receipts for this function (production source mutated, run, restored from a pristine copy taken before the mutation):

| # | mutation | result | killed by |
|---|---|---|---|
| M9 | delete the two `productionRouteIdentity` calls for the calibration seal | KILLED | `TestProductionRouteManifestRefusesAMissingCalibrationSeal` |
| M11 | roll `productionRouteSchema` back to `strategy-lane-authority:v1` | KILLED | `TestProductionRouteManifestRefusesTheLegacySchemaVersion` |

| Branch | Anchor | Measured disposition |
|---|---|---|
| B1 | if at 480:2 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |
| B2 | if at 495:2 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |
| B3 | if at 499:2 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |

A row states what was measured, not what is intended. An arm recorded as not entered is a coverage gap, not a pass.

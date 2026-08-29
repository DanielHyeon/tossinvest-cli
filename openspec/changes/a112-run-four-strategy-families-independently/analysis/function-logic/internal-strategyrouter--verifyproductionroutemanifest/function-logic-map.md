# Function Logic Map: `verifyProductionRouteManifest`

- Source: `internal/strategyrouter/production.go` (475-506)
- Function: `verifyProductionRouteManifest` in package `strategyrouter`
- Signature: `verifyProductionRouteManifest(params=2, results=2)`
- File SHA-256: `1175f67d72d78cc9f3ef65d505d97112382de26ea1eae89165314529dafb26d9`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 3.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Verifies one signed market manifest body against the engine-supplied config and the Ed25519 trust pin. The calibration seal is checked here: a body without an approved `arbitration_score_version` or `calibration_digest` is refused before the signature is even consulted, so an otherwise well-signed manifest that names no approved calibration is not an activation authority.

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- untagged package suite: `go test -count=1 -covermode=count ./internal/strategyrouter/`
- tagged package suite: `go test -count=1 -tags tossos_testseams -covermode=count ./internal/strategyrouter/`
- Measured entry: the function body was executed 23x (untagged package suite); executed 23x (tagged package suite).

Exact AST return positions: 492:3, 496:3, 500:3, 505:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 480:2 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |
| B2 | if | 495:2 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |
| B3 | if | 499:2 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `productionRouteTime` | 477:37 |
| `productionRouteTime` | 478:26 |
| `productionRouteTime` | 479:20 |
| `productionRouteIdentity` | 487:4 |
| `productionRouteIdentity` | 487:62 |
| `productionRouteIdentity` | 489:4 |
| `config.ActivationExpiresAt.IsZero` | 490:5 |
| `activationExpires.Equal` | 490:45 |
| `observed.After` | 490:101 |
| `config.ObservedAt.Before` | 490:139 |
| `config.ObservedAt.Before` | 491:4 |
| `observed.After` | 491:51 |
| `validProductionRouteScopes` | 494:20 |
| `json.Marshal` | 498:20 |
| `DecodeString` | 502:20 |
| `base64.StdEncoding.Strict` | 502:20 |
| `base64.StdEncoding.EncodeToString` | 503:28 |
| `len` | 503:98 |
| `ed25519.Verify` | 504:3 |

## State mutations and fallbacks

- AST assignments: 8. Defers: 0. Goroutine statements: 0.

## Safety conclusion

Every refusal here is fail-closed for the whole market snapshot: the function returns a zero scope and false, and the only caller turns that into `ErrProductionRouteUnavailable`. The refusal carries no distinguishing reason — that diagnosability gap is recorded as a residual in review.md decision 51, not closed here.

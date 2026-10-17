# Function Logic Map: `verifyProductionRouteManifest`

- Source: `internal/strategyrouter/production.go` (491-522)
- Function: `verifyProductionRouteManifest` in package `strategyrouter`
- Signature: `verifyProductionRouteManifest(params=2, results=2)`
- File SHA-256: `617163a508030dea2768655db66006d52097f17be9c04bb7c9fc300d3adf8a21`
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

Exact AST return positions: 508:3, 512:3, 516:3, 521:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 496:2 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |
| B2 | if | 511:2 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |
| B3 | if | 515:2 | arm not entered (untagged package suite); arm not entered (tagged package suite); no per-test profile in the attribution set entered it |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `productionRouteTime` | 493:37 |
| `productionRouteTime` | 494:26 |
| `productionRouteTime` | 495:20 |
| `productionRouteIdentity` | 503:4 |
| `productionRouteIdentity` | 503:62 |
| `productionRouteIdentity` | 505:4 |
| `config.ActivationExpiresAt.IsZero` | 506:5 |
| `activationExpires.Equal` | 506:45 |
| `observed.After` | 506:101 |
| `config.ObservedAt.Before` | 506:139 |
| `config.ObservedAt.Before` | 507:4 |
| `observed.After` | 507:51 |
| `validProductionRouteScopes` | 510:20 |
| `json.Marshal` | 514:20 |
| `DecodeString` | 518:20 |
| `base64.StdEncoding.Strict` | 518:20 |
| `base64.StdEncoding.EncodeToString` | 519:28 |
| `len` | 519:98 |
| `ed25519.Verify` | 520:3 |

## State mutations and fallbacks

- AST assignments: 8. Defers: 0. Goroutine statements: 0.

## Safety conclusion

Every refusal here is fail-closed for the whole market snapshot: the function returns a zero scope and false, and the only caller turns that into `ErrProductionRouteUnavailable`. The refusal carries no distinguishing reason — that diagnosability gap is recorded as a residual in review.md decision 51, not closed here.

a112 6.1: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)

a112 게이트 준비(2026-10-04): a127 82080177(전략 권한 적재기가 현재 원장을 읽음)이 같은 파일을 편집해 이 번들이 낡았다 — 분기 · return · 호출 구조 동일

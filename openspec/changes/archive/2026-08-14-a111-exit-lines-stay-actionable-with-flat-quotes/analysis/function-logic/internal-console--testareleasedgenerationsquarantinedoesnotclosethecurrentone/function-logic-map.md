# Function Logic Map: `TestAReleasedGenerationsQuarantineDoesNotCloseTheCurrentOne`

- Source: `internal/console/a077_live_protection_test.go`
- Post-edit AST evidence: `ast.json` (1 branches; source SHA-256 recorded by extractor)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| A111 state | persisted evidence, request/cycle state, and injected clock/marker | current source + approved A111 delta | invalid, stale, or incomplete evidence is fail-closed |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | AST branch at `internal/console/a077_live_protection_test.go:183`; legacy lifecycle coverage retains its isolation assertion under the shared freshness renderer | branch-local only; no authority added | function-defined fail-closed result | `TestA111ConsoleHidesStoppedFutureSeedAndCorruptEvidence` |
| Return | all admitted paths | preserves the A111 safety boundary | propagates typed error or read-only result | `TestA111ConsoleHidesStoppedFutureSeedAndCorruptEvidence` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| A111 direct collaborators | legacy lifecycle coverage retains its isolation assertion under the shared freshness renderer | typed conflict/stale/invalid results do not authorize an order or fresh line | current AST + A111 RED |

## State mutations and fallbacks

- legacy lifecycle coverage retains its isolation assertion under the shared freshness renderer
- The function remains inside its existing journal, engine, or projection authority; it does not create a new LIVE-order path.
- Every AST branch is paired with the named A111 RED in `branch-test-map.md`.

## Safety conclusion

- Safe edit boundary: the A111 heartbeat/evidence freshness seam only; retain full judgement and existing order ordering where applicable.
- High-risk impact: operator-facing fail-closed projection.
